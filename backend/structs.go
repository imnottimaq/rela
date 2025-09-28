package main

import (
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type Task struct {
	Id          bson.ObjectID `json:"_id" bson:"_id,omitempty"`
	Name        string        `json:"name" bson:"name"`
	Description string        `json:"description" bson:"description"`
	CreatedAt   int64         `json:"created_at" bson:"created_at"`
	CreatedBy   bson.ObjectID `json:"created_by" bson:"created_by"`
	Board       bson.ObjectID `json:"board" bson:"board"`
	AssignedTo  bson.ObjectID `json:"assigned_to" bson:"assigned_to,omitempty"`
	Deadline    int64         `json:"deadline" bson:"deadline"`
}

type AllTasksResponse struct {
	Tasks []Task `json:"tasks"`
}

type CreateTask struct {
	Name        string        `json:"name" bson:"name"`
	Description string        `json:"description" bson:"description"`
	Board       bson.ObjectID `json:"board" bson:"board"`
}

type EditTask struct {
	Name        string `json:"name" bson:"name"`
	Description string `json:"description" bson:"description"`
	CompletedAt int64  `json:"completed_at" bson:"completed_at"`
	Board       string `json:"board" bson:"board"`
	Deadline    int64  `json:"deadline" bson:"deadline"`
}

type User struct {
	Avatar         string        `json:"avatar" bson:"avatar"`
	Id             bson.ObjectID `json:"_id" bson:"_id,omitempty"`
	Name           string        `json:"name" bson:"name"`
	Email          string        `json:"email" bson:"email"`
	HashedPassword string        `json:"-" bson:"hashed_password"`
	Salt           string        `json:"-" bson:"salt"`
}

type CreateUser struct {
	Name     string `json:"name" bson:"name"`
	Email    string `json:"email" bson:"email"`
	Password string `json:"password" bson:"password"`
}

type TokenSwagger struct {
	Token string `json:"token" bson:"token"`
}

type ErrorSwagger struct {
	Error string `json:"error"`
}

type Token struct {
	Id   bson.ObjectID `json:"id" bson:"id"`
	Type string        `json:"type" bson:"type"`
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

type AllBoardsResponse struct {
	Boards []Board `json:"boards"`
}

type CreateBoard struct {
	Name string `json:"name" bson:"name"`
}

type Workspace struct {
	Id      bson.ObjectID   `bson:"_id,omitempty" json:"_id"`
	Avatar  string          `json:"avatar" bson:"avatar"`
	Name    string          `json:"name"`
	OwnedBy bson.ObjectID   `bson:"owned_by" json:"owned_by"`
	Members []bson.ObjectID `bson:"members" json:"members"`
}

type CreateWorkspace struct {
	Name string `json:"name"`
}

type EditWorkspace struct {
	Name   string `json:"name"`
	Avatar string `json:"avatar"`
}

type AllWorkspacesResponse struct {
	Workspaces []Workspace `json:"workspaces"`
}

type AllMembersResponse struct {
	Members []Member `json:"members"`
}

type WorkspaceInfo struct {
	Id            bson.ObjectID   `bson:"_id,omitempty" json:"_id"`
	Avatar        string          `json:"avatar" bson:"avatar"`
	Name          string          `json:"name"`
	OwnedBy       bson.ObjectID   `bson:"owned_by" json:"owned_by"`
	Members       []bson.ObjectID `bson:"members" json:"-"`
	MemberDetails []Member        `bson:"memberDetails" json:"memberDetails"`
	Boards        []Board         `bson:"boards" json:"boards"`
}

type KickUser struct {
	Id bson.ObjectID `bson:"id" json:"id"`
}

type Member struct {
	Id     bson.ObjectID `bson:"_id" json:"_id"`
	Name   string        `json:"name" bson:"name"`
	Avatar string        `json:"avatar" bson:"avatar"`
}

type AssignTask struct {
	TaskId bson.ObjectID `bson:"taskId" json:"taskId"`
	UserId bson.ObjectID `bson:"userId" json:"userId"`
}
