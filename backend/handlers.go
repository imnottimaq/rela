package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/goccy/go-json"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"golang.org/x/crypto/argon2"
)

var _ = godotenv.Load()
var pepper = os.Getenv("PEPPER")
var mongodbCredentials = os.Getenv("MONGO_CREDS")
var dbClient, _ = mongo.Connect(options.Client().ApplyURI(mongodbCredentials).SetServerAPIOptions(options.ServerAPI(options.ServerAPIVersion1)))

var tasksDb = dbClient.Database("rela").Collection("tasks")
var usersDb = dbClient.Database("rela").Collection("users")
var boardsDb = dbClient.Database("rela").Collection("boards")

var emailRegex = regexp.MustCompile(`^[a-z0-9._%+-]+@[a-z0-9.-]+\.[a-z]{2,}$`)

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
	cursor, _ := tasksDb.Find(context.TODO(), bson.D{{"created_by", id}})
	defer cursor.Close(context.TODO())
	var tasks []Task
	_ = cursor.All(context.TODO(), &tasks)
	c.IndentedJSON(200, tasks)
	return
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
	json.NewDecoder(c.Request.Body).Decode(&input)
	if input.Name == "" {
		c.AbortWithStatusJSON(400, gin.H{"error": "Field 'name' is not specified"})
		return
	} else if input.Board.IsZero() {
		c.AbortWithStatusJSON(400, gin.H{"error": "Field 'board' is not specified"})
		return
	}
	err := boardsDb.FindOne(context.TODO(), bson.D{{"_id", input.Board}}).Decode(&board)
	if errors.Is(err, mongo.ErrNoDocuments) {
		c.AbortWithStatusJSON(400, gin.H{"error": "Board does not exist"})
		return
	} else {
		c.AbortWithStatusJSON(500, gin.H{"error": err})
	}
	newTask := Task{
		Name:        input.Name,
		Description: input.Description,
		CreatedAt:   time.Now().UTC().Unix(),
		OwnedBy:     id.(bson.ObjectID),
		Board:       input.Board,
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
	id, _ := c.Get("id")
	var previousVersion Task
	var valuesToEdit EditTask
	taskId, _ := bson.ObjectIDFromHex(c.Param("taskId"))
	json.NewDecoder(c.Request.Body).Decode(&valuesToEdit)
	if taskId.IsZero() {
		c.AbortWithStatusJSON(400, gin.H{"error": "you must specify task id"})
		return
	}
	err := tasksDb.FindOne(context.TODO(), bson.D{{"_id", taskId}}).Decode(&previousVersion)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			c.AbortWithStatusJSON(404, gin.H{"error": "not found"})
			return
		} else {
			c.AbortWithStatusJSON(500, gin.H{"error": "Internal Server Error"})
			return
		}
	}
	if previousVersion.OwnedBy != id {
		c.AbortWithStatusJSON(400, gin.H{"error": "This task isn't owned by you"})
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
	id, _ := c.Get("id")
	taskId, _ := bson.ObjectIDFromHex(c.Param("taskId"))
	if taskId.IsZero() {
		c.AbortWithStatusJSON(400, gin.H{"error": "you must specify task id"})

	}
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
	if result.OwnedBy != id {
		c.AbortWithStatusJSON(400, gin.H{"error": "This task isn't owned by you"})
		return
	}
	_, err = tasksDb.DeleteOne(context.TODO(), bson.D{{Key: "_id", Value: taskId}})
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": "Failed to delete task"})
		return
	}
}

