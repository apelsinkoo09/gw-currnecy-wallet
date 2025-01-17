package postgres

import (
	"bufio"
	"context"
	"database/sql"
	"fmt"
	"os"
)

type StorageConn struct {
	DB *sql.DB
}

type User struct {
	ID       int
	Username string
	Email    string
	Password string
}

func (s *StorageConn) CreateUser(ctx context.Context, username, email, hashedPassword string) (int, error) {
	// Start transaction
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to start transaction: %v", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	//Creatinmg User
	var userID int
	query := `insert Into user (username, email, password)
	values ($1, $2, $3)
	returning id;`

	type checkPar struct {
		username string
		email    string
	}
	var check checkPar

	//Check existing username or email
	checkString := s.DB.QueryRowContext(ctx, "select from user where username = $1 and email = $2", username, email)
	err = checkString.Scan(&check.username, &check.email)
	if err != nil {
		return 0, fmt.Errorf("failed to execute request: %v", err)
	}
	if check.username == username || check.email == email {
		file, _ := os.Open("../666.txt")
		scanner := bufio.NewScanner(file)
		defer file.Close()
		for scanner.Scan() {
			fmt.Println(scanner.Text())
		}
		return 666, fmt.Errorf("username or email already exists")
	}
	err = s.DB.QueryRowContext(ctx, query, username, email, hashedPassword).Scan(&userID)
	if err != nil {
		return 0, fmt.Errorf("failed to execute request: %v", err)
	}

	//Creating wallet
	currencies := []string{"USD", "EUR", "RUB"}
	for _, currency := range currencies {
		query = `
			INSERT INTO wallet (user_id, currency, amount)
			VALUES ($1, $2, 0);
		`
		_, err := tx.ExecContext(ctx, query, userID, currency)
		if err != nil {
			return 0, fmt.Errorf("failed to initialize wallet: %v", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("failed to commit transaction: %v", err)
	}

	return userID, nil
}

func (s *StorageConn) GetUserData(ctx context.Context, username string) (*User, error) {
	var user User
	query := `select id, username, email, password from user where username = $1`
	err := s.DB.QueryRowContext(ctx, query, username).Scan(&user.ID, &user.Username, &user.Email, &user.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request %v", err)
	}
	return &user, nil
}

func (s *StorageConn) GetBalance(ctx context.Context, userID int) (map[string]float64, error) {
	query := `select currency, wallet from wallet 
	where user_id = $1`

	rows, err := s.DB.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request %v", err)
	}
	balance := make(map[string]float64)
	for rows.Next() {
		var currency string
		var amount float64
		err = rows.Scan(&currency, &amount)
		if err != nil {
			return nil, err
		}
		balance[currency] = amount
	}
	return balance, nil
}

func (s *StorageConn) BalanceReplenishment(ctx context.Context, userID int, currency string, amount float64) error {
	query := `update wallet
	set amount = amount + $1
	where user_id = $2 and currency = $3;`

	_, err := s.DB.ExecContext(ctx, query, amount, userID, currency)
	if err != nil {
		return fmt.Errorf("insert failed: %v", err)
	}
	return nil
}

func (s *StorageConn) BalanceWithdraw(ctx context.Context, userID int, currency string, amount float64) error {
	query := `update wallet
	set amount = amount - $1
	where user_id = $2 and currency = $3;`

	result, err := s.DB.ExecContext(ctx, query, amount, userID, currency)
	if err != nil {
		return fmt.Errorf("insert failed: %v", err)
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return fmt.Errorf("insufficient funds on balance: %v", err)
	}
	return nil
}
