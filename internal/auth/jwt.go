package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var jwtKey = []byte("1")


// GenerateToken - generate JWT token for user_id
func GenerateToken(userID int) (string, error){
	payLoad := jwt.MapClaims{
		"user_id": userID,
		"exp": time.Now().Add(24 * time.Hour).Unix(),
	}
	// Generate token with payload and signing method SHA256
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, payLoad)

	// signing token by secret key
	return token.SignedString(jwtKey)
}

//Checking the token and return user_id
func ValidateToken(tokenString string) (int, error){
	// Parse token string from request
	token, err := jwt.Parse(tokenString,func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok{
			return nil, jwt.ErrSignatureInvalid
		}
		return jwtKey, nil
	})
	if err != nil {
		return 0, nil

	}
	// Checking the validity and presence of the payload
	payLoad, ok := token.Claims.(jwt.MapClaims)
	if ok && token.Valid{
		userID := int(payLoad["user_id"].(float64))
		return userID, nil
	}
	return 0, jwt.ErrSignatureInvalid
}

