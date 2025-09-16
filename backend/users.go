package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/goccy/go-json"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"golang.org/x/crypto/argon2"
	"log"
	"os"
	"path/filepath"
	"time"
)

// @Summary Create new user
// @Router /api/v1/users/create [post]
// @Accept json
// @Success 200 {array} TokenSwagger
// @Produce json
// @Tags Users
// @Param data body CreateUser true "Create user request"
func createUser(c *gin.Context) {
	randomBytes := make([]byte, 32)
	_, err := rand.Read(randomBytes)
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": "Internal server error"})
		return
	}
	generatedSalt := base64.URLEncoding.EncodeToString(randomBytes)
	var input CreateUser
	var i bson.M
	err = json.NewDecoder(c.Request.Body).Decode(&input)
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": "Something went wrong when parsing request"})
		return
	} else if input.Email == "" {
		c.AbortWithStatusJSON(400, gin.H{"error": "Email is required"})
		return
	} else if !emailRegex.MatchString(input.Email) {
		c.AbortWithStatusJSON(400, gin.H{"error": "Bad email"})
		return
	} else if input.Name == "" {
		c.AbortWithStatusJSON(400, gin.H{"error": "Name is required"})
		return
	} else if input.Password == "" {
		c.AbortWithStatusJSON(400, gin.H{"error": "Password is required"})
		return
	} else if !validatePassword(input.Password) {
		c.AbortWithStatusJSON(400, gin.H{"": "Your password must contain 1 uppercase letter, 1 lowercase letter, 1 special character and be at least 8 characters long"})
		return
	}
	newUser := User{
		Salt:           generatedSalt,
		Name:           input.Name,
		HashedPassword: base64.RawStdEncoding.EncodeToString(argon2.IDKey([]byte(input.Password+pepper), []byte(generatedSalt), uint32(3), uint32(128*1024), uint8(2), uint32(32))),
		Email:          input.Email,
	}
	err = usersDb.FindOne(context.TODO(), bson.D{{Key: "email", Value: input.Email}}).Decode(&i)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			user, err := usersDb.InsertOne(context.TODO(), &newUser)
			if err != nil {
				c.AbortWithStatusJSON(500, gin.H{"error": "Internal server error"})
				return
			}
			refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
				"id":  user.InsertedID.(bson.ObjectID),
				"exp": time.Now().UTC().Unix() + 604800, // Current expire time is 7 days, this is subject to change
			})
			signedRefreshToken, err := refreshToken.SignedString([]byte(pepper))
			if err != nil {
				fmt.Println(err)
				c.AbortWithStatusJSON(500, gin.H{"error": "Internal server error"})
				return
			} else {
				c.SetCookie("refreshToken", signedRefreshToken, 604800, "/", "", true, true)
				bearerToken, err := generateAccessToken(signedRefreshToken, "access")
				if err != nil {
					fmt.Println(err)
					c.AbortWithStatusJSON(500, gin.H{"error": "Internal server error"})
					return
				}
				c.AbortWithStatusJSON(200, gin.H{"token": bearerToken})
				return
			}
		}
	}
	c.AbortWithStatusJSON(400, gin.H{"error": "Bad request"})
	return
}

