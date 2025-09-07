package main

import (
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type Task struct {
	Id          bson.ObjectID `json:"_id" bson:"_id,omitempty"`
	Name        string        `json:"name"`
	Description *string       `json:"description"`
	CreatedAt   int64         `json:"created_at"`
	CreatedBy   bson.ObjectID `json:"created_by"`
	CompletedAt *int64        `json:"completed_at"`
	Board       *string       `json:"board"`
}

type TaskEdit struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	CompletedAt *int64  `json:"completed_at"`
	Board       *string `json:"board"`
}
type User struct {
	Id             bson.ObjectID `json:"_id" bson:"_id,omitempty"`
	Name           string        `json:"name"`
	Email          string        `json:"email"`
	HashedPassword string        `json:"hashed_password"`
	Salt           string        `json:"salt"`
	Boards         []string      `json:"boards"`
}

type CreateUser struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginToken struct {
	Id        bson.ObjectID `json:"id"`
	ExpiresAt int64         `json:"exp"`
	jwt.RegisteredClaims
}
type LoginUser struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
type Board struct {
	Name string `json:"name"`
}
