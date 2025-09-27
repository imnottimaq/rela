package main

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/goccy/go-json"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"slices"
)

// @Summary 		Create a new board
// @Description 	Creates a new board for a workspace.
// @Router 			/workspaces/{workspaceId}/boards [post]
// @Tags 			Boards
// @Security 		BearerAuth
// @Accept 			json
// @Produce 		json
// @Param 			workspaceId path string true "Workspace ID"
// @Param 			data body CreateBoard true "Board creation data"
// @Success 		200 {object} Board "The created board"
// @Failure 		400 {object} ErrorSwagger "Bad request - no name given"
// @Failure			403 {object} ErrorSwagger "Forbidden - you are not a member of this workspace"
// @Failure 		500 {object} ErrorSwagger "Internal server error"
func addBoard(c *gin.Context) {
	userId, _ := c.Get("id")
	wId := c.Param("workspaceId")
	workspaceId, _ := bson.ObjectIDFromHex(wId)

	// Authorize
	var workspace Workspace
	if err := workspacesDb.FindOne(context.TODO(), bson.M{"_id": workspaceId}).Decode(&workspace); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			c.AbortWithStatusJSON(404, gin.H{"error": "Workspace not found"})
		} else {
			c.AbortWithStatusJSON(500, gin.H{"error": "Internal server error"})
		}
		return
	}
	if slices.Index(workspace.Members, userId.(bson.ObjectID)) == -1 && workspace.OwnedBy != userId.(bson.ObjectID) {
		c.AbortWithStatusJSON(403, gin.H{"error": "You are not a member of this workspace"})
		return
	}

	var input Board
	err := json.NewDecoder(c.Request.Body).Decode(&input)
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": "Something went wrong when parsing request"})
		return
	}
	if input.Name == "" {
		c.AbortWithStatusJSON(400, gin.H{"error": "No name given"})
		return
	}

	input.OwnedBy = workspaceId
	result, err := boardsDb.InsertOne(context.TODO(), input)
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": "Failed to create board"})
		return
	}
	input.Id = result.InsertedID.(bson.ObjectID)
	c.JSON(200, input)
}

// @Summary 		Delete a board
// @Description 	Deletes a specific board from a workspace.
// @Router 			/workspaces/{workspaceId}/boards/{boardId} [delete]
// @Tags 			Boards
// @Security 		BearerAuth
// @Produce 		json
// @Param 			boardId path string true "Board ID"
// @Param 			workspaceId path string true "Workspace ID"
// @Success 		200 "Board deleted successfully"
// @Failure 		403 {object} ErrorSwagger "Forbidden - you do not have permission"
// @Failure 		404 {object} ErrorSwagger "Not Found - board or workspace not found"
// @Failure 		500 {object} ErrorSwagger "Internal server error"
func deleteBoard(c *gin.Context) {
	userId, _ := c.Get("id")
	boardIdStr := c.Param("boardId")
	boardId, err := bson.ObjectIDFromHex(boardIdStr)
	if err != nil {
		c.AbortWithStatusJSON(400, gin.H{"error": "invalid boardId"})
		return
	}
	wId := c.Param("workspaceId")
	workspaceId, _ := bson.ObjectIDFromHex(wId)

	var board Board
	if err := boardsDb.FindOne(context.TODO(), bson.M{"_id": boardId}).Decode(&board); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			c.AbortWithStatusJSON(404, gin.H{"error": "Board does not exist"})
		} else {
			c.AbortWithStatusJSON(500, gin.H{"error": "Failed to retrieve board"})
		}
		return
	}

	// Authorization
	var workspace Workspace
	if err := workspacesDb.FindOne(context.TODO(), bson.M{"_id": workspaceId}).Decode(&workspace); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			c.AbortWithStatusJSON(404, gin.H{"error": "Workspace not found"})
		} else {
			c.AbortWithStatusJSON(500, gin.H{"error": "Internal server error"})
		}
		return
	}
	// Check if board belongs to the workspace
	if board.OwnedBy != workspaceId {
		c.AbortWithStatusJSON(403, gin.H{"error": "This board does not belong to this workspace"})
		return
	}
	// Only workspace owner can delete boards
	if workspace.OwnedBy != userId.(bson.ObjectID) {
		c.AbortWithStatusJSON(403, gin.H{"error": "Only workspace owner can delete boards"})
		return
	}

	if _, err := boardsDb.DeleteOne(context.TODO(), bson.D{{"_id", boardId}}); err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": "Failed to remove board"})
		return
	}
	// Also delete all tasks on this board
	if _, err := tasksDb.DeleteMany(context.TODO(), bson.D{{"board", boardId}}); err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": "Failed to remove tasks on board"})
		return
	}
	c.AbortWithStatus(200)
}

