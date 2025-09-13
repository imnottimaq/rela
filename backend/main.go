package main

import (
	_ "Rela/docs"
	"log"
	_ "net/http/pprof"
	"os"
	"time"

	ratelimit "github.com/JGLTechnologies/gin-rate-limit"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

var _ = godotenv.Load(".env")
var port = os.Getenv("PORT")

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
	protected.Use(authMiddleware())
	{
		//Tasks
		r.Static("/app", "../frontend/dist/")
		r.Static("/assets", "../frontend/dist/assets")
		r.Static("/img", "./img")
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
