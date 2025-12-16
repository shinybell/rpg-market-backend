package main

import (
	"log"

	"github.com/shinybell/rpg-market-backend/config"

	"github.com/gin-gonic/gin"

	_ "github.com/shinybell/rpg-market-backend/docs"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title RPG Market API
// @version 1.0
// @description RPGアイテム取引マーケットプレイスのAPI
// @host localhost:8080
// @BasePath /
func main() {
	cfg := config.Load()

	r := gin.Default()

	r.GET("/", handleRoot)
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	r.GET("/health", handleHealth)

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