// @Summary 		Edit a board
// @Description 	Edits the details of a specific board in a workspace.
// @Router 			/workspaces/{workspaceId}/boards/{boardId} [patch]
// @Tags 			Boards
// @Security 		BearerAuth
// @Accept 			json
// @Produce 		json
// @Param 			boardId path string true "Board ID"
// @Param 			workspaceId path string true "Workspace ID"
// @Param 			data body CreateBoard true "Fields to edit in the board"
// @Success 		200 {object} Board "The updated board"
// @Failure 		400 {object} ErrorSwagger "Bad request - invalid input"
// @Failure 		403 {object} ErrorSwagger "Forbidden - you do not have permission"
// @Failure 		404 {object} ErrorSwagger "Not Found - board or workspace not found"
// @Failure 		500 {object} ErrorSwagger "Internal server error"
func editBoard(c *gin.Context) {
	userId, _ := c.Get("id")
	boardIdStr := c.Param("boardId")
	boardId, err := bson.ObjectIDFromHex(boardIdStr)
	if err != nil {
		c.AbortWithStatusJSON(400, gin.H{"error": "invalid boardId"})
		return
	}
	wId := c.Param("workspaceId")
	workspaceId, _ := bson.ObjectIDFromHex(wId)

	var valuesToEdit CreateBoard
	if err := json.NewDecoder(c.Request.Body).Decode(&valuesToEdit); err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": "Something went wrong when parsing request"})
		return
	}

	var board Board
	if err := boardsDb.FindOne(context.TODO(), bson.M{"_id": boardId}).Decode(&board); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			c.AbortWithStatusJSON(404, gin.H{"error": "Board does not exist"})
		} else {
			c.AbortWithStatusJSON(500, gin.H{"error": "Failed to retrieve board"})
		}
		return
	}

	// Authorization
	var workspace Workspace
	if err := workspacesDb.FindOne(context.TODO(), bson.M{"_id": workspaceId}).Decode(&workspace); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			c.AbortWithStatusJSON(404, gin.H{"error": "Workspace not found"})
		} else {
			c.AbortWithStatusJSON(500, gin.H{"error": "Internal server error"})
		}
		return
	}
	if board.OwnedBy != workspaceId {
		c.AbortWithStatusJSON(403, gin.H{"error": "This board does not belong to this workspace"})
		return
	}
	if slices.Index(workspace.Members, userId.(bson.ObjectID)) == -1 && workspace.OwnedBy != userId.(bson.ObjectID) {
		c.AbortWithStatusJSON(403, gin.H{"error": "You are not a member of this workspace"})
		return
	}

	if valuesToEdit.Name != "" {
		board.Name = valuesToEdit.Name
	}

	if _, err := boardsDb.ReplaceOne(context.TODO(), bson.D{{"_id", boardId}}, board); err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": "Failed to update board"})
		return
	}
	c.JSON(200, board)
}

