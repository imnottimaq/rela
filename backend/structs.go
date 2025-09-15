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
	Board       bson.ObjectID `json:"board" bson:"board"`
}

type CreateTask struct {
	Name        string        `json:"name" bson:"name"`
	Description *string       `json:"description" bson:"description"`
	Board       bson.ObjectID `json:"board" bson:"board"`
}

type EditTask struct {
	Name        *string `json:"name" bson:"name"`
	Description *string `json:"description" bson:"description"`
	CompletedAt *int64  `json:"completed_at" bson:"completed_at"`
	Board       *string `json:"board" bson:"board"`
}
type User struct {
	Avatar         string        `json:"avatar" bson:"avatar"`
	Id             bson.ObjectID `json:"_id" bson:"_id,omitempty"`
	Name           string        `json:"name" bson:"name"`
	Email          string        `json:"email" bson:"email"`
	HashedPassword string        `json:"hashed_password" bson:"hashed_password"`
	Salt           string        `json:"salt" bson:"salt"`
}

type CreateUser struct {
	Name     string `json:"name" bson:"name"`
	Email    string `json:"email" bson:"email"`
	Password string `json:"password" bson:"password"`
}
type Token struct {
	Token string `json:"token" bson:"token"`
}
type LoginToken struct {
	Id        bson.ObjectID `json:"id" bson:"id"`
	ExpiresAt int64         `json:"exp" bson:"expires_at"`
	Type      string        `json:"type" bson:"type"`
	jwt.RegisteredClaims
}
type LoginUser struct {
	Email    string `json:"email" bson:"email"`
	Password string `json:"password" bson:"password"`
}
type Board struct {
	Id      bson.ObjectID `json:"_id" bson:"_id,omitempty"`
	Name    string        `json:"name" bson:"name"`
	OwnedBy bson.ObjectID `json:"owned_by" bson:"owned_by"`
}
type CreateBoard struct {
	Name string `json:"name" bson:"name"`
}

type Workspace struct {
	Id      bson.ObjectID   `bson:"_id" json:"_id"`
	Name    string          `json:"name"`
	Members []bson.ObjectID `bson:"members" json:"members"`
}