// @Summary Login user
// @Router /api/v1/users/login [post]
// @Accept json
// @Success 200 {array} TokenSwagger
// @Produce json
// @Tags Users
// @Param data body LoginUser true "Login user request"
func loginUser(c *gin.Context) {
	var input LoginUser
	var i User
	err := json.NewDecoder(c.Request.Body).Decode(&input)
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
	}
	if err = usersDb.FindOne(context.TODO(), bson.D{{Key: "email", Value: input.Email}}).Decode(&i); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			c.AbortWithStatusJSON(404, gin.H{"error": "User doesnt exist"})
			return
		} else {
			c.AbortWithStatusJSON(500, gin.H{"error": "Internal server error"})
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
			c.AbortWithStatusJSON(500, gin.H{"error": "Internal server error"})
			return
		} else {
			c.SetCookie("refreshToken", signedRefreshToken, 604800, "/", "", true, true)
			bearerToken, err := generateAccessToken(signedRefreshToken, "access")
			if err != nil {
				c.AbortWithStatusJSON(500, gin.H{"error": "Internal server error"})
				return
			}
			c.JSON(200, gin.H{"token": bearerToken})
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
	err := json.NewDecoder(c.Request.Body).Decode(&input)
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
	} else if !validatePassword(input.Password) {
		c.AbortWithStatusJSON(400, gin.H{"error": "Password does not meet requirements"})
		return
	}
	err = usersDb.FindOne(context.TODO(), bson.D{{Key: "_id", Value: userId}}).Decode(&user)
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": "Failed to delete user"})
		return
	}
	if user.Email == input.Email && base64.RawStdEncoding.EncodeToString(argon2.IDKey([]byte(input.Password+pepper), []byte(user.Salt), uint32(3), uint32(128*1024), uint8(2), uint32(32))) == user.HashedPassword && userId == user.Id {
		_, err = usersDb.DeleteOne(context.TODO(), bson.D{{Key: "_id", Value: userId}})
		if err != nil {
			c.AbortWithStatusJSON(500, gin.H{"error": "Failed to delete user"})
			return
		}
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

// @Summary Upload avatar for user
// @Router /api/v1/users/upload_avatar [post]
// @Accept mpfd
// @Success 200
// @Tags Users
// @Param image formData string true "Avatar"
// @Param X-Authorization header string true "Bearer Token"
func uploadAvatar(c *gin.Context) {
	if _, err := os.Stat("./img/"); err != nil {
		if os.IsNotExist(err) {
			err = os.Mkdir("./img/", 0777)
			if err != nil {
				log.Fatal("Failed to create img/")
			}
		} else {
			log.Fatal("Something went wrong when checking if img/ exists")
		}
	}
	userId, _ := c.Get("id")
	user := User{}
	avatar, err := c.FormFile("img")
	if err != nil {
		fmt.Printf(fmt.Sprintf("%v", err))
		c.AbortWithStatusJSON(400, gin.H{"error": "No file given"})
		return
	} else if filepath.Ext(avatar.Filename) != ".png" && filepath.Ext(avatar.Filename) != ".jpg" && filepath.Ext(avatar.Filename) != ".jpeg" {
		c.AbortWithStatusJSON(400, gin.H{"error": "Wrong file format"})
		return
	}
	err = usersDb.FindOne(context.TODO(), bson.D{{"_id", userId}}).Decode(&user)
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": "Failed to find user"})
		return
	}
	filename := filepath.Join("img", "/"+fmt.Sprintf("%v", uuid.New()))
	if err := c.SaveUploadedFile(avatar, filename); err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": "Error while saving image"})
		return
	} else {
		user.Avatar = filename
		_, err = usersDb.ReplaceOne(context.TODO(), bson.D{{"_id", userId}}, user)
		if err != nil {
			c.AbortWithStatusJSON(500, gin.H{"error": "Failed to pin image to user"})
			err = os.Remove("./img/" + filename)
			if err != nil {
				log.Fatal("Failed to remove avatar")
			}
		}

		c.AbortWithStatus(200)
	}

}

// @Summary Refresh bearer token
// @Description For this route, you must have refresh token, that is sent to your browser when you log into user account as an http-only cookie
// @Router /api/v1/users/refresh [get]
// @Success 200 {array} TokenSwagger
// @Produce json
// @Tags Users

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

// @Summary Get user info
// @Description For this route, you must have bearer token
// @Router /api/v1/users/get_info [get]
// @Success 200 {array} User
// @Produce json
// @Tags Users
// @Param X-Authorization header string true "Bearer Token"
func getUserDetails(c *gin.Context) {
	userId, _ := c.Get("id")
	var user User
	usersDb.FindOne(context.TODO(), bson.D{{"_id", userId}}).Decode(&user)
	c.AbortWithStatusJSON(200, user)
}