// @Summary 		Get all boards
// @Description 	Returns all boards for a given workspace.
// @Router 			/workspaces/{workspaceId}/boards [get]
// @Tags 			Boards
// @Security 		BearerAuth
// @Produce 		json
// @Param 			workspaceId path string true "Workspace ID"
// @Success 		200 {object} AllBoardsResponse "A list of boards"
// @Failure			403 {object} ErrorSwagger "Forbidden"
// @Failure 		500 {object} ErrorSwagger "Internal Server Error"
func getAllBoards(c *gin.Context) {
	userId, _ := c.Get("id")
	wId := c.Param("workspaceId")
	workspaceId, _ := bson.ObjectIDFromHex(wId)

	// Authorize
	var workspace Workspace
	if err := workspacesDb.FindOne(context.TODO(), bson.M{"_id": workspaceId}).Decode(&workspace); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			c.AbortWithStatusJSON(404, gin.H{"error": "Workspace not found"})
		} else {
			c.AbortWithStatusJSON(500, gin.H{"error": "Internal server error"})
		}
		return
	}
	if slices.Index(workspace.Members, userId.(bson.ObjectID)) == -1 && workspace.OwnedBy != userId.(bson.ObjectID) {
		c.AbortWithStatusJSON(403, gin.H{"error": "You are not a member of this workspace"})
		return
	}

	cursor, _ := boardsDb.Find(context.TODO(), bson.D{{"owned_by", workspaceId}})
	boards := make([]Board, 0)
	_ = cursor.All(context.TODO(), &boards)
	c.IndentedJSON(200, gin.H{"boards": boards})
	if err := cursor.Close(context.TODO()); err != nil {
		// Log the error but don't abort, as the response has already been sent.
		println("Failed to close cursor")
	}
}

// @Summary 		Get a board by ID
// @Description 	Retrieves a specific board by its ID from a workspace.
// @Router 			/workspaces/{workspaceId}/boards/{boardId} [get]
// @Tags 			Boards
// @Security 		BearerAuth
// @Produce 		json
// @Param 			boardId path string true "Board ID"
// @Param 			workspaceId path string true "Workspace ID"
// @Success 		200 {object} Board "The requested board"
// @Failure 		400 {object} ErrorSwagger "Bad request - invalid board ID"
// @Failure 		403 {object} ErrorSwagger "Forbidden - you do not have access to this board"
// @Failure 		404 {object} ErrorSwagger "Not Found - board not found"
// @Failure 		500 {object} ErrorSwagger "Internal server error"
func getBoard(c *gin.Context) {
	userId, _ := c.Get("id")
	boardIdStr := c.Param("boardId")
	if boardIdStr == "" {
		c.AbortWithStatusJSON(400, gin.H{"error": "Board id cannot be empty"})
		return
	}
	boardId, err := bson.ObjectIDFromHex(boardIdStr)
	if err != nil {
		c.AbortWithStatusJSON(400, gin.H{"error": "invalid boardId"})
		return
	}
	wId := c.Param("workspaceId")
	workspaceId, _ := bson.ObjectIDFromHex(wId)

	var board Board
	if err := boardsDb.FindOne(context.TODO(), bson.M{"_id": boardId}).Decode(&board); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			c.AbortWithStatusJSON(404, gin.H{"error": "Board does not exist"})
			return
		}
		c.AbortWithStatusJSON(500, gin.H{"error": "Failed to fetch board"})
		return
	}

	// Authorization
	if board.OwnedBy != workspaceId {
		c.AbortWithStatusJSON(403, gin.H{"error": "This board does not belong to this workspace"})
		return
	}

	var workspace Workspace
	if err := workspacesDb.FindOne(context.TODO(), bson.M{"_id": board.OwnedBy}).Decode(&workspace); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			// This case should ideally not be reached if board.OwnedBy is valid
			c.AbortWithStatusJSON(404, gin.H{"error": "Containing workspace not found"})
			return
		}
		c.AbortWithStatusJSON(500, gin.H{"error": "Failed to check permissions"})
		return
	}

	if slices.Index(workspace.Members, userId.(bson.ObjectID)) == -1 && workspace.OwnedBy != userId.(bson.ObjectID) {
		c.AbortWithStatusJSON(403, gin.H{"error": "You are not a member of this workspace"})
		return
	}
	c.JSON(200, board)
}
