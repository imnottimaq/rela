package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/goccy/go-json"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

var (
	lowerRegex   = regexp.MustCompile(`[a-z]`)
	upperRegex   = regexp.MustCompile(`[A-Z]`)
	digitRegex   = regexp.MustCompile(`\d`)
	specialRegex = regexp.MustCompile(`[!"#$%&'()*+,\-./:;<=>?@[\\\]^_{|}~]`)
)

func validatePassword(password string) bool {
	if len(password) < 8 {
		return false
	}

	hasLower := lowerRegex.MatchString(password)
	hasUpper := upperRegex.MatchString(password)
	hasDigit := digitRegex.MatchString(password)
	hasSpecial := specialRegex.MatchString(password)

	return hasLower && hasUpper && hasDigit && hasSpecial
}

func generateAccessToken(token string, tokenType string) (accessToken string, err error) {
	parsedToken, _ := jwt.ParseWithClaims(token, &Token{}, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unknown signing method: %s", token.Method)
		}
		return []byte(pepper), nil
	})
	claims := parsedToken.Claims.(*Token)
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

func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if strings.Contains(c.Request.URL.Path, "/api/v1/docs") {
			c.Next()
		} else {
			header := c.GetHeader("Authorization")
			if header == "" {
				c.AbortWithStatusJSON(403, gin.H{"error": "no access token"})
				return
			}
			trimmedHeader := strings.ReplaceAll(header, "Bearer ", "")
			token, err := jwt.ParseWithClaims(trimmedHeader, &Token{}, func(token *jwt.Token) (any, error) {
				if token.Method != jwt.SigningMethodHS256 {
					return nil, fmt.Errorf("unknown signing method: %s", token.Method)
				}
				return []byte(pepper), nil
			})
			if err != nil {
				c.AbortWithStatusJSON(403, gin.H{"error": "invalid or malformed access token"})
				return
			}
			claims, ok := token.Claims.(*Token)
			if !ok || !token.Valid {
				c.AbortWithStatusJSON(403, gin.H{"error": "invalid or malformed access token"})
				return
			}
			if claims.ExpiresAt < time.Now().UTC().Unix() {
				c.AbortWithStatusJSON(403, "Authorization Required")
				return
			} else if claims.Type == "refresh" || claims.Type == "invite" {
				c.AbortWithStatusJSON(400, "Invalid Token")
			} else {
				c.Set("id", claims.Id)
				c.Next()
			}
		}
	}
}

func taskMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !(c.Request.Method == "PATCH") && !(c.Request.Method == "DELETE") {
			c.Next()
			return
		}
		// input, _ := c.Get("id")
		// userId := input.(bson.ObjectID)
		taskId := c.Param("taskId")
		if taskId == "" {
			c.AbortWithStatusJSON(400, gin.H{"error": "Task id is required"})
			return
		}
		// Convert the hex taskId from URL to bson.ObjectID
		taskObjectID, err := bson.ObjectIDFromHex(taskId)
		if err != nil {
			c.AbortWithStatusJSON(400, gin.H{"error": "Invalid task id"})
			return
		}
		var output Task
		if err := tasksDb.FindOne(context.TODO(), bson.D{{"_id", taskObjectID}}).Decode(&output); err != nil {
			if errors.Is(err, mongo.ErrNoDocuments) {
				c.AbortWithStatusJSON(404, gin.H{"error": "Not Found"})
				return
			}
		} else {
			// Authorization is handled in the task handlers via authorizeTaskAccess.
			// Here we only load the task into context for downstream handlers.
			c.Set("taskObj", output)
			c.Next()
		}
	}

}

func userMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.URL.Path == "/api/v1/users/refresh" {
			c.Next()
			return
		}
		bodyBytes, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.AbortWithStatusJSON(500, gin.H{"error": "Something went wrong when parsing request"})
			return
		}
		c.Request.Body = io.NopCloser(bytes.NewReader(bodyBytes))
		switch c.Request.URL.Path {
		case "/api/v1/users/create":
			var input CreateUser
			if err := json.Unmarshal(bodyBytes, &input); err != nil {
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
			} else if input.Name == "" {
				c.AbortWithStatusJSON(400, gin.H{"error": "Name is required"})
				return
			}
			c.Set("createInput", input)
		case "/api/v1/users/login":
			var input LoginUser
			if err := json.Unmarshal(bodyBytes, &input); err != nil {
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
			c.Set("input", input)
		}
		c.Next()
	}
}
