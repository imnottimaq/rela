package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/goccy/go-json"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"slices"
	"time"
)

func createWorkspace(c *gin.Context) {
	userId, _ := c.Get("id")
	var input Workspace
	var i Workspace
	if err := json.NewDecoder(c.Request.Body).Decode(&input); err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": "Failed to parse request"})
		return
	}
	if input.Name == "" {
		c.AbortWithStatusJSON(400, gin.H{"error": "Field 'name' is not specified"})
		return
	} else if err := workspacesDb.FindOne(context.TODO(), bson.D{{"name", input.Name}}).Decode(&i); err == nil {
		c.AbortWithStatusJSON(400, gin.H{"error": "Workspace with this name already exists"})
		return
	}
	output := Workspace{
		Name:    input.Name,
		OwnedBy: userId.(bson.ObjectID),
	}
	output.Members = append(output.Members, userId.(bson.ObjectID))
	if _, err := workspacesDb.InsertOne(context.TODO(), output); err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": "Internal Server Error"})
		return
	}
	c.AbortWithStatus(200)
}

func editWorkspace(c *gin.Context) {
	id, _ := c.Get("id")
	workspaceId := c.Param("workspaceId")
	previousVersion := Workspace{}
	input := Workspace{}
	if err := json.NewDecoder(c.Request.Body).Decode(&input); err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": "Failed to parse request"})
		return
	} else if input.Name == "" {
		c.AbortWithStatusJSON(400, gin.H{"error": "Name cant be null"})
		return
	}
	if err := workspacesDb.FindOne(context.TODO(), bson.D{{"_id", bson.ObjectIDFromHex(workspaceId)}}).Decode(&previousVersion); err != nil {
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
	previousVersion.Name = input.Name
	previousVersion.Avatar = input.Avatar
	if _, err := workspacesDb.ReplaceOne(context.TODO(), bson.D{{"_id", bson.ObjectIDFromHex(workspaceId)}}, previousVersion); err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": "Internal Server Error"})
		return
	}
}
func getAllWorkspaces(c *gin.Context) {
	id, _ := c.Get("id")
	cursor, _ := workspacesDb.Find(context.TODO(), bson.D{{"members", id}})
	defer cursor.Close(context.TODO())
	workspaces := []Workspace{}
	if err := cursor.All(context.TODO(), &workspaces); err != nil {
		c.JSON(500, gin.H{"error": "failed to decode workspaces"})
		return
	}
	c.IndentedJSON(200, workspaces)
}
func deleteWorkspace(c *gin.Context) {
	id, _ := c.Get("id")
	workspaceId := c.Param("workspaceId")
	previousVersion := Workspace{}
	if err := workspacesDb.FindOne(context.TODO(), bson.D{{"_id", bson.ObjectIDFromHex(workspaceId)}}).Decode(&previousVersion); err != nil {
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
	if _, err := workspacesDb.DeleteOne(context.TODO(), bson.D{{"_id", bson.ObjectIDFromHex(workspaceId)}}); err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": "Internal Server Error"})
		return
	}
}

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

func createNewInvite(c *gin.Context) {
	if workspaceId, exists := c.Get("workspaceId"); exists == true {
		var i Workspace
		if err := workspacesDb.FindOne(context.TODO(), bson.D{{"_id", workspaceId}}).Decode(&i); err != nil {
			c.AbortWithStatusJSON(404, gin.H{"error": "Not Found"})
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

func kickMember(c *gin.Context) {
	id, _ := c.Get("id")
	workspaceId := c.Param("workspaceId")
	var input KickUser
	json.NewDecoder(c.Request.Body).Decode(&input)
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
		} else if workspace.OwnedBy != id {
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

func promoteMember(c *gin.Context) {
	id, _ := c.Get("id")
	workspaceId := c.Param("workspaceId")
	var input KickUser
	json.NewDecoder(c.Request.Body).Decode(&input)
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
		} else if workspace.OwnedBy != id {
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
	defer cursor.Close(context.TODO())

	members := []Member{}
	if err := cursor.All(context.TODO(), &members); err != nil {
		c.JSON(500, gin.H{"error": "failed to decode users"})
		return
	}

	c.IndentedJSON(200, members)
}

func assignTask(c *gin.Context) {
	id, _ := c.Get("id")
	workspaceId := c.Param("workspaceId")
	var input struct {
		TaskId bson.ObjectID `bson:"taskId" json:"taskId"`
		UserId bson.ObjectID `bson:"userId" json:"userId"`
	}
	json.NewDecoder(c.Request.Body).Decode(&input)
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
	if input.UserId == id || workspace.OwnedBy == id {
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
		task.AssignedTo = id.(bson.ObjectID)
		if _, err := tasksDb.ReplaceOne(context.TODO(), bson.D{{"_id", input.TaskId}}, task); err != nil {
			c.AbortWithStatusJSON(500, gin.H{"error": "Internal Server Error"})
			return
		}
		c.AbortWithStatus(200)
		return
	}
}
