package main

import (
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type Task struct {
	Id          bson.ObjectID `json:"_id" bson:"_id,omitempty"`
	Name        string        `json:"name" bson:"name"`
	Description *string       `json:"description" bson:"description"`
	CreatedAt   int64         `json:"created_at" bson:"created_at"`
	OwnedBy     bson.ObjectID `json:"created_by" bson:"created_by"`
	CompletedAt *int64        `json:"completed_at" bson:"completed_at"`
	Board       bson.ObjectID `json:"board" bson:"board"`
}

type TaskEdit struct {
	Name        *string `json:"name" bson:"name"`
	Description *string `json:"description" bson:"description"`
	CompletedAt *int64  `json:"completed_at" bson:"completed_at"`
	Board       *string `json:"board" bson:"board"`
}
type User struct {
	Id             bson.ObjectID `json:"_id" bson:"_id,omitempty"`
	Name           string        `json:"name" bson:"name"`
	Email          string        `json:"email" bson:"email"`
	HashedPassword string        `json:"hashed_password" bson:"hashed_password"`
	Salt           string        `json:"salt" bson:"salt"`
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
	Name    string        `json:"name"`
	OwnedBy bson.ObjectID `json:"owned_by"`
}
