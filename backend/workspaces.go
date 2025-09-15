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
)

func createWorkspace(c *gin.Context) {
	userId, _ := c.Get("id")
	var input Workspace
	json.NewDecoder(c.Request.Body).Decode(&input)
	if input.Name == "" {

	}
	output := Workspace{
		Name:    input.Name,
		Members: userId.([]bson.ObjectID),
	}
	workspacesDb.InsertOne(context.TODO(), output)
}

func addMember(c *gin.Context) {
	userId, _ := c.Get("id")
	joinToken, _ := c.Get("joinToken")
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

		} else {

		}
	}
	_ = append(workspace.Members, userId.(bson.ObjectID))
	if _, err := workspacesDb.ReplaceOne(context.TODO(), bson.D{{"_id", claims.Id}}, workspace); err != nil {

	}
}
