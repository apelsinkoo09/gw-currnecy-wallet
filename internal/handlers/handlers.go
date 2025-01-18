package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type DepositRequest struct {
	Currency string  `json:"currency" binding:"required"`
	Amount   float64 `json:"amount" binding:"required,gt=0"`
}

type WithdrawRequest struct {
	Currency string  `json:"currency" binding:"required"`
	Amount   float64 `json:"amount" binding:"required,gt=0"`
}

// GetBalanceHandler godoc
// @Summary      Get wallet balance
// @Description  Retrieve the balance of the user's wallet in all available currencies
// @Tags         Wallet
// @Accept       json
// @Produce      json
// @Param        Authorization header string true "Bearer token"
// @Success      200  {object}  map[string]float64
// @Failure      401  {object}  map[string]string "Unauthorized"
// @Failure      500  {object}  map[string]string "Internal Server Error"
// @Router       /api/v1/balance [get]
func (s *WalletService) GetBalanceHandler(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in context"})
		return
	}

	ctx := c.Request.Context()
	balances, err := s.db.GetBalance(ctx, userID.(int))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch balance", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"balances": balances})
}

// DepositHandler godoc
// @Summary      Deposit money to wallet
// @Description  Add funds to the user's wallet in a specific currency
// @Tags         Wallet
// @Accept       json
// @Produce      json
// @Param        Authorization header string true "Bearer token"
//
//	@Param       input body DepositRequest true "Deposit information"
//
// @Success      200  {object}  map[string]string "Deposit successful"
// @Failure      401  {object}  map[string]string "Unauthorized"
// @Failure      400  {object}  map[string]string "Invalid input"
// @Failure      500  {object}  map[string]string "Internal Server Error"
// @Router       /api/v1/wallet/deposit [post]
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

// WithdrawHandler godoc
// @Summary      Withdraw money from wallet
// @Description  Withdraw funds from the user's wallet in a specific currency
// @Tags         Wallet
// @Accept       json
// @Produce      json
// @Param        Authorization header string true "Bearer token"
//
//	@Param       input body WithdrawRequest true "Withdrawal information"
//
// @Success      200  {object}  map[string]string "Withdrawal successful"
// @Failure      401  {object}  map[string]string "Unauthorized"
// @Failure      400  {object}  map[string]string "Invalid input"
// @Failure      500  {object}  map[string]string "Internal Server Error"
// @Router       /api/v1/wallet/withdraw [post]
func (s *WalletService) WithdrawHandler(c *gin.Context) {
	var req struct {
		Currency string  `json:"currency" binding:"required"`
		Amount   float64 `json:"amount" binding:"required,gt=0"`
	}

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
