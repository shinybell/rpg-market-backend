package main

import (
	"log"

	"github.com/shinybell/rpg-market-backend/config"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()

	r := gin.Default()

	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Welcome to the RPG Market Backend!",
		})
	})

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})

	log.Printf("Server listening on port %s (env: %s)", cfg.Port, cfg.Environment)
	r.Run(":" + cfg.Port)
}
