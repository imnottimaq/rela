package main

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/goccy/go-json"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// @Summary Create new workspace
// @Router /api/v1/workspaces/create [post]
// @Accept json
// @Success 200
// @Tags Workspaces
// @Param data body CreateBoard true "Create new workspace"
// @Param Authorization header string true "Bearer Token"
func createWorkspace(c *gin.Context) {
	id, _ := c.Get("id")
	userId := id.(bson.ObjectID)
	var input Workspace
	var i Workspace
	if err := json.NewDecoder(c.Request.Body).Decode(&input); err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": "Something went wrong when parsing request"})
		return
	} else if input.Name == "" {
		c.AbortWithStatusJSON(400, gin.H{"error": "Field 'name' is not specified"})
		return
	} else if err := workspacesDb.FindOne(context.TODO(), bson.D{{"name", input.Name}}).Decode(&i); err == nil {
		c.AbortWithStatusJSON(400, gin.H{"error": "Workspace with this name already exists"})
		return
	}
	output := Workspace{
		Name:    input.Name,
		OwnedBy: userId,
	}
	output.Members = append(output.Members, userId)
	workspace, err := workspacesDb.InsertOne(context.TODO(), output)
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": "Internal Server Error"})
		return
	}
	output.Id = workspace.InsertedID.(bson.ObjectID)
	c.AbortWithStatusJSON(200, output)
}

// @Summary Edit workspace
// @Router /api/v1/workspaces/{workspaceId}/ [patch]
// @Success 200
// @Tags Workspaces
// @Param Authorization header string true "Bearer Token"
// @Param workspaceId path string true "Workspace ID"
func editWorkspace(c *gin.Context) {
	id, _ := c.Get("id")
	userId := id.(bson.ObjectID)
	wId := c.Param("workspaceId")
	workspaceId, _ := bson.ObjectIDFromHex(wId)
	previousVersion := Workspace{}
	input := Workspace{}
	if err := json.NewDecoder(c.Request.Body).Decode(&input); err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": "Failed to parse request"})
		return
	} else if input.Name == "" {
		c.AbortWithStatusJSON(400, gin.H{"error": "Name cant be null"})
		return
	}
	if err := workspacesDb.FindOne(context.TODO(), bson.D{{"_id", workspaceId}}).Decode(&previousVersion); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			c.AbortWithStatusJSON(404, gin.H{"error": "Not Found"})
			return
		} else {
			c.AbortWithStatusJSON(500, gin.H{"error": "Internal Server Error"})
			return
		}
	}
	if previousVersion.OwnedBy != userId {
		c.AbortWithStatusJSON(400, gin.H{"error": "You are not an owner of this workplace"})
		return
	}
	previousVersion.Name = input.Name
	previousVersion.Avatar = input.Avatar
	if _, err := workspacesDb.ReplaceOne(context.TODO(), bson.D{{"_id", workspaceId}}, previousVersion); err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": "Internal Server Error"})
		return
	}
}

// @Summary Get all workspaces for the current user
// @Router /api/v1/users/workspaces [get]
// @Success 200
// @Tags Workspaces
// @Param Authorization header string true "Bearer Token"
func getAllWorkspaces(c *gin.Context) {
	id, _ := c.Get("id")
	userId := id.(bson.ObjectID)
	cursor, _ := workspacesDb.Find(context.TODO(), bson.D{{"members", userId}})
	workspaces := make([]Workspace, 0)
	if err := cursor.All(context.TODO(), &workspaces); err != nil {
		c.JSON(500, gin.H{"error": "failed to decode workspaces"})
		return
	} else if err := cursor.Close(context.TODO()); err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": "Internal Server Error"})
		return
	} else {
		c.IndentedJSON(200, gin.H{"workspaces": workspaces})
		return
	}
}

// @Summary Delete workspace
// @Router /api/v1/workspaces/{workspaceId}/ [delete]
// @Success 200
// @Tags Workspaces
// @Param Authorization header string true "Bearer Token"
// @Param workspaceId path string true "Workspace ID"
func deleteWorkspace(c *gin.Context) {
	id, _ := c.Get("id")
	workspaceId := c.Param("workspaceId")
	previousVersion := Workspace{}
	if err := workspacesDb.FindOne(context.TODO(), bson.D{{"_id", workspaceId}}).Decode(&previousVersion); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			c.AbortWithStatusJSON(404, gin.H{"error": "Not Found"})
			return
		} else {
			c.AbortWithStatusJSON(500, gin.H{"error": "Internal Server Error"})
			return
		}
	}
	if previousVersion.OwnedBy != id {
		c.AbortWithStatusJSON(400, gin.H{"error": "You are not an owner of this workplace"})
		return
	}
	if _, err := workspacesDb.DeleteOne(context.TODO(), bson.D{{"_id", workspaceId}}); err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": "Internal Server Error"})
		return
	}
}

