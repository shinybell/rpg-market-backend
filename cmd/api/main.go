package main

import (
	"log"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/shinybell/rpg-market-backend/config"
	"github.com/shinybell/rpg-market-backend/internal/adapter/controller"
	"github.com/shinybell/rpg-market-backend/internal/infrastructure/auth"
	"github.com/shinybell/rpg-market-backend/internal/infrastructure/db"
	"github.com/shinybell/rpg-market-backend/internal/infrastructure/db/mysql"
	"github.com/shinybell/rpg-market-backend/internal/infrastructure/middleware"
	"github.com/shinybell/rpg-market-backend/internal/usecase"

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

	// Initialize Firebase
	_, err := auth.NewClient(cfg.FirebaseCredentialsPath)
	if err != nil {
		log.Fatalf("Failed to initialize Firebase: %v", err)
	}

	// Initialize Database
	database, err := db.NewDatabase(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()

	// Initialize repositories
	userRepo := mysql.NewUserRepository(database.GetDB())

	// Initialize use cases
	userUseCase := usecase.NewUserUseCase(userRepo)

	// Initialize controllers
	userController := controller.NewUserController(userUseCase)

	// Setup router
	r := gin.Default()

	// CORS設定（最優先で適用）
	r.Use(cors.New(cors.Config{
		AllowOrigins: []string{
			"http://localhost:5173",
			"http://localhost:5174",
			"http://localhost:3000",
			cfg.VercelDeployURL,
		},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Public routes
	r.GET("/", handleRoot)
	r.GET("/health", handleHealth)
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Protected routes
	api := r.Group("/api")
	api.Use(middleware.FirebaseAuth())
	{
		// 認証関連
		api.POST("/auth/login", userController.Login)
		api.GET("/auth/me", userController.GetMe)

		// ユーザー管理
		api.PUT("/users/profile", userController.UpdateProfile)
		api.DELETE("/users", userController.DeleteUser)
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
