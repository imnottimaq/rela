package main

import (
    "context"
    "errors"
    "fmt"
    "slices"
    "github.com/gin-gonic/gin"
    "github.com/goccy/go-json"
    "go.mongodb.org/mongo-driver/v2/bson"
    "go.mongodb.org/mongo-driver/v2/mongo"
    "log"
)

// @Summary Create new board
// @Router /boards [post]
// @Router /workspaces/{workspaceId}/boards [post]
// @Accept json
// @Success 200
// @Produce json
// @Tags Boards
// @Param data body CreateBoard true "Create board request"
// @Param workspaceId path string true "Workspace ID"
// @Param Authorization header string true "Bearer Token"
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
	print(fmt.Sprintf("%v", id))
	input.OwnedBy = id.(bson.ObjectID)
	if workspaceId != "" {
		i, _ := bson.ObjectIDFromHex(workspaceId)
		input.OwnedBy = i
	}
	if _, err = boardsDb.InsertOne(context.TODO(), input); err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": "Failed to create board"})
		return
	}
	c.AbortWithStatus(200)
}

// @Summary Delete board
// @Router /boards/{boardId} [delete]
// @Router /workspaces/{workspaceId}/boards/{boardId} [delete]
// @Accept json
// @Success 200
// @Produce json
// @Tags Boards
// @Param boardId path bson.ObjectID true "Board ID"
// @Param workspaceId path string true "Workspace ID"
// @Param Authorization header string true "Bearer Token"
func deleteBoard(c *gin.Context) {
	id, _ := c.Get("id")
	boardId := c.Param("boardId")
	workspaceId := c.Param("workspaceId")
	i, _ := bson.ObjectIDFromHex(workspaceId)
	var board Board
	if err := boardsDb.FindOne(context.TODO(), bson.D{{"_id", boardId}}).Decode(board); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			c.AbortWithStatusJSON(404, gin.H{"error": "Board does not exist"})
			return
		} else {
			c.AbortWithStatusJSON(500, gin.H{"error": "Failed to execute search"})
			return
		}
	}
	if board.OwnedBy == id || board.OwnedBy == i {
		if _, err := boardsDb.DeleteOne(context.TODO(), bson.D{{"_id", boardId}}); err != nil {
			c.AbortWithStatusJSON(500, gin.H{"error": "Failed to remove board"})
		}
		c.AbortWithStatus(200)
		return
	} else {
		c.AbortWithStatusJSON(400, gin.H{"error": "You dont own this board"})
	}
}

// @Summary Edit board
// @Router /boards/{boardId} [patch]
// @Router /workspaces/{workspaceId}/boards/{boardId} [patch]
// @Accept json
// @Success 200
// @Produce json
// @Tags Boards
// @Param boardId path bson.ObjectID true "Board ID"
// @Param workspaceId path string true "Workspace ID"
// @Param body path CreateBoard true "Edit board request"
// @Param Authorization header string true "Bearer Token"
func editBoard(c *gin.Context) {
	id, _ := c.Get("id")
	boardId := c.Param("boardId")
	workspaceId := c.Param("workspaceId")
	i, _ := bson.ObjectIDFromHex(workspaceId)
	var valuesToEdit Board
	var a Board
	err := json.NewDecoder(c.Request.Body).Decode(&valuesToEdit)
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": "Something went wrong when parsing request"})
		return
	} else if valuesToEdit.OwnedBy != id || valuesToEdit.OwnedBy != i {
		c.AbortWithStatusJSON(400, gin.H{"error": "You dont own this board"})
		return
	} else if valuesToEdit.OwnedBy.IsZero() {
		if err := boardsDb.FindOne(context.TODO(), bson.D{{"_id", boardId}}).Decode(i); err != nil {
			if boardId == "" {
				c.AbortWithStatusJSON(400, gin.H{"error": "Board id cannot be empty"})
				return
			} else if errors.Is(err, mongo.ErrNoDocuments) {
				c.AbortWithStatusJSON(404, gin.H{"error": "Board does not exist"})
				return
			} else {
				c.AbortWithStatusJSON(500, gin.H{"error": "Failed to execute search"})
				return
			}
		}
		if valuesToEdit.Name == "" {
			c.AbortWithStatusJSON(400, gin.H{"error": "You cant change board name to nothing"})
			return
		}
		a.Name = valuesToEdit.Name
		if _, err := boardsDb.InsertOne(context.TODO(), i); err != nil {
			c.AbortWithStatusJSON(500, gin.H{"error": "Failed to insert board"})
		}
	}
}

// @Summary Get all boards
// @Router /boards [get]
// @Router /workspaces/{workspaceId}/boards [get]
// @Accept json
// @Success 200 {array} Board
// @Produce json
// @Tags Boards
// @Param workspaceId path string true "Workspace ID"
// @Param Authorization header string true "Bearer Token"
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
		log.Print("Failed to close cursor")
	}
	return
}

// @Summary Get board by id
// @Router /boards/{boardId} [get]
// @Router /workspaces/{workspaceId}/boards/{boardId} [get]
// @Accept json
// @Success 200 {object} Board
// @Produce json
// @Tags Boards
// @Param boardId path string true "Board ID"
// @Param workspaceId path string true "Workspace ID"
// @Param Authorization header string true "Bearer Token"
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
