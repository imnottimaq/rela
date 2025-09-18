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
	json.NewDecoder(c.Request.Body).Decode(&input)
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

func assignTask(c *gin.Context) {

}

func getAllMembers(c *gin.Context) {

}
