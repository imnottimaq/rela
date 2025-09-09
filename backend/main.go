package main

import (
	_ "Rela/docs"
	"context"
	"log"
	_ "net/http/pprof"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	ratelimit "github.com/JGLTechnologies/gin-rate-limit"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
)

var _ = godotenv.Load(".env")
var mongodbCredentials = os.Getenv("MONGO_CREDS")
var port = os.Getenv("PORT")
var dbClient, _ = mongo.Connect(options.Client().ApplyURI(mongodbCredentials).SetServerAPIOptions(options.ServerAPI(options.ServerAPIVersion1)))

func keyFunc(c *gin.Context) string {
	return c.ClientIP()
}
func rateLimitHandler(c *gin.Context, info ratelimit.Info) {
	c.AbortWithStatusJSON(429, gin.H{"error": "too many requests"})
}

func main() {
	if err := dbClient.Ping(context.TODO(), readpref.Primary()); err != nil {
		log.Fatal("Failed to ping MongoDB database")
	}
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
	protected.Use(authMiddleware())
	{
		//Tasks
		protected.GET("/tasks", rateLimiter, getAllTasks)
		protected.POST("/tasks", rateLimiter, createNewTask)
		protected.PATCH("/tasks/:taskId", rateLimiter, editExistingTask)
		protected.DELETE("/tasks/:taskId", rateLimiter, deleteExistingTask)

		//Boards
		protected.POST("/boards", rateLimiter, addBoard)
		protected.DELETE("/boards/:boardId", rateLimiter, deleteBoard)
		protected.PATCH("/boards/:boardId", rateLimiter, editBoard)

		//Users
		r.POST("/api/v1/users/create", rateLimiter, createUser)
		r.POST("/api/v1/users/login", rateLimiter, loginUser)
		r.GET("/api/v1/users/refresh", rateLimiter, refreshAccessToken)
		protected.DELETE("/users/delete", rateLimiter, deleteUser)
		protected.POST("/users/upload_avatar", rateLimiter, uploadAvatar)

		r.GET("/docs", ginSwagger.WrapHandler(swaggerfiles.Handler))
	}
	if err := r.Run(port); err != nil {
		log.Fatal("Failed to start server")
	}
}