// @Summary Create new user
// @Router /api/v1/users/create [post]
// @Accept json
// @Success 200 {array} Token
// @Produce json
// @Tags Users
// @Param data body CreateUser true "Create user request"
func createUser(c *gin.Context) {
	randomBytes := make([]byte, 32)
	rand.Read(randomBytes)
	generatedSalt := base64.URLEncoding.EncodeToString(randomBytes)
	var input CreateUser
	var i bson.M
	json.NewDecoder(c.Request.Body).Decode(&input)
	if input.Email == "" {
		c.AbortWithStatusJSON(400, gin.H{"error": "email is required"})
		return
	} else if !emailRegex.MatchString(input.Email) {
		c.AbortWithStatusJSON(400, gin.H{"error": "bad email"})
		return
	} else if input.Name == "" {
		c.AbortWithStatusJSON(400, gin.H{"error": "name is required"})
		return
	} else if input.Password == "" {
		c.AbortWithStatusJSON(400, gin.H{"error": "password is required"})
		return
	} else if !validatePassword(input.Password) {
		c.AbortWithStatusJSON(400, gin.H{"": "your password must contain 1 uppercase letter, 1 lowercase letter, 1 special character and be atleast 8 characters long"})
		return
	}
	newUser := User{
		Salt:           generatedSalt,
		Name:           input.Name,
		HashedPassword: base64.RawStdEncoding.EncodeToString(argon2.IDKey([]byte(input.Password+pepper), []byte(generatedSalt), uint32(3), uint32(128*1024), uint8(2), uint32(32))),
		Email:          input.Email,
	}
	err := usersDb.FindOne(context.TODO(), bson.D{{Key: "email", Value: input.Email}}).Decode(&i)
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
				bearerToken, err := generateAccessToken(signedRefreshToken, "access")
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

// @Summary Login user
// @Router /api/v1/users/login [post]
// @Accept json
// @Success 200 {array} Token
// @Produce json
// @Tags Users
// @Param data body LoginUser true "Login user request"
func loginUser(c *gin.Context) {
	var input LoginUser
	var i User
	json.NewDecoder(c.Request.Body).Decode(&input)
	if input.Email == "" {
		c.AbortWithStatusJSON(400, gin.H{"error": "email is required"})
		return
	} else if !emailRegex.MatchString(input.Email) {
		c.AbortWithStatusJSON(400, gin.H{"error": "bad email"})
		return
	} else if input.Password == "" {
		c.AbortWithStatusJSON(400, gin.H{"error": "password is required"})
		return
	} // else if !validatePassword(input.Password) {
	//c.AbortWithStatusJSON(400, gin.H{"error": "password does not meet requirements"})
	//return
	//}
	if err := usersDb.FindOne(context.TODO(), bson.D{{Key: "email", Value: input.Email}}).Decode(&i); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			c.AbortWithStatusJSON(404, gin.H{"error": "User doesnt exist"})
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
			bearerToken, err := generateAccessToken(signedRefreshToken, "access")
			if err != nil {
				c.AbortWithStatusJSON(500, err)
				return
			}
			c.JSON(200, gin.H{"token": string(bearerToken)})
			return
		}
	}
}

// @Summary Delete user
// @Router /api/v1/users/delete [delete]
// @Accept json
// @Success 200
// @Tags Users
// @Param data body LoginUser true "Delete user request"
// @Param X-Authorization header string true "Bearer Token"
func deleteUser(c *gin.Context) {
	userId, _ := c.Get("id")
	var input LoginUser
	var user User
	json.NewDecoder(c.Request.Body).Decode(&input)
	if input.Email == "" {
		c.AbortWithStatusJSON(400, gin.H{"error": "email is required"})
		return
	} else if !emailRegex.MatchString(input.Email) {
		c.AbortWithStatusJSON(400, gin.H{"error": "bad email"})
		return
	} else if input.Password == "" {
		c.AbortWithStatusJSON(400, gin.H{"error": "password is required"})
		return
	} else if !validatePassword(input.Password) {
		c.AbortWithStatusJSON(400, gin.H{"error": "password does not meet requirements"})
		return
	}
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

func uploadAvatar(c *gin.Context) {
	userId, _ := c.Get("id")
	user := User{}
	avatar, err := c.FormFile("img")
	if err != nil {
		fmt.Printf(fmt.Sprintf("%v", err))
		c.AbortWithStatusJSON(400, gin.H{"error": "no file given"})
		return
	} else if filepath.Ext(avatar.Filename) != ".png" && filepath.Ext(avatar.Filename) != ".jpg" && filepath.Ext(avatar.Filename) != ".jpeg" {
		c.AbortWithStatusJSON(400, gin.H{"error": "wrong format"})
		return
	}
	usersDb.FindOne(context.TODO(), bson.D{{"_id", userId}}).Decode(&user)

	filename := filepath.Join("img", "/"+fmt.Sprintf("%v", uuid.New()))
	if err := c.SaveUploadedFile(avatar, filename); err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": "error while saving image"})
	} else {
		user.Avatar = filename
		usersDb.ReplaceOne(context.TODO(), bson.D{{"_id", userId}}, user)
		c.AbortWithStatus(200)
	}

}

func generateAccessToken(token string, tokenType string) (accessToken string, err error) {
	parsedToken, _ := jwt.ParseWithClaims(token, &LoginToken{}, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("Unknown signing method: ", token.Method)
		}
		return []byte(pepper), nil
	})
	claims := parsedToken.Claims.(*LoginToken)
	if claims.ExpiresAt < time.Now().UTC().Unix() {
		return "", fmt.Errorf("expired refresh token")
	} else {
		newToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"id":   claims.Id,
			"exp":  time.Now().UTC().Unix() + 300,
			"type": tokenType,
		})
		accessToken, _ = newToken.SignedString([]byte(pepper))
		return accessToken, nil
	}
}

