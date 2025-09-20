package main

import (
	_ "Rela/docs"
	"log"
	_ "net/http/pprof"
	"os"
	"regexp"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	ratelimit "github.com/JGLTechnologies/gin-rate-limit"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

var _ = godotenv.Load(".env_local")
var port = os.Getenv("PORT")
var pepper = os.Getenv("PEPPER")
var mongodbCredentials = os.Getenv("MONGO_CREDS")
var frontendOriginEnv = os.Getenv("FRONTEND_ORIGINS")
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

func getAllowedOrigins() []string {
	if frontendOriginEnv == "" {
		return []string{"http://localhost:5173", "http://localhost:8000", "http://localhost:5174"}
	}
	origins := strings.Split(frontendOriginEnv, ",")
	for i := range origins {
		origins[i] = strings.TrimSpace(origins[i])
	}
	return origins
}

// @Title			Rela API Docs
// @Description	Simple WIP task tracker that can be self-hosted
// @Version		1.0
// @BasePath		/api/v1
func main() {
	r := gin.Default()
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = getAllowedOrigins()
	corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}
	corsConfig.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With", "X-Authorization"}
	corsConfig.AllowCredentials = true
	r.Use(cors.New(corsConfig))
	r.MaxMultipartMemory = 8 << 20
	store := ratelimit.InMemoryStore(&ratelimit.InMemoryOptions{
		Rate:  time.Second,
		Limit: 5,
	})
	rateLimiter := ratelimit.RateLimiter(store, &ratelimit.Options{
		ErrorHandler: rateLimitHandler,
		KeyFunc:      keyFunc,
	})

	// Groups
	protected := r.Group("/api/v1")
	workspaces := protected.Group("/workspaces/:workspaceId")
	tasks := protected.Group("/tasks")
	users := r.Group("/api/v1/users")
	boards := protected.Group("/boards")

	//Middleware
	protected.Use(authMiddleware())
	boards.Use(authMiddleware())
	tasks.Use(authMiddleware())
	tasks.Use(taskMiddleware())
	users.Use(userMiddleware())
	workspaces.Use(authMiddleware())
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
		boards.GET("/", rateLimiter, getAllBoards)
		boards.POST("/", rateLimiter, addBoard)
		boards.DELETE("/:boardId", rateLimiter, deleteBoard)
		boards.PATCH("/:boardId", rateLimiter, editBoard)

		//Users
		users.POST("/create", rateLimiter, createUser)
		users.POST("/login", rateLimiter, loginUser)
		users.GET("/refresh", rateLimiter, refreshAccessToken)

		protected.GET("/users/workspaces", rateLimiter, getAllWorkspaces)
		protected.DELETE("/users/delete", rateLimiter, deleteUser)
		protected.POST("/users/upload_avatar", rateLimiter, uploadAvatar)
		protected.GET("/users/get_info", rateLimiter, getUserDetails)

		//Workspace management
		protected.POST("/workspaces/create", rateLimiter, createWorkspace)
		workspaces.POST("/add/:joinToken", rateLimiter, addMember)
		workspaces.GET("/new_invite", rateLimiter, createNewInvite)
		workspaces.DELETE("/kick", rateLimiter, kickMember)
		workspaces.PATCH("/promote/:userId", rateLimiter, promoteMember)
		workspaces.GET("/members", rateLimiter, getAllMembers)
		workspaces.GET("/", rateLimiter, getWorkspace)
		workspaces.GET("/info", rateLimiter, getWorkspaceInfo)
		workspaces.PATCH("/", rateLimiter, editWorkspace)
		workspaces.DELETE("/", rateLimiter, deleteWorkspace)
		workspaces.POST("/upload_avatar", rateLimiter, uploadAvatar)

		//Workspace tasks
		workspaces.GET("/tasks/", rateLimiter, getAllTasks)
		workspaces.POST("/tasks/", rateLimiter, createNewTask)
		workspaces.PATCH("/tasks/:taskId", rateLimiter, editExistingTask)
		workspaces.DELETE("/delete/:taskId", rateLimiter, deleteExistingTask)
		workspaces.POST("/assign", rateLimiter, assignTask)

		//Workspace boards
		workspaces.GET("/boards", rateLimiter, getAllBoards)
		workspaces.GET("/boards/:boardId", rateLimiter, getBoard)
		workspaces.POST("/boards", rateLimiter, addBoard)
		workspaces.DELETE("/boards/:boardId", rateLimiter, deleteBoard)
		workspaces.PATCH("/boards/:boardId", rateLimiter, editBoard)

		//Single board by id
		boards.GET("/:boardId", rateLimiter, getBoard)

		//Docs
		r.GET("/api/v1/docs/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
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
