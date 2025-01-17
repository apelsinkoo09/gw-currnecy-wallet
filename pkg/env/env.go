package env

import (
	"fmt"

	"github.com/joho/godotenv"
)

func LoadConfig(filename string) error {
	err := godotenv.Load(filename)
	if err != nil {
		return fmt.Errorf("failed to load env file: %w", err)
	}
	return nil
}