// @Summary Refresh bearer token
// @Description For this route, you must have refresh token, that is sent to your browser when you log into user account as a http-only cookie
// @Router /api/v1/users/refresh [get]
// @Success 200 {array} Token
// @Produce json
// @Tags Users
func refreshAccessToken(c *gin.Context) {
	refreshToken, _ := c.Cookie("refreshToken")
	token, err := jwt.ParseWithClaims(refreshToken, &LoginToken{}, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("Unknown signing method: ", token.Method)
		}
		return []byte(pepper), nil
	})
	if err != nil {
		c.AbortWithStatusJSON(500, "Internal Server Error")
		return
	}
	claims := token.Claims.(*LoginToken)
	if claims.ExpiresAt < time.Now().UTC().Unix() {
		c.AbortWithStatusJSON(403, "Authorization Required")
		return
	} else if claims.Type == "access" {
		c.AbortWithStatusJSON(400, "Invalid Token")
	}
	bearerToken, err := generateAccessToken(refreshToken, "access")
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": err})
		return
	}
	c.AbortWithStatusJSON(200, gin.H{"token": bearerToken})
	return

}

// @Summary Create new board
// @Router /api/v1/boards [post]
// @Accept json
// @Success 200
// @Produce json
// @Tags Boards
// @Param data body CreateBoard true "Create board request"
// @Param X-Authorization header string true "Bearer Token"
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
	input.OwnedBy = user.Id
	boardsDb.InsertOne(context.TODO(), input)
	c.AbortWithStatus(200)
	return
}

// @Summary Delete board
// @Router /api/v1/boards/{boardId} [delete]
// @Accept json
// @Success 200
// @Produce json
// @Tags Boards
// @Param boardId path bson.ObjectID true "Board ID"
// @Param X-Authorization header string true "Bearer Token"
func deleteBoard(c *gin.Context) {
	id, _ := c.Get("id")
	boardId, _ := c.Get("boardId")
	var board Board
	if err := boardsDb.FindOne(context.TODO(), bson.D{{"_id", boardId}}).Decode(board); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			c.AbortWithStatusJSON(404, gin.H{"error": "board does not exist"})
			return
		} else {
			c.AbortWithStatusJSON(500, gin.H{"error": "failed to execute search"})
			return
		}
	}
	if board.OwnedBy == id {
		if _, err := boardsDb.DeleteOne(context.TODO(), bson.D{{"_id", boardId}}); err != nil {
			c.AbortWithStatusJSON(500, gin.H{"error": "failed to remove board"})
		}
		c.AbortWithStatus(200)
		return
	} else {
		c.AbortWithStatusJSON(400, gin.H{"error": "you dont own this board"})
	}
}

// @Summary Edit board
// @Router /api/v1/boards/{boardId} [patch]
// @Accept json
// @Success 200
// @Produce json
// @Tags Boards
// @Param boardId path bson.ObjectID true "Board ID"
// @Param body path CreateBoard true "Edit board request"
// @Param X-Authorization header string true "Bearer Token"
func editBoard(c *gin.Context) {
	id, _ := c.Get("id")
	boardId, _ := c.Get("boardId")
	var valuesToEdit Board
	var i Board
	json.NewDecoder(c.Request.Body).Decode(valuesToEdit)
	if valuesToEdit.OwnedBy != id && !valuesToEdit.OwnedBy.IsZero() {
		c.AbortWithStatusJSON(400, gin.H{"error": "you dont own this board"})
		return
	} else if valuesToEdit.OwnedBy.IsZero() {
		if err := boardsDb.FindOne(context.TODO(), bson.D{{"_id", boardId}}).Decode(i); err != nil {
			if boardId == "" {
				c.AbortWithStatusJSON(400, gin.H{"error": "board id cannot be empty"})
				return
			} else if errors.Is(err, mongo.ErrNoDocuments) {
				c.AbortWithStatusJSON(404, gin.H{"error": "board does not exist"})
				return
			} else {
				c.AbortWithStatusJSON(500, gin.H{"error": "failed to execute search"})
				return
			}
		}
		if valuesToEdit.Name == "" {
			c.AbortWithStatusJSON(400, gin.H{"error": "you cant change board name to nothing"})
			return
		}
		i.Name = valuesToEdit.Name
		if _, err := boardsDb.InsertOne(context.TODO(), i); err != nil {
			c.AbortWithStatusJSON(500, gin.H{"error": "failed to insert board"})
		}
	}
}

func validatePassword(password string) bool {
	if len(password) < 8 {
		return false
	}

	hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
	hasDigit := regexp.MustCompile(`\d`).MatchString(password)
	hasSpecial := regexp.MustCompile(`[@$!%*?&]`).MatchString(password)

	return hasLower && hasUpper && hasDigit && hasSpecial
}
