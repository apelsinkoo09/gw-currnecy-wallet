package handlers

import (
	"context"
	exchanger "gw-currncy-wallet/internal/changer"
	postgres "gw-currncy-wallet/internal/storages/postgres"
	"net/http"

	"github.com/gin-gonic/gin"
)

type WalletService struct {
	db        *postgres.StorageConn
	exchanger *exchanger.ExchangerClient
}

type ExchangeRequest struct {
	FromCurrency string  `json:"from_currency" binding:"required"` // Исходная валюта
	ToCurrency   string  `json:"to_currency" binding:"required"`   // Целевая валюта
	Amount       float64 `json:"amount" binding:"required,gt=0"`   // Сумма для обмена
}

// NewWalletService create new wallet service
func NewWalletService(db *postgres.StorageConn, exchanger *exchanger.ExchangerClient) *WalletService {
	return &WalletService{
		db:        db,
		exchanger: exchanger,
	}
}

// ExchangeHandler godoc
// @Summary      Exchange currency
// @Description  Exchange one currency for another based on the current exchange rate
// @Tags         Wallet
// @Accept       json
// @Produce      json
// @Param        Authorization header string true "Bearer token"
//
//	@Param       input body ExchangeRequest true "Exchange information"
//
// @Success      200  {object}  map[string]string "Exchange successful"
// @Failure      401  {object}  map[string]string "Unauthorized"
// @Failure      400  {object}  map[string]string "Invalid input"
// @Failure      500  {object}  map[string]string "Internal Server Error"
// @Router       /api/v1/wallet/exchange [post]
func (s *WalletService) ExchangeHandler(c *gin.Context) {
	var req struct {
		FromCurrency string  `json:"from_currency" binding:"required"`
		ToCurrency   string  `json:"to_currency" binding:"required"`
		Amount       float64 `json:"amount" binding:"required,gt=0"`
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
	err := s.Exchange(ctx, userID.(int), req.FromCurrency, req.ToCurrency, req.Amount)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Exchange successful"})
}

func (s *WalletService) Exchange(ctx context.Context, userID int, fromCurrency, toCurrency string, amount float64) error {
	exchangeRate, err := s.exchanger.GetExchangeRate(ctx, fromCurrency, toCurrency)
	if err != nil {
		return err
	}

	tx, err := s.db.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	withdrawQuery := `
		UPDATE wallet
		SET amount = amount - $1
		WHERE user_id = $2 AND currency = $3 AND amount >= $1;
	`
	_, err = tx.ExecContext(ctx, withdrawQuery, amount, userID, fromCurrency)
	if err != nil {
		return err
	}

	toAmount := amount * exchangeRate
	depositQuery := `
		UPDATE wallet
		SET amount = amount + $1
		WHERE user_id = $2 AND currency = $3;
	`
	_, err = tx.ExecContext(ctx, depositQuery, toAmount, userID, toCurrency)
	if err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}
