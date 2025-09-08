package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/goccy/go-json"
	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"golang.org/x/crypto/argon2"
)

var _ = godotenv.Load()
var pepper = os.Getenv("PEPPER")
var tasksDb = dbClient.Database("rela").Collection("tasks")
var usersDb = dbClient.Database("rela").Collection("users")

func getAllTasks(c *gin.Context) {
	id, _ := c.Get("id")
	cursor, _ := tasksDb.Find(context.TODO(), bson.D{{Key: "createdby", Value: id}})
	defer cursor.Close(context.TODO())
	var tasks []Task
	_ = cursor.All(context.TODO(), &tasks)
	c.IndentedJSON(200, tasks)
	return
}

func createNewTask(c *gin.Context) {
	id, _ := c.Get("id")
	var newTask Task
	json.NewDecoder(c.Request.Body).Decode(&newTask)
	if newTask.Name == "" {
		c.AbortWithStatusJSON(400, gin.H{"error": "Field 'name' is not specified"})
		return
	}
	newTask.CreatedAt = time.Now().UTC().Unix()
	newTask.CreatedBy = id.(bson.ObjectID)
	task, err := tasksDb.InsertOne(context.TODO(), newTask)
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": "Failed to push task to database"})
		return
	}
	newTask.Id = task.InsertedID.(bson.ObjectID)
	c.AbortWithStatus(200)
	return
}

func editExistingTask(c *gin.Context) {
	id, _ := c.Get("id")
	var previousVersion Task
	taskId, _ := bson.ObjectIDFromHex(c.Param("taskId"))
	err := tasksDb.FindOne(context.TODO(), bson.D{{Key: "_id", Value: taskId}}).Decode(&previousVersion)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			c.AbortWithStatusJSON(400, gin.H{"error": "There is no task with specified id"})
			return
		} else {
			c.AbortWithStatusJSON(500, gin.H{"error": "Internal Server Error"})
			return
		}
	}
	if previousVersion.CreatedBy != id {
		c.AbortWithStatusJSON(400, gin.H{"error": "This task isn't owned by you"})
		return
	}
	var valuesToEdit TaskEdit
	json.NewDecoder(c.Request.Body).Decode(&valuesToEdit)
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
	if valuesToEdit.Board != nil {
		previousVersion.Board = valuesToEdit.Board
	}
	_, err = tasksDb.ReplaceOne(context.TODO(), bson.D{{Key: "_id", Value: taskId}}, previousVersion)
	if err != nil {
		c.AbortWithStatusJSON(400, gin.H{"error": "Failed to change task"})
		return
	}
	c.AbortWithStatus(200)
	return
}

func deleteExistingTask(c *gin.Context) {
	id, _ := c.Get("id")
	taskId, _ := bson.ObjectIDFromHex(c.Param("id"))
	var result Task
	err := tasksDb.FindOne(context.TODO(), bson.D{{Key: "_id", Value: taskId}}).Decode(&result)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			c.AbortWithStatusJSON(404, gin.H{"error": "There is no task with that id"})
			return
		} else {
			c.AbortWithStatusJSON(500, gin.H{"error": "Internal Server Error"})
			return
		}
	}
	if result.CreatedBy != id {
		c.AbortWithStatusJSON(400, gin.H{"error": "This task isn't owned by you"})
		return
	}
	_, err = tasksDb.DeleteOne(context.TODO(), bson.D{{Key: "_id", Value: taskId}})
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": "Failed to delete task"})
		return
	}
}

func createUser(c *gin.Context) {
	randomBytes := make([]byte, 32)
	rand.Read(randomBytes)
	generatedSalt := base64.URLEncoding.EncodeToString(randomBytes)
	var input CreateUser
	var i bson.M
	json.NewDecoder(c.Request.Body).Decode(&input)
	newUser := User{
		Salt:           generatedSalt,
		Name:           input.Name,
		HashedPassword: base64.RawStdEncoding.EncodeToString(argon2.IDKey([]byte(input.Password+pepper), []byte(generatedSalt), uint32(3), uint32(128*1024), uint8(2), uint32(32))),
		Email:          input.Email,
	}
	err := usersDb.FindOne(context.TODO(), bson.D{{Key: "email", Value: input.Email}}).Decode(i)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			user, _ := usersDb.InsertOne(context.TODO(), &newUser)
			refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
				"id":  user.InsertedID.(bson.ObjectID),
				"exp": time.Now().UTC().Unix() + 604800, // Current expire time is 7 days, this is subject to change
			})
			signedRefreshToken, err := refreshToken.SignedString([]byte(pepper))
			if err != nil {
				fmt.Println(err)
				c.AbortWithStatusJSON(500, gin.H{"error": err})
				return
			} else {
				c.SetCookie("refreshToken", signedRefreshToken, 604800, "/", "", true, true)
				bearerToken, err := generateAccessToken(signedRefreshToken)
				if err != nil {
					fmt.Println(err)
					c.AbortWithStatusJSON(500, err)
					return
				}
				c.AbortWithStatusJSON(200, gin.H{"token": bearerToken})
				return
			}
		}
	}
	c.AbortWithStatusJSON(400, gin.H{"error": err})
	return
}

