package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"golang.org/x/crypto/argon2"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/goccy/go-json"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// @Summary 		Create new user
// @Description 	Creates a new user and returns an access token.
// @Router 			/users/create [post]
// @Tags 			Users
// @Accept 			json
// @Produce 		json
// @Param 			data body CreateUser true "User creation data"
// @Success 		200 {object} TokenSwagger "Access token"
// @Failure 		400 {object} ErrorSwagger "Bad request - check your input"
// @Failure 		409 {object} ErrorSwagger "User with that email already exists"
// @Failure 		500 {object} ErrorSwagger "Internal server error"
func createUser(c *gin.Context) {
	randomBytes := make([]byte, 32)
	_, err := rand.Read(randomBytes)
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": "Internal Server Error"})
		return
	}
	generatedSalt := base64.URLEncoding.EncodeToString(randomBytes)
	var input CreateUser
	var i bson.M
	err = json.NewDecoder(c.Request.Body).Decode(&input)
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": "Something went wrong when parsing request"})
		return
	}
	if input.Email == "" {
		c.AbortWithStatusJSON(400, gin.H{"error": "Email is required"})
		return
	} else if !emailRegex.MatchString(input.Email) {
		c.AbortWithStatusJSON(400, gin.H{"error": "Bad email"})
		return
	} else if input.Password == "" {
		c.AbortWithStatusJSON(400, gin.H{"error": "Password is required"})
		return
	} else if len(input.Password) > 128 {
		c.AbortWithStatusJSON(400, gin.H{"error": "Password too long"})
		return
	} else if !validatePassword(input.Password) {
		c.AbortWithStatusJSON(400, gin.H{"error": "Password does not meet requirements"})
		return
	} else if input.Name == "" {
		c.AbortWithStatusJSON(400, gin.H{"error": "Name is required"})
		return
	}
	err = usersDb.FindOne(context.TODO(), bson.D{{Key: "email", Value: input.Email}}).Decode(&i)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			newUser := User{
				Salt:           generatedSalt,
				Name:           input.Name,
				HashedPassword: base64.RawStdEncoding.EncodeToString(argon2.IDKey([]byte(input.Password+pepper), []byte(generatedSalt), uint32(1), uint32(32*1024), uint8(4), uint32(32))),
				Email:          input.Email,
			}
			type dbResult struct {
				result *mongo.InsertOneResult
				err    error
			}
			dbChan := make(chan dbResult, 1)
			go func() {
				result, err := usersDb.InsertOne(context.TODO(), newUser)
				dbChan <- dbResult{result, err}
			}()
			dbRes := <-dbChan
			if dbRes.err != nil {
				c.AbortWithStatusJSON(500, gin.H{"error": "Internal Server Error"})
				return
			}
			claims := jwt.MapClaims{
				"exp": time.Now().UTC().Unix() + 604800, // Current expire time is 7 days, this is subject to change
			}
			claims["id"] = dbRes.result.InsertedID.(bson.ObjectID)
			refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
			signedRefreshToken, err := refreshToken.SignedString([]byte(pepper))
			if err != nil {
				fmt.Println(err)
				c.AbortWithStatusJSON(500, gin.H{"error": "Internal Server Error"})
				return
			} else {
				c.SetCookie("refreshToken", signedRefreshToken, 604800, "/", "", false, true)
				bearerToken, err := generateAccessToken(signedRefreshToken, "access")
				if err != nil {
					fmt.Println(err)
					c.AbortWithStatusJSON(500, gin.H{"error": "Internal Server Error"})
					return
				}
				c.AbortWithStatusJSON(200, gin.H{"token": bearerToken})
				return
			}
		}
	} else {
		c.AbortWithStatusJSON(409, gin.H{"error": "User with that email already exists"})
		return
	}
}