// @Summary Accept invite
// @Router /api/v1/workspaces/add/{joinToken} [post]
// @Success 200
// @Tags Workspaces
// @Param Authorization header string true "Bearer Token"
// @Param joinToken path string true "Join Token"
func addMember(c *gin.Context) {
	userId, _ := c.Get("id")
	joinToken := c.Param("joinToken")
	var workspace Workspace
	token, _ := jwt.ParseWithClaims(fmt.Sprintf("%v", joinToken), &Token{}, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unknown signing method: %s", token.Method)
		}
		return []byte(pepper), nil
	})
	claims := token.Claims.(*Token)
	if err := workspacesDb.FindOne(context.TODO(), bson.D{{"_id", claims.Id}}).Decode(&workspace); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			c.AbortWithStatusJSON(404, gin.H{"error": "Workspace does not exist"})
			return
		} else {
			c.AbortWithStatusJSON(500, gin.H{"error": "Internal Server Error"})
			return
		}
	}
	_ = append(workspace.Members, userId.(bson.ObjectID))
	if _, err := workspacesDb.ReplaceOne(context.TODO(), bson.D{{"_id", claims.Id}}, workspace); err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": "Internal Server Error"})
		return
	}
	c.AbortWithStatus(200)
}

// @Summary Create invite to the workspace
// @Router /api/v1/workspaces/{workspaceId}/new_invite [get]
// @Success 200
// @Tags Workspaces
// @Param Authorization header string true "Bearer Token"
// @Param workspaceId path string true "Workspace ID"
func createNewInvite(c *gin.Context) {
	userId, _ := c.Get("id")
	if workspaceId := c.Param("workspaceId"); workspaceId != "" {
		var i Workspace
		id, _ := bson.ObjectIDFromHex(workspaceId)
		if err := workspacesDb.FindOne(context.TODO(), bson.D{{"_id", id}}).Decode(&i); err != nil {
			c.AbortWithStatusJSON(404, gin.H{"error": "Not Found"})
			return
		} else if i.OwnedBy != userId.(bson.ObjectID) {
			c.AbortWithStatusJSON(404, gin.H{"error": "You are not an owner of this workplace"})
			return
		} else {
			newToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
				"id":   workspaceId,
				"exp":  time.Now().UTC().Unix() + 1209600, //2 weeks
				"type": "invite",
			})
			inviteToken, _ := newToken.SignedString([]byte(pepper))
			c.AbortWithStatusJSON(200, gin.H{"token": inviteToken})
			return
		}
	} else {
		c.AbortWithStatusJSON(400, gin.H{"error": "You must specify workspace id"})
		return
	}
}

// @Summary Kick member
// @Router /api/v1/workspaces/{workspaceId}/kick [delete]
// @Success 200
// @Tags Workspaces
// @Param Authorization header string true "Bearer Token"
// @Param workspaceId path string true "Workspace ID"
func kickMember(c *gin.Context) {
	id, _ := c.Get("id")
	userId := id.(bson.ObjectID)
	workspaceId := c.Param("workspaceId")
	var input KickUser
	if err := json.NewDecoder(c.Request.Body).Decode(&input); err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": "Something went wrong when parsing request"})
		return
	}
	if workspaceId != "" {
		i, _ := bson.ObjectIDFromHex(workspaceId)
		workspace := Workspace{}
		if err := workspacesDb.FindOne(context.TODO(), bson.D{{"_id", i}}).Decode(&workspace); err != nil {
			if errors.Is(err, mongo.ErrNoDocuments) {
				c.AbortWithStatusJSON(404, gin.H{"error": "Not Found"})
				return
			} else {
				c.AbortWithStatusJSON(500, gin.H{"error": "Internal server error"})
				return
			}
		} else if workspace.OwnedBy != userId {
			c.AbortWithStatusJSON(500, gin.H{"error": "You are not owner of this workspace"})
			return
		}
		workspace.Members = slices.Delete(workspace.Members, slices.Index(workspace.Members, input.Id), slices.Index(workspace.Members, input.Id)+1)
		if _, err := workspacesDb.ReplaceOne(context.TODO(), bson.D{{"_id", workspace.Id}}, workspace); err != nil {
			c.AbortWithStatusJSON(500, gin.H{"error": "Internal server error"})
			return
		}
		c.AbortWithStatus(200)
		return
	} else {
		c.AbortWithStatusJSON(400, gin.H{"error": "You must specify workspace id"})
		return
	}
}

