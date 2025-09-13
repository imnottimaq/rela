package main

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("X-Authorization")
		if header == "" {
			c.AbortWithStatusJSON(403, gin.H{"error": "no access token"})
			return
		}
		token, err := jwt.ParseWithClaims(header, &LoginToken{}, func(token *jwt.Token) (any, error) {
			if token.Method != jwt.SigningMethodHS256 {
				return nil, fmt.Errorf("unknown signing method: %s", token.Method)
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
		} else if claims.Type == "refresh" {
			c.AbortWithStatusJSON(400, "Invalid Token")
		} else {
			c.Set("id", claims.Id)
			c.Next()
		}
	}
}
