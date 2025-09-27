package main

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/goccy/go-json"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// @Summary 		Create a new workspace
// @Description 	Creates a new workspace for the current user.
// @Router 			/workspaces/create [post]
// @Tags 			Workspaces
// @Security 		BearerAuth
// @Accept 			json
// @Produce 		json
// @Param 			data body CreateWorkspace true "Workspace creation data"
// @Success 		200 {object} Workspace "The created workspace"
// @Failure 		400 {object} ErrorSwagger "Bad request - no name specified"
// @Failure 		500 {object} ErrorSwagger "Internal server error"
func createWorkspace(c *gin.Context) {
	id, _ := c.Get("id")
	userId := id.(bson.ObjectID)
	var input CreateWorkspace
	if err := json.NewDecoder(c.Request.Body).Decode(&input); err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": "Something went wrong when parsing request"})
		return
	} else if input.Name == "" {
		c.AbortWithStatusJSON(400, gin.H{"error": "Field 'name' is not specified"})
		return
	}
	output := Workspace{
		Name:    input.Name,
		OwnedBy: userId,
		Members: []bson.ObjectID{userId},
	}
	workspace, err := workspacesDb.InsertOne(context.TODO(), output)
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": "Internal Server Error"})
		return
	}
	output.Id = workspace.InsertedID.(bson.ObjectID)
	c.JSON(200, output)
}

// @Summary 		Edit a workspace
// @Description 	Edits the details of a specific workspace.
// @Router 			/workspaces/{workspaceId} [patch]
// @Tags 			Workspaces
// @Security 		BearerAuth
// @Accept 			json
// @Produce 		json
// @Param 			workspaceId path string true "Workspace ID"
// @Param 			data body EditWorkspace true "Fields to edit in the workspace"
// @Success 		200 {object} Workspace "The updated workspace"
// @Failure 		400 {object} ErrorSwagger "Bad request - invalid input"
// @Failure 		403 {object} ErrorSwagger "Forbidden - you are not the owner of this workspace"
// @Failure 		404 {object} ErrorSwagger "Not Found - workspace not found"
// @Failure 		500 {object} ErrorSwagger "Internal server error"
func editWorkspace(c *gin.Context) {
	userId, _ := c.Get("id")
	wId := c.Param("workspaceId")
	workspaceId, _ := bson.ObjectIDFromHex(wId)
	var previousVersion Workspace
	var input EditWorkspace
	if err := json.NewDecoder(c.Request.Body).Decode(&input); err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": "Failed to parse request"})
		return
	}
	if err := workspacesDb.FindOne(context.TODO(), bson.D{{"_id", workspaceId}}).Decode(&previousVersion); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			c.AbortWithStatusJSON(404, gin.H{"error": "Not Found"})
		} else {
			c.AbortWithStatusJSON(500, gin.H{"error": "Internal Server Error"})
		}
		return
	}
	if previousVersion.OwnedBy != userId.(bson.ObjectID) {
		c.AbortWithStatusJSON(403, gin.H{"error": "You are not an owner of this workplace"})
		return
	}
	if input.Name != "" {
		previousVersion.Name = input.Name
	}
	// Avatar can be empty string to remove it
	previousVersion.Avatar = input.Avatar

	if _, err := workspacesDb.ReplaceOne(context.TODO(), bson.D{{"_id", workspaceId}}, previousVersion); err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": "Internal Server Error"})
		return
	}
	c.JSON(200, previousVersion)
}

// @Summary 		Get all workspaces for the current user
// @Description 	Returns a list of all workspaces the current user is a member of.
// @Router 			/users/workspaces [get]
// @Tags 			Workspaces
// @Security 		BearerAuth
// @Produce 		json
// @Success 		200 {object} AllWorkspacesResponse "A list of workspaces"
// @Failure 		500 {object} ErrorSwagger "Internal Server Error"
func getAllWorkspaces(c *gin.Context) {
	id, _ := c.Get("id")
	userId := id.(bson.ObjectID)
	cursor, _ := workspacesDb.Find(context.TODO(), bson.D{{"members", userId}})
	workspaces := make([]Workspace, 0)
	if err := cursor.All(context.TODO(), &workspaces); err != nil {
		c.JSON(500, gin.H{"error": "failed to decode workspaces"})
		return
	}
	if err := cursor.Close(context.TODO()); err != nil {
		// Log the error but don't abort, as the response has already been sent.
		println("Failed to close cursor: ", err.Error())
	}
	c.IndentedJSON(200, gin.H{"workspaces": workspaces})
}

