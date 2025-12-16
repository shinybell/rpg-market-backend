package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/shinybell/rpg-market-backend/internal/infrastructure/auth"
)

// FirebaseAuth is a middleware that validates Firebase ID tokens
func FirebaseAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		// Extract token from "Bearer <token>"
		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == authHeader {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization format"})
			c.Abort()
			return
		}

		// Get Firebase client from infrastructure layer
		fbClient := auth.GetClient()
		if fbClient == nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Firebase client not initialized"})
			c.Abort()
			return
		}

		// Verify Firebase ID token
		decodedToken, err := fbClient.VerifyIDToken(c.Request.Context(), token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token: " + err.Error()})
			c.Abort()
			return
		}

		// Set user info in context
		c.Set("firebase_uid", decodedToken.UID)
		c.Set("email", decodedToken.Claims["email"])
		c.Next()
	}
}

// OptionalAuth is a middleware that tries to authenticate but doesn't abort if token is missing
func OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		fbClient := auth.GetClient()
		if fbClient == nil {
			c.Next()
			return
		}

		decodedToken, err := fbClient.VerifyIDToken(c.Request.Context(), token)
		if err == nil {
			c.Set("firebase_uid", decodedToken.UID)
			c.Set("email", decodedToken.Claims["email"])
		}
		c.Next()
	}
}
