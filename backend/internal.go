package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/goccy/go-json"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"io"
	"regexp"
	"strings"
	"time"
)

func validatePassword(password string) bool {
	if len(password) < 8 {
		return false
	}

	hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
	hasDigit := regexp.MustCompile(`\d`).MatchString(password)
	hasSpecial := regexp.MustCompile(`[!"#$%&'()*+,\-./:;<=>?@[\\\]^_{|}~]`).MatchString(password)

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
		if c.Request.Method == "PATCH" || c.Request.Method == "DELETE" {
			c.Next()
		} else {
			input, _ := c.Get("id")
			userId := input.(bson.ObjectID)
			if taskId := c.Param("taskId"); taskId == "" {
				print(taskId)
				c.AbortWithStatusJSON(400, gin.H{"error": "Task id is required"})
				return
			} else {
				var output Task
				if err := tasksDb.FindOne(context.TODO(), bson.D{{"_id", taskId}}).Decode(&output); err != nil {
					if errors.Is(err, mongo.ErrNoDocuments) {
						c.AbortWithStatusJSON(404, gin.H{"error": "Not Found"})
						return
					}
				} else {
					if userId != output.CreatedBy {
						c.AbortWithStatusJSON(403, gin.H{"error": "You do not own this task"})
						return
					} else {
						c.Set("taskObj", output)
						c.Next()
					}
				}
			}
		}
	}

}

func userMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.URL.Path == "/api/v1/users/refresh" {
			c.Next()
		} else {
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
				fallthrough
			default:
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
				} else if !validatePassword(input.Password) {
					c.AbortWithStatusJSON(400, gin.H{"error": "Password does not meet requirements"})
					return
				}
				c.Set("input", input)
			}
			c.Next()
		}
	}
}