// @Summary 		Delete a workspace
// @Description 	Deletes a specific workspace.
// @Router 			/workspaces/{workspaceId} [delete]
// @Tags 			Workspaces
// @Security 		BearerAuth
// @Produce 		json
// @Param 			workspaceId path string true "Workspace ID"
// @Success 		200 "Workspace deleted successfully"
// @Failure 		403 {object} ErrorSwagger "Forbidden - you are not the owner of this workspace"
// @Failure 		404 {object} ErrorSwagger "Not Found - workspace not found"
// @Failure 		500 {object} ErrorSwagger "Internal server error"
func deleteWorkspace(c *gin.Context) {
	id, _ := c.Get("id")
	workspaceIdStr := c.Param("workspaceId")
	workspaceId, _ := bson.ObjectIDFromHex(workspaceIdStr)
	var previousVersion Workspace
	if err := workspacesDb.FindOne(context.TODO(), bson.D{{"_id", workspaceId}}).Decode(&previousVersion); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			c.AbortWithStatusJSON(404, gin.H{"error": "Not Found"})
		} else {
			c.AbortWithStatusJSON(500, gin.H{"error": "Internal Server Error"})
		}
		return
	}
	if previousVersion.OwnedBy != id.(bson.ObjectID) {
		c.AbortWithStatusJSON(403, gin.H{"error": "You are not an owner of this workplace"})
		return
	}
	if _, err := workspacesDb.DeleteOne(context.TODO(), bson.D{{"_id", workspaceId}}); err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": "Internal Server Error"})
		return
	}
	c.AbortWithStatus(200)
}

// @Summary 		Join a workspace
// @Description 	Adds the current user to a workspace using an invite token.
// @Router 			/workspaces/add/{joinToken} [post]
// @Tags 			Workspaces
// @Security 		BearerAuth
// @Produce 		json
// @Param 			joinToken path string true "Join Token"
// @Success 		200 "Successfully joined the workspace"
// @Failure 		400 {object} ErrorSwagger "Bad request - invalid token"
// @Failure 		404 {object} ErrorSwagger "Not Found - workspace not found"
// @Failure 		500 {object} ErrorSwagger "Internal server error"
func addMember(c *gin.Context) {
	userId, _ := c.Get("id")
	joinToken := c.Param("joinToken")
	var workspace Workspace
	token, err := jwt.ParseWithClaims(joinToken, &Token{}, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unknown signing method: %s", token.Method)
		}
		return []byte(pepper), nil
	})
	if err != nil || !token.Valid {
		c.AbortWithStatusJSON(400, gin.H{"error": "Invalid or expired token"})
		return
	}
	claims, ok := token.Claims.(*Token)
	if !ok {
		c.AbortWithStatusJSON(400, gin.H{"error": "Invalid token claims"})
		return
	}
	if claims.Type != "invite" {
		c.AbortWithStatusJSON(400, gin.H{"error": "Invalid token type"})
		return
	}
	if err := workspacesDb.FindOne(context.TODO(), bson.D{{"_id", claims.Id}}).Decode(&workspace); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			c.AbortWithStatusJSON(404, gin.H{"error": "Workspace does not exist"})
		} else {
			c.AbortWithStatusJSON(500, gin.H{"error": "Internal Server Error"})
		}
		return
	}
	// Add user to members if not already present
	if slices.Index(workspace.Members, userId.(bson.ObjectID)) == -1 {
		workspace.Members = append(workspace.Members, userId.(bson.ObjectID))
		if _, err := workspacesDb.ReplaceOne(context.TODO(), bson.D{{"_id", claims.Id}}, workspace); err != nil {
			c.AbortWithStatusJSON(500, gin.H{"error": "Internal Server Error"})
			return
		}
	}
	c.AbortWithStatus(200)
}

