package main

import (
	_ "Rela/docs"
	"os"
	"regexp"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	ratelimit "github.com/khaaleoo/gin-rate-limiter/core"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

var _ = godotenv.Load(".env_local")
var port = os.Getenv("PORT")
var pepper = os.Getenv("PEPPER")
var mongodbCredentials = os.Getenv("MONGO_CREDS")
var frontendOriginEnv = os.Getenv("FRONTEND_ORIGINS")
var dbClient, _ = mongo.Connect(options.Client().ApplyURI(mongodbCredentials).SetServerAPIOptions(options.ServerAPI(options.ServerAPIVersion1)).SetMaxPoolSize(100).SetMinPoolSize(10).SetMaxConnIdleTime(30 * time.Second))

var tasksDb = dbClient.Database("rela").Collection("tasks")
var usersDb = dbClient.Database("rela").Collection("users")
var boardsDb = dbClient.Database("rela").Collection("boards")
var workspacesDb = dbClient.Database("rela").Collection("workspaces")

var emailRegex = regexp.MustCompile(`^[a-z0-9._%+-]+@[a-z0-9.-]+\.[a-z]{2,}$`)

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
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	r := gin.Default()
	r.RedirectTrailingSlash = false // Explicitly disable automatic redirects
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = getAllowedOrigins()
	corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}
	corsConfig.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With", "X-Authorization"}
	corsConfig.AllowCredentials = true
	r.Use(cors.New(corsConfig))
	r.MaxMultipartMemory = 8 << 20
	// Rate Limiter
	rateLimiter := ratelimit.RequireRateLimiter(ratelimit.RateLimiter{
		RateLimiterType: ratelimit.IPRateLimiter,
		Key:             "rela",
		Option:          ratelimit.RateLimiterOption{Limit: 5, Burst: 100, Len: 1 * time.Minute},
	})
	// Groups
	protected := r.Group("/api/v1")
	workspaces := protected.Group("/workspaces/:workspaceId")
	tasks := protected.Group("/tasks")
	users := r.Group("/api/v1/users")
	boards := protected.Group("/boards")

	//Middleware
	r.Use(rateLimiter)
	protected.Use(authMiddleware())
	boards.Use(authMiddleware())
	tasks.Use(authMiddleware())
	users.Use(userMiddleware())
	workspaces.Use(authMiddleware())
	{
		r.Static("/img", "./img")
		//Tasks
		tasks.GET("/:boardId", getAllTasks)
		tasks.POST("/", createNewTask)
		tasks.PATCH("/:taskId", taskMiddleware(), editExistingTask)
		tasks.DELETE("/:taskId", taskMiddleware(), deleteExistingTask)

		//Boards
		boards.GET("/", getAllBoards)
		boards.POST("/", addBoard)
		boards.DELETE("/:boardId", deleteBoard)
		boards.PATCH("/:boardId", editBoard)

		//Users
		users.POST("/create", createUser)
		users.POST("/login", loginUser)
		users.GET("/refresh", refreshAccessToken)
		users.POST("/logout", logoutUser)

		protected.GET("/users/workspaces", getAllWorkspaces)
		protected.DELETE("/users/delete", deleteUser)
		protected.POST("/users/upload_avatar", uploadAvatar)
		protected.GET("/users/get_info", getUserDetails)

		//Workspace management
		protected.POST("/workspaces/create", createWorkspace)
		workspaces.POST("/add/:joinToken", addMember)
		workspaces.GET("/new_invite", createNewInvite)
		workspaces.DELETE("/kick", kickMember)
		workspaces.PATCH("/promote/:userId", promoteMember)
		workspaces.GET("/members", getAllMembers)
		workspaces.GET("/", getWorkspace)
		workspaces.GET("/info", getWorkspaceInfo)
		workspaces.PATCH("/", editWorkspace)
		workspaces.DELETE("/", deleteWorkspace)
		workspaces.POST("/upload_avatar", uploadAvatar)

		//Workspace tasks
		workspaces.GET("/tasks/:boardId", getAllTasks)
		workspaces.POST("/tasks", createNewTask)
		workspaces.PATCH("/tasks/:taskId", taskMiddleware(), editExistingTask)
		workspaces.DELETE("/delete/:taskId", taskMiddleware(), deleteExistingTask)
		workspaces.POST("/assign", assignTask)

		//Workspace boards
		workspaces.GET("/boards", getAllBoards)
		workspaces.GET("/boards/:boardId", getBoard)
		workspaces.POST("/boards", addBoard)
		workspaces.DELETE("/boards/:boardId", deleteBoard)
		workspaces.PATCH("/boards/:boardId", editBoard)

		//Single board by id
		boards.GET("/:boardId", getBoard)

		//Docs
		r.GET("/api/v1/docs/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
	}
	if pepper == "" {
		print("WARNING Server-side secret is not present, this is big security flaw")
	} else if mongodbCredentials == "" {
		panic("FATAL MongoDB credentials is not present")
	} else if port == "" {
		print("WARNING Port is not present, falling back to default")
		if err := r.Run(":8080"); err != nil {
			panic("Failed to start server")
		}
	}
	if err := r.Run(port); err != nil {
		panic("Failed to start server")
	}
}
