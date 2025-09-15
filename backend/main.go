package main

import (
	_ "Rela/docs"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"log"
	_ "net/http/pprof"
	"os"
	"regexp"
	"time"

	ratelimit "github.com/JGLTechnologies/gin-rate-limit"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

var _ = godotenv.Load("../.env")
var port = os.Getenv("PORT")
var pepper = os.Getenv("PEPPER")
var mongodbCredentials = os.Getenv("MONGO_CREDS")
var dbClient, _ = mongo.Connect(options.Client().ApplyURI(mongodbCredentials).SetServerAPIOptions(options.ServerAPI(options.ServerAPIVersion1)))

var tasksDb = dbClient.Database("rela").Collection("tasks")
var usersDb = dbClient.Database("rela").Collection("users")
var boardsDb = dbClient.Database("rela").Collection("boards")
var workspacesDb = dbClient.Database("rela").Collection("workspaces")

var emailRegex = regexp.MustCompile(`^[a-z0-9._%+-]+@[a-z0-9.-]+\.[a-z]{2,}$`)

func keyFunc(c *gin.Context) string {
	return c.ClientIP()
}
func rateLimitHandler(c *gin.Context, info ratelimit.Info) {
	c.AbortWithStatusJSON(429, gin.H{"error": "too many requests"})
}

func main() {
	r := gin.Default()
	r.MaxMultipartMemory = 8 << 20
	store := ratelimit.InMemoryStore(&ratelimit.InMemoryOptions{
		Rate:  time.Second,
		Limit: 5,
	})
	rateLimiter := ratelimit.RateLimiter(store, &ratelimit.Options{
		ErrorHandler: rateLimitHandler,
		KeyFunc:      keyFunc,
	})
	protected := r.Group("/api/v1")
	workspaces := protected.Group("/workspaces")
	tasks := protected.Group("/tasks")
	users := r.Group("/api/v1/users")
	boards := protected.Group("/boards")
	protected.Use(func(c *gin.Context) {
		header := c.GetHeader("X-Authorization")
		if header == "" {
			c.AbortWithStatusJSON(403, gin.H{"error": "no access token"})
			return
		}
		token, err := jwt.ParseWithClaims(header, &Token{}, func(token *jwt.Token) (any, error) {
			if token.Method != jwt.SigningMethodHS256 {
				return nil, fmt.Errorf("unknown signing method: %s", token.Method)
			}
			return []byte(pepper), nil
		})
		if err != nil {
			c.AbortWithStatusJSON(500, "Internal Server Error")
			return
		}
		claims := token.Claims.(*Token)
		if claims.ExpiresAt < time.Now().UTC().Unix() {
			c.AbortWithStatusJSON(403, "Authorization Required")
			return
		} else if claims.Type == "refresh" {
			c.AbortWithStatusJSON(400, "Invalid Token")
		} else {
			c.Set("id", claims.Id)
			c.Next()
		}
	})
	{
		//Tasks
		r.Static("/app", "../frontend/dist/")
		r.Static("/assets", "../frontend/dist/assets")
		r.Static("/img", "./img")
		tasks.GET("/", rateLimiter, getAllTasks)
		tasks.POST("/", rateLimiter, createNewTask)
		tasks.PATCH("/:taskId", rateLimiter, editExistingTask)
		tasks.DELETE("/:taskId", rateLimiter, deleteExistingTask)

		//Boards
		boards.POST("/", rateLimiter, addBoard)
		boards.DELETE("/:boardId", rateLimiter, deleteBoard)
		boards.PATCH("/:boardId", rateLimiter, editBoard)

		//Users
		users.POST("/create", rateLimiter, createUser)
		users.POST("/login", rateLimiter, loginUser)
		users.GET("/refresh", rateLimiter, refreshAccessToken)

		protected.DELETE("/users/delete", rateLimiter, deleteUser)
		protected.POST("/users/upload_avatar", rateLimiter, uploadAvatar)

		//Workspaces
		workspaces.POST("/", rateLimiter, createWorkspace)
		workspaces.POST("/add/:joinToken", rateLimiter, addMember)
		//Docs
		r.GET("/docs/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
	}
	if pepper == "" {
		log.Print("WARNING Server-side secret is not present, this is big security flaw")
	} else if mongodbCredentials == "" {
		log.Fatal("FATAL MongoDB credentials is not present")
	} else if port == "" {
		log.Print("WARNING Port is not present, falling back to default")
		if err := r.Run(":8080"); err != nil {
			log.Fatal("Failed to start server")
		}
	}
	if err := r.Run(port); err != nil {
		log.Fatal("Failed to start server")
	}
}
