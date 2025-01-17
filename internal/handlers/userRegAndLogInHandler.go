package handlers

import (
	"context"
	"net/http"

	"gw-currncy-wallet/internal/auth"
	postgres "gw-currncy-wallet/internal/storages/postgres"
	"gw-currncy-wallet/pkg/pswcrypt"

	"github.com/gin-gonic/gin"
)

type UserStruct struct {
	db *postgres.StorageConn
}

func NewUserService(db *postgres.StorageConn) *UserStruct {
	return &UserStruct{db: db}
}

func (u *UserStruct) RegisterHandler(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required,min=6"`
	}
	// Check input data
	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input", "details": err.Error()})
		return
	}
	hashedPasswrd, err := pswcrypt.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	ctx := context.Background()
	userID, err := u.db.CreateUser(ctx, req.Username, req.Email, string(hashedPasswrd))
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "User registered successfully", "user_id": userID})
}

func (u *UserStruct) LoginHandler(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input", "details": err.Error()})
		return
	}

	ctx := context.Background()

	user, err := u.db.GetUserData(ctx, req.Username)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}

	err = pswcrypt.CheckPaswword(user.Password, req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	token, err := auth.GenerateToken(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": token})
}
