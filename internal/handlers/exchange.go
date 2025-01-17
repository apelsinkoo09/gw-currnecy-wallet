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
	exchanger *exchanger.ExchangerClient // gRPC-клиент
}

// NewWalletService создает новый сервис кошельков
func NewWalletService(db *postgres.StorageConn, exchanger *exchanger.ExchangerClient) *WalletService {
	return &WalletService{
		db:        db,
		exchanger: exchanger,
	}
}

func (s *WalletService) ExchangeHandler(c *gin.Context) {
	var req struct {
		FromCurrency string  `json:"from_currency" binding:"required"`
		ToCurrency   string  `json:"to_currency" binding:"required"`
		Amount       float64 `json:"amount" binding:"required,gt=0"`
	}

	// Декодируем тело запроса
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in context"})
		return
	}

	// Выполняем обмен валют
	ctx := c.Request.Context()
	err := s.Exchange(ctx, userID.(int), req.FromCurrency, req.ToCurrency, req.Amount)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Exchange successful"})
}

// Exchange выполняет обмен валют
func (s *WalletService) Exchange(ctx context.Context, userID int, fromCurrency, toCurrency string, amount float64) error {
	// 1. Получаем курс обмена из кеша или gRPC
	exchangeRate, err := s.exchanger.GetExchangeRate(ctx, fromCurrency, toCurrency)
	if err != nil {
		return err
	}

	// 2. Открываем транзакцию
	tx, err := s.db.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// 3. Уменьшаем баланс в исходной валюте
	withdrawQuery := `
		UPDATE wallet
		SET amount = amount - $1
		WHERE user_id = $2 AND currency = $3 AND amount >= $1;
	`
	_, err = tx.ExecContext(ctx, withdrawQuery, amount, userID, fromCurrency)
	if err != nil {
		return err
	}

	// 4. Увеличиваем баланс в целевой валюте
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

	// 5. Фиксируем транзакцию
	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}
