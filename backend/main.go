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
	v1 := r.Group("/api/v1")
	protected := v1.Group("/")
	protected.Use(authMiddleware())

	// Middleware
	r.Use(rateLimiter)

	{
		v1.Static("/img", "./img")

		// Users
		usersGroup := v1.Group("/users")
		usersGroup.Use(userMiddleware())
		{
			usersGroup.POST("/create", createUser)
			usersGroup.POST("/login", loginUser)
			usersGroup.GET("/refresh", refreshAccessToken)
			usersGroup.POST("/logout", logoutUser)
		}

		// Protected User Routes
		protectedUsersGroup := protected.Group("/users")
		{
			protectedUsersGroup.GET("/workspaces", getAllWorkspaces)
			protectedUsersGroup.DELETE("/delete", deleteUser)
			protectedUsersGroup.POST("/upload_avatar", uploadAvatar)
			protectedUsersGroup.GET("/get_info", getUserDetails)
		}

		// Workspaces
		workspacesGroup := protected.Group("/workspaces")
		workspaceByIdGroup := workspacesGroup.Group("/:workspaceId")
		{
			workspacesGroup.POST("/create", createWorkspace)
			workspacesGroup.POST("/invite/accept/:joinToken", addMember)
			r.GET("/workspaces/invite/:joinToken", getWorkspaceByInviteToken)

			workspaceByIdGroup.GET("/new_invite", createNewInvite)
			workspaceByIdGroup.DELETE("/kick", kickMember)
			workspaceByIdGroup.PATCH("/promote/:userId", promoteMember)
			workspaceByIdGroup.GET("/", getWorkspace)
			workspaceByIdGroup.GET("/info", getWorkspaceInfo)
			workspaceByIdGroup.PATCH("/", editWorkspace)
			workspaceByIdGroup.DELETE("/", deleteWorkspace)
			workspaceByIdGroup.POST("/upload_avatar", uploadAvatar)

			// Workspace Tasks
			workspaceByIdGroup.GET("/tasks/:boardId", getAllTasks)
			workspaceByIdGroup.POST("/tasks", createNewTask)
			workspaceByIdGroup.PATCH("/tasks/:taskId", taskMiddleware(), editExistingTask)
			workspaceByIdGroup.DELETE("/delete/:taskId", taskMiddleware(), deleteExistingTask)
			workspaceByIdGroup.POST("/assign", assignTask)

			// Workspace Boards
			workspaceByIdGroup.GET("/boards", getAllBoards)
			workspaceByIdGroup.POST("/boards", addBoard)
			workspaceByIdGroup.DELETE("/boards/:boardId", deleteBoard)
			workspaceByIdGroup.PATCH("/boards/:boardId", editBoard)
		}

		// Public invite route
		v1.GET("/workspaces/invite/:joinToken", getWorkspaceByInviteToken)

		// Docs
		v1.GET("/docs/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
	}

	if pepper == "" {
		print("WARNING Server-side secret is not present, this is a big security flaw")
	} else if mongodbCredentials == "" {
		panic("FATAL MongoDB credentials are not present")
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
