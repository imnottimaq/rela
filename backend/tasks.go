package main

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/goccy/go-json"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"slices"
	"time"
)

// @Summary 		Get all tasks
// @Description 	Returns all tasks owned by the current user, or all tasks for a given workspace.
// @Router 			/tasks [get]
// @Router 			/workspaces/{workspaceId}/tasks/{boardId} [get]
// @Tags 			Tasks
// @Security 		BearerAuth
// @Produce 		json
// @Param 			workspaceId path string false "Workspace ID (optional)"
// @Param 			boardId path string true "Board ID"
// @Success 		200 {object} AllTasksResponse "A list of tasks"
// @Failure 		500 {object} ErrorSwagger "Internal Server Error"
func getAllTasks(c *gin.Context) {
	id, _ := c.Get("id")
	bId := c.Param("boardId")
	boardId, _ := bson.ObjectIDFromHex(bId)
	tasks := make([]Task, 0)
	if wId := c.Param("workspaceId"); wId == "" {
		cursor, _ := tasksDb.Find(context.TODO(), bson.D{
			{"created_by", id.(bson.ObjectID)},
			{"board", boardId},
		})
		_ = cursor.All(context.TODO(), &tasks)
		c.IndentedJSON(200, gin.H{"tasks": tasks})
		if err := cursor.Close(context.TODO()); err != nil {
			c.AbortWithStatusJSON(500, gin.H{"error": "Internal Server Error"})
			return
		}
		return
	} else {
		workspaceId, _ := bson.ObjectIDFromHex(wId)
		cursor, _ := tasksDb.Find(context.TODO(), bson.D{
			{"created_by", workspaceId},
			{"board", boardId},
		})
		_ = cursor.All(context.TODO(), &tasks)
		c.IndentedJSON(200, gin.H{"tasks": tasks})
		if err := cursor.Close(context.TODO()); err != nil {
			c.AbortWithStatusJSON(500, gin.H{"error": "Internal Server Error"})
			return
		}
	}
}

// @Summary 		Create a new task
// @Description 	Creates a new task for the current user or a workspace.
// @Router 			/tasks [post]
// @Router 			/workspaces/{workspaceId}/tasks [post]
// @Tags 			Tasks
// @Security 		BearerAuth
// @Accept 			json
// @Produce 		json
// @Param 			workspaceId path string false "Workspace ID (optional)"
// @Param 			data body CreateTask true "Task creation data"
// @Success 		200 {object} Task "The created task"
// @Failure 		400 {object} ErrorSwagger "Bad request - check your input"
// @Failure 		500 {object} ErrorSwagger "Internal server error"
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
	} else if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": "Internal Server Error"})
		return
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
		c.AbortWithStatusJSON(500, gin.H{"error": "Internal Server Error"})
		return
	}
	newTask.Id = task.InsertedID.(bson.ObjectID)
	c.AbortWithStatusJSON(200, newTask)
	return
}

func authorizeTaskAccess(c *gin.Context, task Task) bool {
	userId, _ := c.Get("id")
	wId := c.Param("workspaceId")

	if wId == "" {
		if task.CreatedBy != userId.(bson.ObjectID) {
			c.AbortWithStatusJSON(403, gin.H{"error": "You do not own this task"})
			return false
		}
	} else {
		workspaceId, err := bson.ObjectIDFromHex(wId)
		if err != nil {
			c.AbortWithStatusJSON(400, gin.H{"error": "Invalid workspaceId"})
			return false
		}

		var workspace Workspace
		if err := workspacesDb.FindOne(context.TODO(), bson.D{{"_id", workspaceId}}).Decode(&workspace); err != nil {
			if errors.Is(err, mongo.ErrNoDocuments) {
				c.AbortWithStatusJSON(404, gin.H{"error": "Workspace not found"})
			} else {
				c.AbortWithStatusJSON(500, gin.H{"error": "Internal server error"})
			}
			return false
		}

		if slices.Index(workspace.Members, userId.(bson.ObjectID)) == -1 {
			c.AbortWithStatusJSON(403, gin.H{"error": "You are not a member of this workspace"})
			return false
		}

		if task.CreatedBy != workspaceId {
			c.AbortWithStatusJSON(400, gin.H{"error": "This task does not belong to this workspace"})
			return false
		}
	}
	return true
}

// @Summary 		Edit an existing task
// @Description 	Edits the details of a specific task.
// @Router 			/tasks/{taskId} [patch]
// @Router 			/workspaces/{workspaceId}/tasks/{taskId} [patch]
// @Tags 			Tasks
// @Security 		BearerAuth
// @Accept 			json
// @Param 			taskId path string true "Task ID"
// @Param 			workspaceId path string false "Workspace ID (optional)"
// @Param 			data body EditTask true "Fields to edit in the task"
// @Success 		200 "Task updated successfully"
// @Failure 		400 {object} ErrorSwagger "Bad request - invalid input or deadline in the past"
// @Failure 		403 {object} ErrorSwagger "Forbidden - you do not have access to this task"
// @Failure 		404 {object} ErrorSwagger "Not Found - task or workspace not found"
// @Failure 		500 {object} ErrorSwagger "Internal server error"
func editExistingTask(c *gin.Context) {
	taskInput, exists := c.Get("taskObj")
	if !exists {
		c.AbortWithStatusJSON(500, gin.H{"error": "Task object not found in context"})
		return
	}
	task := taskInput.(Task)

	if !authorizeTaskAccess(c, task) {
		return
	}

	var valuesToEdit EditTask
	if err := json.NewDecoder(c.Request.Body).Decode(&valuesToEdit); err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": "Something went wrong when parsing request"})
		return
	}

	if valuesToEdit.Name != "" {
		task.Name = valuesToEdit.Name
	}
	if valuesToEdit.Description != "" {
		task.Description = valuesToEdit.Description
	}
	if valuesToEdit.Deadline != 0 {
		if valuesToEdit.Deadline <= time.Now().UTC().Unix() {
			c.AbortWithStatusJSON(400, gin.H{"error": "Deadline cant be past current time"})
			return
		}
		task.Deadline = valuesToEdit.Deadline
	}

	if _, err := tasksDb.ReplaceOne(context.TODO(), bson.D{{Key: "_id", Value: task.Id}}, &task); err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": "Internal Server Error"})
		return
	}
	c.AbortWithStatus(200)
}

// @Summary 		Delete an existing task
// @Description 	Deletes a specific task.
// @Router 			/tasks/{taskId} [delete]
// @Router 			/workspaces/{workspaceId}/tasks/{taskId} [delete]
// @Tags 			Tasks
// @Security 		BearerAuth
// @Param 			taskId path string true "Task ID"
// @Param 			workspaceId path string false "Workspace ID (optional)"
// @Success 		200 "Task deleted successfully"
// @Failure 		403 {object} ErrorSwagger "Forbidden - you do not have access to this task"
// @Failure 		404 {object} ErrorSwagger "Not Found - task or workspace not found"
// @Failure 		500 {object} ErrorSwagger "Internal server error"
func deleteExistingTask(c *gin.Context) {
	taskInput, exists := c.Get("taskObj")
	if !exists {
		c.AbortWithStatusJSON(500, gin.H{"error": "Task object not found in context"})
		return
	}
	task := taskInput.(Task)

	if !authorizeTaskAccess(c, task) {
		return
	}

	if _, err := tasksDb.DeleteOne(context.TODO(), bson.D{{Key: "_id", Value: task.Id}}); err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": "Failed to delete task"})
		return
	}
	c.AbortWithStatus(200)
}