// @Summary Promote member
// @Router /api/v1/workspaces/{workspaceId}/promote/{userId} [patch]
// @Success 200
// @Tags Workspaces
// @Param Authorization header string true "Bearer Token"
// @Param workspaceId path string true "Workspace ID"
// @Param userId path string true "User ID"
func promoteMember(c *gin.Context) {
	id, _ := c.Get("id")
	userId := id.(bson.ObjectID)
	workspaceId := c.Param("workspaceId")
	var input KickUser
	if err := json.NewDecoder(c.Request.Body).Decode(&input); err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": "Something went wrong when parsing request"})
		return
	}
	if workspaceId != "" {
		i, _ := bson.ObjectIDFromHex(workspaceId)
		workspace := Workspace{}
		if err := workspacesDb.FindOne(context.TODO(), bson.D{{"_id", i}}).Decode(&workspace); err != nil {
			if errors.Is(err, mongo.ErrNoDocuments) {
				c.AbortWithStatusJSON(404, gin.H{"error": "Not Found"})
				return
			} else {
				c.AbortWithStatusJSON(500, gin.H{"error": "Internal server error"})
				return
			}
		} else if workspace.OwnedBy != userId {
			c.AbortWithStatusJSON(500, gin.H{"error": "You are not owner of this workspace"})
			return
		} else if slices.Index(workspace.Members, input.Id) == -1 {
			c.AbortWithStatusJSON(400, gin.H{"error": "User is not part of this workspace"})
			return
		}
		workspace.OwnedBy = input.Id
		if _, err := workspacesDb.ReplaceOne(context.TODO(), bson.D{{"_id", workspace.Id}}, workspace); err != nil {
			c.AbortWithStatusJSON(500, gin.H{"error": "Internal server error"})
			return
		}
		c.AbortWithStatus(200)
		return
	} else {

	}
}

// @Summary Get all members info
// @Router /api/v1/workspaces/{workspaceId}/members [get]
// @Success 200
// @Tags Workspaces
// @Param Authorization header string true "Bearer Token"
// @Param workspaceId path string true "Workspace ID"
func getAllMembers(c *gin.Context) {
	id, _ := c.Get("id")
	userId := id.(bson.ObjectID)
	workspaceIdStr := c.Param("workspaceId")
	workspaceId, err := bson.ObjectIDFromHex(workspaceIdStr)
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid workspaceId"})
		return
	}

	workspace := Workspace{}
	if err := workspacesDb.FindOne(context.TODO(), bson.M{"_id": workspaceId}).Decode(&workspace); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			c.JSON(404, gin.H{"error": "workspace not found"})
			return
		}
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	if slices.Index(workspace.Members, userId) == -1 {
		c.JSON(403, gin.H{"error": "not a member of this workspace"})
		return
	}
	cursor, err := usersDb.Find(context.TODO(), bson.M{
		"_id": bson.M{"$in": workspace.Members},
	})
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to fetch users"})
		return
	}
	members := make([]Member, 0)
	if err := cursor.All(context.TODO(), &members); err != nil {
		c.JSON(500, gin.H{"error": "failed to decode users"})
		return
	} else if err := cursor.Close(context.TODO()); err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": "Internal Server Error"})
		return
	} else {
		c.IndentedJSON(200, gin.H{"members": members})
		return
	}

}

// @Summary Assign task to someone
// @Router /api/v1/workspaces/{workspaceId}/assign [post]
// @Accept json
// @Success 200
// @Tags Tasks
// @Param data body AssignTask true "Assign Task"
// @Param Authorization header string true "Bearer Token"
// @Param workspaceId path string true "Workspace ID"
func assignTask(c *gin.Context) {
	id, _ := c.Get("id")
	userId := id.(bson.ObjectID)
	wId := c.Param("workspaceId")
	workspaceId, _ := bson.ObjectIDFromHex(wId)
	var input AssignTask
	if err := json.NewDecoder(c.Request.Body).Decode(&input); err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": "Something went wrong when parsing request"})
		return
	}
	workspace := Workspace{}
	if err := workspacesDb.FindOne(context.TODO(), bson.D{{"_id", workspaceId}}).Decode(&workspace); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			c.AbortWithStatusJSON(404, gin.H{"error": "Workspace does not exist"})
			return
		} else {
			c.AbortWithStatusJSON(500, gin.H{"error": "Internal Server Error"})
			return
		}
	}
	if input.UserId == userId || workspace.OwnedBy == userId {
		task := Task{}
		if err := tasksDb.FindOne(context.TODO(), bson.D{{"_id", input.TaskId}}).Decode(&task); err != nil {
			if errors.Is(err, mongo.ErrNoDocuments) {
				c.AbortWithStatusJSON(404, gin.H{"error": "Task does not exist"})
				return
			} else {
				c.AbortWithStatusJSON(500, gin.H{"error": "Internal Server Error"})
				return
			}
		}
		task.AssignedTo = userId
		insert, err := tasksDb.ReplaceOne(context.TODO(), bson.D{{"_id", input.TaskId}}, task)
		if err != nil {
			c.AbortWithStatusJSON(500, gin.H{"error": "Internal Server Error"})
			return
		}
		task.Id = insert.UpsertedID.(bson.ObjectID)
		c.AbortWithStatusJSON(200, task)
		return
	}
}
