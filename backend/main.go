package main

import (
	"context"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
)

var _ = godotenv.Load()
var mongodbCredentials = os.Getenv("MONGO_CREDS")
var port = os.Getenv("PORT")
var dbClient, _ = mongo.Connect(options.Client().ApplyURI(mongodbCredentials).SetServerAPIOptions(options.ServerAPI(options.ServerAPIVersion1)))

func main() {
	if err := dbClient.Ping(context.TODO(), readpref.Primary()); err != nil {
		log.Fatal("Failed to ping MongoDB database")
	}
	r := gin.Default()
	protected := r.Group("/api/v1")
	protected.Use(authMiddleware())
	{
		protected.GET("/tasks", getAllTasks)
		protected.POST("/tasks", createNewTask)
		protected.PATCH("/tasks/:taskId", editExistingTask)
		protected.DELETE("/tasks/:taskId", deleteExistingTask)
		protected.POST("/boards", addBoard)
		protected.DELETE("/users/delete", deleteUser)

		r.POST("/api/v1/users/create", createUser)
		r.POST("/api/v1/users/login", loginUser)
		r.POST("/api/v1/users/refresh", refreshAccessToken)
	}
	if err := r.Run(port); err != nil {
		log.Fatal("Failed to start server")
	}
}