// @Summary 		Create a workspace invite
// @Description 	Generates a new invite token for a workspace.
// @Router 			/workspaces/{workspaceId}/new_invite [get]
// @Tags 			Workspaces
// @Security 		BearerAuth
// @Produce 		json
// @Param 			workspaceId path string true "Workspace ID"
// @Success 		200 {object} TokenSwagger "The invite token"
// @Failure 		400 {object} ErrorSwagger "Bad request - workspace ID not specified"
// @Failure 		403 {object} ErrorSwagger "Forbidden - you are not the owner of this workspace"
// @Failure 		404 {object} ErrorSwagger "Not Found - workspace not found"
func createNewInvite(c *gin.Context) {
	userId, _ := c.Get("id")
	workspaceIdStr := c.Param("workspaceId")
	if workspaceIdStr == "" {
		c.AbortWithStatusJSON(400, gin.H{"error": "You must specify workspace id"})
		return
	}
	var workspace Workspace
	workspaceId, _ := bson.ObjectIDFromHex(workspaceIdStr)
	if err := workspacesDb.FindOne(context.TODO(), bson.D{{"_id", workspaceId}}).Decode(&workspace); err != nil {
		c.AbortWithStatusJSON(404, gin.H{"error": "Not Found"})
		return
	}
	if workspace.OwnedBy != userId.(bson.ObjectID) {
		c.AbortWithStatusJSON(403, gin.H{"error": "You are not an owner of this workplace"})
		return
	}
	claims := Token{
		Id:   workspaceId,
		Type: "invite",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(2 * 7 * 24 * time.Hour)), // 2 weeks
		},
	}
	newToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	inviteToken, _ := newToken.SignedString([]byte(pepper))
	c.JSON(200, gin.H{"token": inviteToken})
}

// @Summary 		Get workspace details from invite token
// @Description 	Retrieves workspace details using an invite token.
// @Router 			/workspaces/invite/{joinToken} [get]
// @Tags 			Workspaces
// @Produce 		json
// @Param 			joinToken path string true "Join Token"
// @Success 		200 {object} Workspace "The workspace details"
// @Failure 		400 {object} ErrorSwagger "Bad request - invalid token"
// @Failure 		404 {object} ErrorSwagger "Not Found - workspace not found"
// @Failure 		500 {object} ErrorSwagger "Internal server error"
func getWorkspaceByInviteToken(c *gin.Context) {
	joinTokenStr := c.Param("joinToken")

	// 1. Parse and validate the JWT
	token, err := jwt.ParseWithClaims(joinTokenStr, &Token{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(pepper), nil
	})

	if err != nil || !token.Valid {
		c.AbortWithStatusJSON(400, gin.H{"error": "Invalid or expired token"})
		return
	}

	// 2. Extract claims
	claims, ok := token.Claims.(*Token)
	if !ok {
		c.AbortWithStatusJSON(500, gin.H{"error": "Failed to process token claims"})
		return
	}

	// 3. Check token type
	if claims.Type != "invite" {
		c.AbortWithStatusJSON(400, gin.H{"error": "Invalid token type"})
		return
	}

	// 5. Fetch workspace from DB
	var workspace Workspace
	if err := workspacesDb.FindOne(context.TODO(), bson.D{{"_id", claims.Id}}).Decode(&workspace); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			c.AbortWithStatusJSON(404, gin.H{"error": "Workspace does not exist"})
		} else {
			c.AbortWithStatusJSON(500, gin.H{"error": "Internal Server Error"})
		}
		return
	}

	// 6. Return workspace details
	c.JSON(200, workspace)
}