func loginUser(c *gin.Context) {
	var input LoginUser
	var i User
	json.NewDecoder(c.Request.Body).Decode(&input)
	if err := usersDb.FindOne(context.TODO(), bson.D{{Key: "email", Value: input.Email}}).Decode(i); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			c.AbortWithStatusJSON(400, gin.H{"error": "User doesnt exist"})
			return
		} else {
			c.AbortWithStatusJSON(500, gin.H{"error": err})
			return
		}
	}
	if base64.RawStdEncoding.EncodeToString(argon2.IDKey([]byte(input.Password+pepper), []byte(i.Salt), uint32(3), uint32(128*1024), uint8(2), uint32(32))) == i.HashedPassword {
		refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"id":  i.Id,
			"exp": time.Now().UTC().Unix() + 604800, // Current expire time is 7 days, this is subject to change
		})
		signedRefreshToken, err := refreshToken.SignedString([]byte(pepper))
		if err != nil {
			c.AbortWithStatusJSON(500, gin.H{"error": err})
			return
		} else {
			c.SetCookie("refreshToken", signedRefreshToken, 604800, "/", "", true, true)
			bearerToken, err := generateAccessToken(signedRefreshToken)
			if err != nil {
				c.AbortWithStatusJSON(500, err)
				return
			}
			c.AbortWithStatusJSON(200, gin.H{"token": bearerToken})
			return
		}
	}
}

func deleteUser(c *gin.Context) {
	userId, _ := c.Get("id")
	var input LoginUser
	var user User
	json.NewDecoder(c.Request.Body).Decode(&input)
	usersDb.FindOne(context.TODO(), bson.D{{Key: "_id", Value: userId}}).Decode(&user)
	if user.Email == input.Email && base64.RawStdEncoding.EncodeToString(argon2.IDKey([]byte(input.Password+pepper), []byte(user.Salt), uint32(3), uint32(128*1024), uint8(2), uint32(32))) == user.HashedPassword && userId == user.Id {
		usersDb.DeleteOne(context.TODO(), bson.D{{Key: "_id", Value: userId}})
		c.AbortWithStatus(200)
		return
	} else if user.Email != input.Email || base64.RawStdEncoding.EncodeToString(argon2.IDKey([]byte(input.Password+pepper), []byte(user.Salt), uint32(3), uint32(128*1024), uint8(2), uint32(32))) != user.HashedPassword {
		c.AbortWithStatus(400)
		return
	} else if userId != user.Id {
		c.AbortWithStatus(403)
		return
	}
}

func generateAccessToken(token string) (accessToken string, err error) {
	parsedToken, _ := jwt.ParseWithClaims(token, &LoginToken{}, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("Unknown signing method: ", token.Method)
		}
		return []byte(pepper), nil
	})
	claims := parsedToken.Claims.(*LoginToken)
	if claims.ExpiresAt < time.Now().UTC().Unix() {
		return "", fmt.Errorf("Expired refresh token")
	} else {
		newToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"id":  claims.Id,
			"exp": time.Now().UTC().Unix() + 300,
		})
		accessToken, _ = newToken.SignedString([]byte(pepper))
		return accessToken, nil
	}
}

func refreshAccessToken(c *gin.Context) {
	refreshToken, _ := c.Cookie("refreshToken")
	bearerToken, err := generateAccessToken(refreshToken)
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": err})
		return
	}
	c.AbortWithStatusJSON(200, gin.H{"token": bearerToken})
	return

}

func addBoard(c *gin.Context) {
	id, _ := c.Get("id")
	var input Board
	var user User
	json.NewDecoder(c.Request.Body).Decode(&input)
	if input.Name == "" {
		c.AbortWithStatusJSON(400, gin.H{"error": "no name given"})
		return
	}
	usersDb.FindOne(context.TODO(), bson.D{{Key: "_id", Value: id}}).Decode(&user)
	user.Boards = append(user.Boards, input.Name)
	usersDb.ReplaceOne(context.TODO(), bson.D{{Key: "_id", Value: id}}, user)
	c.AbortWithStatusJSON(200, gin.H{"error": "no name given"})
	return
}