// @Summary 		Login user
// @Description 	Logs in a user and returns an access token.
// @Router 			/users/login [post]
// @Tags 			Users
// @Accept 			json
// @Produce 		json
// @Param 			data body LoginUser true "User login data"
// @Success 		200 {object} TokenSwagger "Access token"
// @Failure 		404 {object} ErrorSwagger "User not found"
// @Failure 		500 {object} ErrorSwagger "Internal server error"
func loginUser(c *gin.Context) {
	middlewareInput, _ := c.Get("input")
	input := middlewareInput.(LoginUser)
	var i User
	if err := usersDb.FindOne(context.TODO(), bson.D{{Key: "email", Value: input.Email}}).Decode(&i); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			c.AbortWithStatusJSON(404, gin.H{"error": "User doesnt exist"})
			return
		} else {
			c.AbortWithStatusJSON(500, gin.H{"error": "Internal Server Error"})
			return
		}
	}
	if base64.RawStdEncoding.EncodeToString(argon2.IDKey([]byte(input.Password+pepper), []byte(i.Salt), uint32(1), uint32(32*1024), uint8(4), uint32(32))) == i.HashedPassword {
		refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"id":  i.Id,
			"exp": time.Now().UTC().Unix() + 604800, // Current expire time is 7 days, this is subject to change
		})
		signedRefreshToken, err := refreshToken.SignedString([]byte(pepper))
		if err != nil {
			c.AbortWithStatusJSON(500, gin.H{"error": "Internal Server Error"})
			return
		} else {
			c.SetCookie("refreshToken", signedRefreshToken, 604800, "/", "", false, true)
			bearerToken, err := generateAccessToken(signedRefreshToken, "access")
			if err != nil {
				c.AbortWithStatusJSON(500, gin.H{"error": "Internal Server Error"})
				return
			}
			c.JSON(200, gin.H{"token": bearerToken})
			return
		}
	}
}

// @Summary 		Logout user
// @Description 	Logs out the current user by clearing the refresh token cookie.
// @Router 			/users/logout [post]
// @Tags 			Users
// @Success 		200 "Successfully logged out"
func logoutUser(c *gin.Context) {
	c.SetCookie("refreshToken", "", -1, "/", "", false, true)
	c.AbortWithStatus(200)
}

// @Summary 		Delete user
// @Description 	Deletes the currently authenticated user.
// @Router 			/users/delete [delete]
// @Tags 			Users
// @Security 		BearerAuth
// @Accept 			json
// @Param 			data body LoginUser true "User password for confirmation"
// @Success 		200 "User deleted successfully"
// @Failure 		400 {object} ErrorSwagger "Bad request - password mismatch"
// @Failure 		403 {object} ErrorSwagger "Forbidden"
// @Failure 		500 {object} ErrorSwagger "Internal server error"
func deleteUser(c *gin.Context) {
	userId, _ := c.Get("id")
	middlewareInput, _ := c.Get("input")
	input := middlewareInput.(LoginUser)
	var user User
	err := usersDb.FindOne(context.TODO(), bson.D{{Key: "_id", Value: userId}}).Decode(&user)
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": "Failed to delete user"})
		return
	}
	if user.Email == input.Email && base64.RawStdEncoding.EncodeToString(argon2.IDKey([]byte(input.Password+pepper), []byte(user.Salt), uint32(1), uint32(32*1024), uint8(4), uint32(32))) == user.HashedPassword && userId == user.Id {
		_, err = usersDb.DeleteOne(context.TODO(), bson.D{{Key: "_id", Value: userId}})
		if err != nil {
			c.AbortWithStatusJSON(500, gin.H{"error": "Failed to delete user"})
			return
		}
		c.SetCookie("refreshToken", "", -1, "/", "", true, true)
		c.AbortWithStatus(200)
		return
	} else if user.Email != input.Email || base64.RawStdEncoding.EncodeToString(argon2.IDKey([]byte(input.Password+pepper), []byte(user.Salt), uint32(1), uint32(32*1024), uint8(4), uint32(32))) != user.HashedPassword {
		c.AbortWithStatus(400)
		return
	} else if userId != user.Id {
		c.AbortWithStatus(403)
		return
	}
}