// @Summary 		Kick a member from a workspace
// @Description 	Removes a member from a specific workspace.
// @Router 			/workspaces/{workspaceId}/kick [delete]
// @Tags 			Workspaces
// @Security 		BearerAuth
// @Accept 			json
// @Param 			workspaceId path string true "Workspace ID"
// @Param 			data body KickUser true "User ID to kick"
// @Success 		200 "Member kicked successfully"
// @Failure 		400 {object} ErrorSwagger "Bad request - workspace ID not specified"
// @Failure 		403 {object} ErrorSwagger "Forbidden - you are not the owner of this workspace"
// @Failure 		404 {object} ErrorSwagger "Not Found - workspace or user not found"
// @Failure 		500 {object} ErrorSwagger "Internal server error"
func kickMember(c *gin.Context) {
	currentUserId, _ := c.Get("id")
	workspaceIdStr := c.Param("workspaceId")
	if workspaceIdStr == "" {
		c.AbortWithStatusJSON(400, gin.H{"error": "You must specify workspace id"})
		return
	}
	var input KickUser
	if err := json.NewDecoder(c.Request.Body).Decode(&input); err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": "Something went wrong when parsing request"})
		return
	}
	workspaceId, _ := bson.ObjectIDFromHex(workspaceIdStr)
	var workspace Workspace
	if err := workspacesDb.FindOne(context.TODO(), bson.D{{"_id", workspaceId}}).Decode(&workspace); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			c.AbortWithStatusJSON(404, gin.H{"error": "Not Found"})
		} else {
			c.AbortWithStatusJSON(500, gin.H{"error": "Internal server error"})
		}
		return
	}
	if workspace.OwnedBy != currentUserId.(bson.ObjectID) {
		c.AbortWithStatusJSON(403, gin.H{"error": "You are not owner of this workspace"})
		return
	}
	// Prevent owner from being kicked
	if input.Id == workspace.OwnedBy {
		c.AbortWithStatusJSON(400, gin.H{"error": "Cannot kick the owner"})
		return
	}
	// Remove member
	newMembers := make([]bson.ObjectID, 0)
	found := false
	for _, memberId := range workspace.Members {
		if memberId == input.Id {
			found = true
			continue
		}
		newMembers = append(newMembers, memberId)
	}
	if !found {
		c.AbortWithStatusJSON(404, gin.H{"error": "User not found in workspace"})
		return
	}
	workspace.Members = newMembers
	if _, err := workspacesDb.ReplaceOne(context.TODO(), bson.D{{"_id", workspace.Id}}, workspace); err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": "Internal server error"})
		return
	}
	c.AbortWithStatus(200)
}

// @Summary 		Promote a member to owner
// @Description 	Promotes a member of a workspace to be the new owner.
// @Router 			/workspaces/{workspaceId}/promote/{userId} [patch]
// @Tags 			Workspaces
// @Security 		BearerAuth
// @Param 			workspaceId path string true "Workspace ID"
// @Param 			userId path string true "User ID to promote"
// @Success 		200 "Member promoted successfully"
// @Failure 		400 {object} ErrorSwagger "Bad request - user is not part of this workspace"
// @Failure 		403 {object} ErrorSwagger "Forbidden - you are not the owner of this workspace"
// @Failure 		404 {object} ErrorSwagger "Not Found - workspace or user not found"
// @Failure 		500 {object} ErrorSwagger "Internal server error"
func promoteMember(c *gin.Context) {
	currentUserId, _ := c.Get("id")
	workspaceIdStr := c.Param("workspaceId")
	userIdToPromoteStr := c.Param("userId")
	workspaceId, _ := bson.ObjectIDFromHex(workspaceIdStr)
	userIdToPromote, _ := bson.ObjectIDFromHex(userIdToPromoteStr)

	var workspace Workspace
	if err := workspacesDb.FindOne(context.TODO(), bson.D{{"_id", workspaceId}}).Decode(&workspace); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			c.AbortWithStatusJSON(404, gin.H{"error": "Not Found"})
		} else {
			c.AbortWithStatusJSON(500, gin.H{"error": "Internal server error"})
		}
		return
	}
	if workspace.OwnedBy != currentUserId.(bson.ObjectID) {
		c.AbortWithStatusJSON(403, gin.H{"error": "You are not owner of this workspace"})
		return
	}
	if slices.Index(workspace.Members, userIdToPromote) == -1 {
		c.AbortWithStatusJSON(400, gin.H{"error": "User is not part of this workspace"})
		return
	}
	workspace.OwnedBy = userIdToPromote
	if _, err := workspacesDb.ReplaceOne(context.TODO(), bson.D{{"_id", workspace.Id}}, workspace); err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": "Internal server error"})
		return
	}
	c.AbortWithStatus(200)
}

