package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (s *WalletService) GetBalanceHandler(c *gin.Context) {
	userID := 1

	ctx := c.Request.Context()
	balances, err := s.db.GetBalance(ctx, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch balance", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"balances": balances})
}

func (s *WalletService) DepositHandler(c *gin.Context) {
	var req struct {
		Currency string  `json:"currency" binding:"required"`
		Amount   float64 `json:"amount" binding:"required,gt=0"`
	}

	// Парсим тело запроса
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
    c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in context"})
    return
	}

	ctx := c.Request.Context()
	err := s.db.BalanceReplenishment(ctx, userID.(int), req.Currency, req.Amount)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to deposit", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Deposit successful"})
}

// WithdrawHandler обрабатывает запрос на снятие средств
func (s *WalletService) WithdrawHandler(c *gin.Context) {
	var req struct {
		Currency string  `json:"currency" binding:"required"`
		Amount   float64 `json:"amount" binding:"required,gt=0"`
	}

	// Парсим тело запроса
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
    c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in context"})
    return
	}

	ctx := c.Request.Context()
	err := s.db.BalanceWithdraw(ctx, userID.(int), req.Currency, req.Amount)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to withdraw", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Withdrawal successful"})
}