// @Summary 		Upload avatar for user or workspace
// @Description 	Uploads an avatar image for the current user or a specified workspace.
// @Router 			/users/upload_avatar [post]
// @Router 			/workspaces/{workspaceId}/upload_avatar [post]
// @Tags 			Users
// @Security 		BearerAuth
// @Accept 			multipart/form-data
// @Produce 		json
// @Param 			img formData file true "Avatar image file (jpg or png)"
// @Param 			workspaceId path string false "Workspace ID (if uploading for a workspace)"
// @Success 		200 "Avatar uploaded successfully"
// @Failure 		400 {object} ErrorSwagger "Bad request - no file or wrong format"
// @Failure 		404 {object} ErrorSwagger "Workspace not found"
// @Failure 		500 {object} ErrorSwagger "Internal server error"
func uploadAvatar(c *gin.Context) {
	if _, err := os.Stat("./img/"); err != nil {
		if os.IsNotExist(err) {
			err = os.Mkdir("./img/", 0777)
			if err != nil {
				print("Failed to create img/")
			}
		} else {
			print("Something went wrong when checking if img/ exists")
		}
	}
	userId, _ := c.Get("id")
	wId := c.Param("workspaceId")
	workspaceId, _ := bson.ObjectIDFromHex(wId)
	avatar, err := c.FormFile("img")
	if err != nil {
		fmt.Printf(fmt.Sprintf("%v", err))
		c.AbortWithStatusJSON(400, gin.H{"error": "No file given"})
		return
	} else if filepath.Ext(avatar.Filename) != ".png" && filepath.Ext(avatar.Filename) != ".jpg" && filepath.Ext(avatar.Filename) != ".jpeg" {
		c.AbortWithStatusJSON(400, gin.H{"error": "Wrong file format"})
		return
	}
	filename := filepath.Join("img", "/"+fmt.Sprintf("%v", uuid.New()))
	if err := c.SaveUploadedFile(avatar, filename); err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": "Internal Server Error"})
		return
	}
	if workspaceId.IsZero() {
		user := User{}
		if err := usersDb.FindOne(context.TODO(), bson.D{{"_id", userId}}).Decode(&user); err != nil {
			c.AbortWithStatusJSON(500, gin.H{"error": "Internal Server Error"})
			return
		}
		user.Avatar = filename
		_, err = usersDb.ReplaceOne(context.TODO(), bson.D{{"_id", userId}}, user)
		if err != nil {
			c.AbortWithStatusJSON(500, gin.H{"error": "Internal Server Error"})
			err = os.Remove("./img/" + filename)
			if err != nil {
				panic("Failed to remove avatar")
			}
		}
		c.AbortWithStatus(200)
	} else {
		user := Workspace{}
		if err := workspacesDb.FindOne(context.TODO(), bson.D{{"_id", workspaceId}}).Decode(&user); err != nil {
			if errors.Is(err, mongo.ErrNoDocuments) {
				c.AbortWithStatusJSON(404, gin.H{"error": "Workspace doesnt exist"})
				return
			} else {
				c.AbortWithStatusJSON(500, gin.H{"error": "Internal Server Error"})
				return
			}
		}
		user.Avatar = filename
		_, err = workspacesDb.ReplaceOne(context.TODO(), bson.D{{"_id", workspaceId}}, user)
		if err != nil {
			c.AbortWithStatusJSON(500, gin.H{"error": "Internal Server Error"})
			err = os.Remove("./img/" + filename)
			if err != nil {
				panic("Failed to remove avatar")
			}
		}
		c.AbortWithStatus(200)
	}

}

// @Summary 		Refresh bearer token
// @Description 	Generates a new access token using the refresh token stored in an http-only cookie.
// @Router 			/users/refresh [get]
// @Tags 			Users
// @Produce 		json
// @Success 		200 {object} TokenSwagger "New access token"
// @Failure 		400 {object} ErrorSwagger "Invalid token"
// @Failure 		403 {object} ErrorSwagger "Authorization required"
// @Failure 		500 {object} ErrorSwagger "Internal server error"
func refreshAccessToken(c *gin.Context) {
	refreshToken, _ := c.Cookie("refreshToken")
	token, err := jwt.ParseWithClaims(refreshToken, &Token{}, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unknown signing method: %s", token.Method)
		}
		return []byte(pepper), nil
	})
	if err != nil {
		c.AbortWithStatusJSON(500, "Internal Server Error")
		return
	}
	claims := token.Claims.(*Token)
	if claims.ExpiresAt < time.Now().UTC().Unix() {
		c.AbortWithStatusJSON(403, "Authorization Required")
		return
	} else if claims.Type == "access" {
		c.AbortWithStatusJSON(400, "Invalid Token")
	}
	bearerToken, err := generateAccessToken(refreshToken, "access")
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": "Something went wrong when generating access token"})
		return
	}
	c.AbortWithStatusJSON(200, gin.H{"token": bearerToken})
}

// @Summary 		Get user info
// @Description 	Retrieves details for the currently authenticated user.
// @Router 			/users/get_info [get]
// @Tags 			Users
// @Security 		BearerAuth
// @Produce 		json
// @Success 		200 {object} User "User details"
// @Failure 		403 {object} ErrorSwagger "Forbidden"
// @Failure 		500 {object} ErrorSwagger "Internal server error"
func getUserDetails(c *gin.Context) {
	userId, _ := c.Get("id")
	var user User
	if err := usersDb.FindOne(context.TODO(), bson.D{{"_id", userId}}).Decode(&user); err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": "Internal Server Error"})
		return
	}
	c.AbortWithStatusJSON(200, user)
}
