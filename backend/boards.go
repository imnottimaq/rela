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
// @Description 	Creates a new board for the current user or a workspace.
// @Router 			/boards [post]
// @Router 			/workspaces/{workspaceId}/boards [post]
// @Tags 			Boards
// @Security 		BearerAuth
// @Accept 			json
// @Produce 		json
// @Param 			workspaceId path string false "Workspace ID (optional)"
// @Param 			data body CreateBoard true "Board creation data"
// @Success 		200 {object} Board "The created board"
// @Failure 		400 {object} ErrorSwagger "Bad request - no name given"
// @Failure 		500 {object} ErrorSwagger "Internal server error"
func addBoard(c *gin.Context) {
	id, _ := c.Get("id")
	workspaceId := c.Param("workspaceId")
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
	input.OwnedBy = id.(bson.ObjectID)
	if workspaceId != "" {
		i, _ := bson.ObjectIDFromHex(workspaceId)
		input.OwnedBy = i
	}
	result, err := boardsDb.InsertOne(context.TODO(), input)
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": "Failed to create board"})
		return
	}
	input.Id = result.InsertedID.(bson.ObjectID)
	c.JSON(200, input)
}

// @Summary 		Delete a board
// @Description 	Deletes a specific board.
// @Router 			/boards/{boardId} [delete]
// @Router 			/workspaces/{workspaceId}/boards/{boardId} [delete]
// @Tags 			Boards
// @Security 		BearerAuth
// @Produce 		json
// @Param 			boardId path string true "Board ID"
// @Param 			workspaceId path string false "Workspace ID (optional)"
// @Success 		200 "Board deleted successfully"
// @Failure 		403 {object} ErrorSwagger "Forbidden - you do not own this board"
// @Failure 		404 {object} ErrorSwagger "Not Found - board not found"
// @Failure 		500 {object} ErrorSwagger "Internal server error"
func deleteBoard(c *gin.Context) {
	userId, _ := c.Get("id")
	boardIdStr := c.Param("boardId")
	boardId, err := bson.ObjectIDFromHex(boardIdStr)
	if err != nil {
		c.AbortWithStatusJSON(400, gin.H{"error": "invalid boardId"})
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

	if board.OwnedBy != userId.(bson.ObjectID) {
		c.AbortWithStatusJSON(403, gin.H{"error": "You dont own this board"})
		return
	}

	if _, err := boardsDb.DeleteOne(context.TODO(), bson.D{{"_id", boardId}}); err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": "Failed to remove board"})
		return
	}
	c.AbortWithStatus(200)
}

// @Summary 		Edit a board
// @Description 	Edits the details of a specific board.
// @Router 			/boards/{boardId} [patch]
// @Router 			/workspaces/{workspaceId}/boards/{boardId} [patch]
// @Tags 			Boards
// @Security 		BearerAuth
// @Accept 			json
// @Produce 		json
// @Param 			boardId path string true "Board ID"
// @Param 			workspaceId path string false "Workspace ID (optional)"
// @Param 			data body CreateBoard true "Fields to edit in the board"
// @Success 		200 {object} Board "The updated board"
// @Failure 		400 {object} ErrorSwagger "Bad request - invalid input"
// @Failure 		403 {object} ErrorSwagger "Forbidden - you do not own this board"
// @Failure 		404 {object} ErrorSwagger "Not Found - board not found"
// @Failure 		500 {object} ErrorSwagger "Internal server error"
func editBoard(c *gin.Context) {
	userId, _ := c.Get("id")
	boardIdStr := c.Param("boardId")
	boardId, err := bson.ObjectIDFromHex(boardIdStr)
	if err != nil {
		c.AbortWithStatusJSON(400, gin.H{"error": "invalid boardId"})
		return
	}

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

	if board.OwnedBy != userId.(bson.ObjectID) {
		c.AbortWithStatusJSON(403, gin.H{"error": "You dont own this board"})
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
// @Description 	Returns all boards owned by the current user, or all boards for a given workspace.
// @Router 			/boards [get]
// @Router 			/workspaces/{workspaceId}/boards [get]
// @Tags 			Boards
// @Security 		BearerAuth
// @Produce 		json
// @Param 			workspaceId path string false "Workspace ID (optional)"
// @Success 		200 {object} AllBoardsResponse "A list of boards"
// @Failure 		500 {object} ErrorSwagger "Internal Server Error"
func getAllBoards(c *gin.Context) {
	var cursor *mongo.Cursor
	userId, _ := c.Get("id")
	workspaceId := c.Param("workspaceId")
	if workspaceId != "" {
		i, _ := bson.ObjectIDFromHex(workspaceId)
		cursor, _ = boardsDb.Find(context.TODO(), bson.D{{"owned_by", i}})
	} else {
		cursor, _ = boardsDb.Find(context.TODO(), bson.D{{"owned_by", userId}})
	}
	boards := make([]Board, 0)
	_ = cursor.All(context.TODO(), &boards)
	c.IndentedJSON(200, gin.H{"boards": boards})
	if err := cursor.Close(context.TODO()); err != nil {
		// Log the error but don't abort, as the response has already been sent.
		println("Failed to close cursor")
	}
}

// @Summary 		Get a board by ID
// @Description 	Retrieves a specific board by its ID.
// @Router 			/boards/{boardId} [get]
// @Router 			/workspaces/{workspaceId}/boards/{boardId} [get]
// @Tags 			Boards
// @Security 		BearerAuth
// @Produce 		json
// @Param 			boardId path string true "Board ID"
// @Param 			workspaceId path string false "Workspace ID (optional, for authorization context)"
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

	var board Board
	if err := boardsDb.FindOne(context.TODO(), bson.M{"_id": boardId}).Decode(&board); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			c.AbortWithStatusJSON(404, gin.H{"error": "Board does not exist"})
			return
		}
		c.AbortWithStatusJSON(500, gin.H{"error": "Failed to fetch board"})
		return
	}

	// Authorization: allow if the board is owned by user or by a workspace where user is a member
	if board.OwnedBy == userId.(bson.ObjectID) {
		c.JSON(200, board)
		return
	}
	// If owned by a workspace, ensure membership
	var workspace Workspace
	if err := workspacesDb.FindOne(context.TODO(), bson.M{"_id": board.OwnedBy}).Decode(&workspace); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			// Not a workspace owner id; deny
			c.AbortWithStatusJSON(403, gin.H{"error": "forbidden"})
			return
		}
		c.AbortWithStatusJSON(500, gin.H{"error": "Failed to check permissions"})
		return
	}
	// Check membership
	if slices.Index(workspace.Members, userId.(bson.ObjectID)) == -1 && workspace.OwnedBy != userId.(bson.ObjectID) {
		c.AbortWithStatusJSON(403, gin.H{"error": "forbidden"})
		return
	}
	c.JSON(200, board)
}
