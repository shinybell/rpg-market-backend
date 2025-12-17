package middleware

import (
	"context"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/shinybell/rpg-market-backend/internal/infrastructure/auth"
)

// FirebaseAuth はFirebase認証を検証するミドルウェア
func FirebaseAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// OPTIONSリクエスト（プリフライト）はスキップ
		if c.Request.Method == "OPTIONS" {
			c.Next()
			return
		}

		authHeader := c.GetHeader("Authorization")

		log.Printf("[AUTH] Method: %s, Path: %s", c.Request.Method, c.Request.URL.Path)
		log.Printf("[AUTH] Authorization header: %s", authHeader)

		// Authorizationヘッダーの確認
		if authHeader == "" {
			log.Println("[AUTH] No authorization header")
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header is required",
			})
			c.Abort()
			return
		}

		// Bearer トークンの抽出
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			log.Printf("[AUTH] Invalid header format: %v", parts)
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid authorization header format. Expected 'Bearer <token>'",
			})
			c.Abort()
			return
		}

		idToken := parts[1]
		log.Printf("[AUTH] Token length: %d", len(idToken))

		// Firebaseトークンの検証
		client := auth.GetClient()
		if client == nil {
			log.Println("[AUTH] Firebase client is nil")
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Firebase client not initialized",
			})
			c.Abort()
			return
		}

		token, err := client.VerifyIDToken(context.Background(), idToken)
		if err != nil {
			log.Printf("[AUTH] Token verification failed: %v", err)
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Invalid or expired token",
				"details": err.Error(),
			})
			c.Abort()
			return
		}

		log.Printf("[AUTH] Token verified successfully for UID: %s", token.UID)

		// ユーザー情報をコンテキストに保存
		c.Set("firebase_uid", token.UID)
		c.Set("email", token.Claims["email"])

		c.Next()
	}
}
