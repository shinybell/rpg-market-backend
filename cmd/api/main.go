package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/shinybell/rpg-market-backend/config"
	"github.com/shinybell/rpg-market-backend/internal/infrastructure/auth"
	"github.com/shinybell/rpg-market-backend/internal/infrastructure/middleware"

	_ "github.com/shinybell/rpg-market-backend/docs"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title RPG Market API
// @version 1.0
// @description RPGアイテム取引マーケットプレイスのAPI
// @host localhost:8080
// @BasePath /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	cfg := config.Load()

	// Initialize Firebase (Infrastructure Layer)
	_, err := auth.NewClient(cfg.FirebaseCredentialsPath)
	if err != nil {
		log.Fatalf("Failed to initialize Firebase: %v", err)
	}

	r := gin.Default()

	// Public routes
	r.GET("/", handleRoot)
	r.GET("/health", handleHealth)
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Protected routes (Interface Layer Middleware)
	protected := r.Group("/api")
	protected.Use(middleware.FirebaseAuth())
	{
		protected.GET("/me", handleMe)
	}

	log.Printf("Server listening on port %s (env: %s)", cfg.Port, cfg.Environment)
	r.Run(":" + cfg.Port)
}

// @Summary ルートエンドポイント
// @Description APIのウェルカムメッセージを返す
// @Tags root
// @Produce json
// @Success 200 {object} map[string]string
// @Router / [get]
func handleRoot(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "Welcome to the RPG Market Backend!",
	})
}

// @Summary ヘルスチェック
// @Description サーバーの稼働状態を確認
// @Tags health
// @Produce json
// @Success 200 {object} map[string]string
// @Router /health [get]
func handleHealth(c *gin.Context) {
	c.JSON(200, gin.H{
		"status": "ok",
	})
}

// @Summary ユーザープロフィール取得
// @Description 認証されたユーザーのプロフィール情報を取得
// @Tags user
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/me [get]
func handleMe(c *gin.Context) {
	firebaseUID, exists := c.Get("firebase_uid")
	if !exists {
		c.JSON(500, gin.H{"error": "Failed to get user from context"})
		return
	}

	email, _ := c.Get("email")

	c.JSON(200, gin.H{
		"firebase_uid": firebaseUID,
		"email":        email,
	})
}