// @Summary 		Get all members of a workspace
// @Description 	Returns a list of all members in a specific workspace.
// @Router 			/workspaces/{workspaceId}/members [get]
// @Tags 			Workspaces
// @Security 		BearerAuth
// @Produce 		json
// @Param 			workspaceId path string true "Workspace ID"
// @Success 		200 {object} AllMembersResponse "A list of members"
// @Failure 		400 {object} ErrorSwagger "Bad request - invalid workspace ID"
// @Failure 		403 {object} ErrorSwagger "Forbidden - you are not a member of this workspace"
// @Failure 		404 {object} ErrorSwagger "Not Found - workspace not found"
// @Failure 		500 {object} ErrorSwagger "Internal server error"
func getAllMembers(c *gin.Context) {
	userId, _ := c.Get("id")
	workspaceIdStr := c.Param("workspaceId")
	workspaceId, err := bson.ObjectIDFromHex(workspaceIdStr)
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid workspaceId"})
		return
	}

	var workspace Workspace
	if err := workspacesDb.FindOne(context.TODO(), bson.M{"_id": workspaceId}).Decode(&workspace); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			c.JSON(404, gin.H{"error": "workspace not found"})
		} else {
			c.JSON(500, gin.H{"error": err.Error()})
		}
		return
	}

	if slices.Index(workspace.Members, userId.(bson.ObjectID)) == -1 {
		c.JSON(403, gin.H{"error": "not a member of this workspace"})
		return
	}
	cursor, err := usersDb.Find(context.TODO(), bson.M{
		"_id": bson.M{"$in": workspace.Members},
	})
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to fetch users"})
		return
	}
	defer cursor.Close(context.TODO())

	members := make([]Member, 0)
	if err := cursor.All(context.TODO(), &members); err != nil {
		c.JSON(500, gin.H{"error": "failed to decode users"})
		return
	}
	c.IndentedJSON(200, gin.H{"members": members})
}

// @Summary 		Get a workspace by ID
// @Description 	Retrieves a specific workspace by its ID.
// @Router 			/workspaces/{workspaceId} [get]
// @Tags 			Workspaces
// @Security 		BearerAuth
// @Produce 		json
// @Param 			workspaceId path string true "Workspace ID"
// @Success 		200 {object} Workspace "The requested workspace"
// @Failure 		400 {object} ErrorSwagger "Bad request - invalid workspace ID"
// @Failure 		403 {object} ErrorSwagger "Forbidden - you are not a member of this workspace"
// @Failure 		404 {object} ErrorSwagger "Not Found - workspace not found"
// @Failure 		500 {object} ErrorSwagger "Internal server error"
func getWorkspace(c *gin.Context) {
	id, _ := c.Get("id")
	userId := id.(bson.ObjectID)
	workspaceIdStr := c.Param("workspaceId")
	workspaceId, err := bson.ObjectIDFromHex(workspaceIdStr)
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid workspaceId"})
		return
	}

	var workspace Workspace
	if err := workspacesDb.FindOne(context.TODO(), bson.M{"_id": workspaceId}).Decode(&workspace); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			c.JSON(404, gin.H{"error": "workspace not found"})
		} else {
			c.JSON(500, gin.H{"error": err.Error()})
		}
		return
	}

	// Ensure requester is a member or owner
	if slices.Index(workspace.Members, userId) == -1 && workspace.OwnedBy != userId {
		c.JSON(403, gin.H{"error": "forbidden"})
		return
	}

	c.IndentedJSON(200, workspace)
}

