package auth

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// JWTMiddleware checks the token on every request and adds the user_id to the context.
func JWTMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract the Authorization header from the request.
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			// If the header is missing or the token does not start with "Bearer", return 401 Unauthorized.
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing or invalid token"})
			c.Abort()
			return
		}

		// Remove "Bearer" from the token string.
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// Check valid token
		userID, err := ValidateToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		// Save the user_id into the request context.
		// This will allow other handlers to use the user_id.
		c.Set("user_id", userID)

		// Pass control to the next handler.
		c.Next()
	}
}