package api

import (
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// JWTClaims represents the claims in our JWT tokens
type JWTClaims struct {
	UserID       string `json:"user_id"`
	EncryptedDEK string `json:"encrypted_dek"`
	jwt.RegisteredClaims
}

// ContextKey type for context values
type ContextKey string

const (
	UserIDKey       ContextKey = "user_id"
	EncryptedDEKKey ContextKey = "encrypted_dek"
)

// jwtMiddleware validates JWT tokens and extracts user info
func (s *Server) jwtMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		authHeader := c.Request().Header.Get("Authorization")
		if authHeader == "" {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "missing authorization header"})
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid authorization header"})
		}

		tokenString := parts[1]
		claims := &JWTClaims{}

		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(s.config.JWTSecret), nil
		})

		if err != nil || !token.Valid {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid token"})
		}

		userID, err := uuid.Parse(claims.UserID)
		if err != nil {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid user ID in token"})
		}

		// Store user info in context
		c.Set(string(UserIDKey), userID)
		c.Set(string(EncryptedDEKKey), claims.EncryptedDEK)

		return next(c)
	}
}

// getUserID extracts the user ID from the context
func getUserID(c echo.Context) (uuid.UUID, bool) {
	userID, ok := c.Get(string(UserIDKey)).(uuid.UUID)
	return userID, ok
}

// getEncryptedDEK extracts the encrypted DEK from the context
func getEncryptedDEK(c echo.Context) (string, bool) {
	dek, ok := c.Get(string(EncryptedDEKKey)).(string)
	return dek, ok
}