// @Summary 		Get workspace info
// @Description 	Retrieves detailed information about a workspace, including members and boards.
// @Router 			/workspaces/{workspaceId}/info [get]
// @Tags 			Workspaces
// @Security 		BearerAuth
// @Produce 		json
// @Param 			workspaceId path string true "Workspace ID"
// @Success 		200 {object} WorkspaceInfo "Detailed workspace information"
// @Failure 		400 {object} ErrorSwagger "Bad request - invalid workspace ID"
// @Failure 		403 {object} ErrorSwagger "Forbidden - you are not a member of this workspace"
// @Failure 		404 {object} ErrorSwagger "Not Found - workspace not found"
// @Failure 		500 {object} ErrorSwagger "Internal server error"
func getWorkspaceInfo(c *gin.Context) {
	id, _ := c.Get("id")
	userId := id.(bson.ObjectID)

	workspaceIdStr := c.Param("workspaceId")
	workspaceId, err := bson.ObjectIDFromHex(workspaceIdStr)
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid workspaceId"})
		return
	}

	pipeline := mongo.Pipeline{
		{{"$match", bson.D{{"_id", workspaceId}}}},
		{{"$lookup", bson.D{
			{"from", "users"},
			{"localField", "members"},
			{"foreignField", "_id"},
			{"as", "memberDetails"},
			{"pipeline", bson.A{
				bson.D{{"$project", bson.D{
					{"name", 1},
					{"avatar", 1},
				}}},
			}},
		}}},
		{{"$lookup", bson.D{
			{"from", "boards"},
			{"localField", "_id"},
			{"foreignField", "owned_by"},
			{"as", "boards"},
		}}},
	}

	cursor, err := workspacesDb.Aggregate(context.TODO(), pipeline)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to fetch workspace info"})
		return
	}
	defer cursor.Close(context.TODO())

	var results []WorkspaceInfo
	if err := cursor.All(context.TODO(), &results); err != nil {
		c.JSON(500, gin.H{"error": "failed to decode results"})
		return
	}

	if len(results) == 0 {
		c.JSON(404, gin.H{"error": "workspace not found"})
		return
	}

	result := results[0]

	authorized := result.OwnedBy == userId
	if !authorized {
		for _, member := range result.Members {
			if member == userId {
				authorized = true
				break
			}
		}
	}

	if !authorized {
		c.JSON(403, gin.H{"error": "forbidden"})
		return
	}

	c.JSON(200, result)
}

// @Summary 		Assign a task to a user
// @Description 	Assigns a task within a workspace to a specific user.
// @Router 			/workspaces/{workspaceId}/assign [post]
// @Tags 			Tasks
// @Security 		BearerAuth
// @Accept 			json
// @Produce 		json
// @Param 			workspaceId path string true "Workspace ID"
// @Param 			data body AssignTask true "Task and User IDs"
// @Success 		200 {object} Task "The updated task"
// @Failure 		403 {object} ErrorSwagger "Forbidden - you are not the owner or the user to be assigned"
// @Failure 		404 {object} ErrorSwagger "Not Found - workspace or task not found"
// @Failure 		500 {object} ErrorSwagger "Internal server error"
func assignTask(c *gin.Context) {
	currentUserId, _ := c.Get("id")
	wId := c.Param("workspaceId")
	workspaceId, _ := bson.ObjectIDFromHex(wId)
	var input AssignTask
	if err := json.NewDecoder(c.Request.Body).Decode(&input); err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": "Something went wrong when parsing request"})
		return
	}
	var workspace Workspace
	if err := workspacesDb.FindOne(context.TODO(), bson.D{{"_id", workspaceId}}).Decode(&workspace); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			c.AbortWithStatusJSON(404, gin.H{"error": "Workspace does not exist"})
		} else {
			c.AbortWithStatusJSON(500, gin.H{"error": "Internal Server Error"})
		}
		return
	}

	// Allow assignment if the current user is the workspace owner or the user being assigned the task.
	// A more robust implementation might check if the assigned user is a member of the workspace.
	if input.UserId == currentUserId.(bson.ObjectID) || workspace.OwnedBy == currentUserId.(bson.ObjectID) {
		var task Task
		if err := tasksDb.FindOne(context.TODO(), bson.D{{"_id", input.TaskId}}).Decode(&task); err != nil {
			if errors.Is(err, mongo.ErrNoDocuments) {
				c.AbortWithStatusJSON(404, gin.H{"error": "Task does not exist"})
			} else {
				c.AbortWithStatusJSON(500, gin.H{"error": "Internal Server Error"})
			}
			return
		}
		task.AssignedTo = input.UserId
		if _, err := tasksDb.ReplaceOne(context.TODO(), bson.D{{"_id", input.TaskId}}, task); err != nil {
			c.AbortWithStatusJSON(500, gin.H{"error": "Internal Server Error"})
			return
		}
		c.JSON(200, task)
	} else {
		c.AbortWithStatusJSON(403, gin.H{"error": "You are not authorized to assign this task"})
	}
}
