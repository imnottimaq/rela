package main

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/goccy/go-json"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"log"
	"time"
)

// @Title			Rela API Docs
// @Description	Simple WIP task tracker that can be self-hosted
// @Version		1.0
// @BasePath		/api/v1

// @Summary Get all tasks
// @Description Return all tasks that current user owns
// @Router /api/v1/tasks [get]
// @Success 200 {array} Task
// @Produce json
// @Tags Tasks
// @Param X-Authorization header string true "Bearer Token"
func getAllTasks(c *gin.Context) {
	id, _ := c.Get("id")
	var tasks []Task
	if workspaceId := c.Param("workspaceId"); workspaceId == "" {
		cursor, _ := tasksDb.Find(context.TODO(), bson.D{{"created_by", id}})

		_ = cursor.All(context.TODO(), &tasks)
		c.IndentedJSON(200, tasks)
		if err := cursor.Close(context.TODO()); err != nil {
			log.Print("Failed to close cursor")
		}
		return
	} else {
		cursor, _ := tasksDb.Find(context.TODO(), bson.D{{"created_by", workspaceId}})
		_ = cursor.All(context.TODO(), &tasks)
		c.IndentedJSON(200, tasks)
		if err := cursor.Close(context.TODO()); err != nil {
			log.Print("Failed to close cursor")
		}
	}
}

// @Summary Create new task
// @Router /api/v1/tasks [post]
// @Accept json
// @Success 200 {array} Task
// @Produce json
// @Tags Tasks
// @Param data body CreateTask true "Create task request"
// @Param X-Authorization header string true "Bearer Token"
func createNewTask(c *gin.Context) {
	id, _ := c.Get("id")
	var input CreateTask
	var board Board
	err := json.NewDecoder(c.Request.Body).Decode(&input)
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": "Something went wrong when parsing request"})
		return
	} else if input.Name == "" {
		c.AbortWithStatusJSON(400, gin.H{"error": "Field 'name' is not specified"})
		return
	} else if input.Board.IsZero() {
		c.AbortWithStatusJSON(400, gin.H{"error": "Field 'board' is not specified"})
		return
	}
	err = boardsDb.FindOne(context.TODO(), bson.D{{"_id", input.Board}}).Decode(&board)
	if errors.Is(err, mongo.ErrNoDocuments) {
		c.AbortWithStatusJSON(400, gin.H{"error": "Board does not exist"})
		return
	} else {
		c.AbortWithStatusJSON(500, gin.H{"error": "Internal server error"})
	}
	newTask := Task{
		Name:        input.Name,
		Description: input.Description,
		CreatedAt:   time.Now().UTC().Unix(),
		CreatedBy:   id.(bson.ObjectID),
		Board:       input.Board,
	}
	if workspaceId := c.Param("workspaceId"); workspaceId != "" {
		transformedId, _ := bson.ObjectIDFromHex(workspaceId)
		newTask.CreatedBy = transformedId
	}
	task, err := tasksDb.InsertOne(context.TODO(), newTask)
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": "Failed to push task to database"})
		return
	}
	newTask.Id = task.InsertedID.(bson.ObjectID)
	c.AbortWithStatusJSON(200, newTask)
	return
}

// @Summary Edit existing task
// @Router /api/v1/tasks/{taskId} [patch]
// @Accept json
// @Success 200
// @Tags Tasks
// @Param taskId path bson.ObjectID true "Task ID"
// @Param data body EditTask true "Edit task request"
// @Param X-Authorization header string true "Bearer Token"
func editExistingTask(c *gin.Context) {
	input, _ := c.Get("output")
	previousVersion := input.(Task)
	var valuesToEdit EditTask
	taskId := c.Param("taskId")
	err := json.NewDecoder(c.Request.Body).Decode(&valuesToEdit)
	if err != nil {
		c.AbortWithStatusJSON(400, gin.H{"error": "Something went wrong when parsing request"})
		return
	}
	if valuesToEdit.Name != nil {
		previousVersion.Name = *valuesToEdit.Name
		if previousVersion.Name == "" {
			c.AbortWithStatusJSON(400, gin.H{"error": "Name cannot be empty"})
			return
		}
	}
	if valuesToEdit.Description != nil {
		previousVersion.Description = valuesToEdit.Description
	}
	_, err = tasksDb.ReplaceOne(context.TODO(), bson.D{{Key: "_id", Value: taskId}}, &previousVersion)
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": "Failed to change task"})
		return
	}
	c.AbortWithStatus(200)
	return
}

// @Summary Edit existing task
// @Router /api/v1/tasks/{taskId} [delete]
// @Success 200
// @Produce json
// @Tags Tasks
// @Param taskId path bson.ObjectID true "Task ID"
// @Param X-Authorization header string true "Bearer Token"
func deleteExistingTask(c *gin.Context) {
	input, _ := c.Get("output")
	result := input.(Task)
	_, err := tasksDb.DeleteOne(context.TODO(), bson.D{{Key: "_id", Value: result.Id}})
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": "Failed to delete task"})
		return
	}
}
