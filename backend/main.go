package main

import (
	"context"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"log"
	_ "net/http/pprof"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
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
		//Tasks
		protected.GET("/tasks", getAllTasks)
		protected.POST("/tasks", createNewTask)
		protected.PATCH("/tasks/:taskId", editExistingTask)
		protected.DELETE("/tasks/:taskId", deleteExistingTask)

		//Boards
		protected.POST("/boards", addBoard)
		protected.DELETE("/boards/:boardId", deleteBoard)
		protected.PATCH("/boards/:boardId", editBoard)

		//Users
		r.POST("/api/v1/users/create", createUser)
		r.POST("/api/v1/users/login", loginUser)
		r.POST("/api/v1/users/refresh", refreshAccessToken)
		protected.DELETE("/users/delete", deleteUser)
	}
	if err := r.Run(port); err != nil {
		log.Fatal("Failed to start server")
	}
}
