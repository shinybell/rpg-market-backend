package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/shinybell/rpg-market-backend/config"
	"github.com/shinybell/rpg-market-backend/internal/adapter/controller"
	"github.com/shinybell/rpg-market-backend/internal/infrastructure/auth"
	"github.com/shinybell/rpg-market-backend/internal/infrastructure/db"
	"github.com/shinybell/rpg-market-backend/internal/infrastructure/db/mysql"
	"github.com/shinybell/rpg-market-backend/internal/infrastructure/gemini"
	"github.com/shinybell/rpg-market-backend/internal/infrastructure/middleware"
	"github.com/shinybell/rpg-market-backend/internal/infrastructure/storage"
	"github.com/shinybell/rpg-market-backend/internal/infrastructure/websocket"
	"github.com/shinybell/rpg-market-backend/internal/usecase"

	_ "github.com/shinybell/rpg-market-backend/docs"
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
	defer func() {
		if err := database.Close(); err != nil {
			log.Printf("Error closing database: %v", err)
		}
	}()

	// Initialize repositories
	userRepo := mysql.NewUserRepository(database.GetDB())
	itemRepo := mysql.NewItemRepository(database.GetDB())
	likeRepo := mysql.NewLikeRepository(database.GetDB())
	commentRepo := mysql.NewCommentRepository(database.GetDB())
	followRepo := mysql.NewFollowRepository(database.GetDB())
	transactionRepo := mysql.NewTransactionRepository(database.GetDB())
	walletRepo := mysql.NewWalletRepository(database.GetDB())
	notificationRepo := mysql.NewNotificationRepository(database.GetDB())
	addressRepo := mysql.NewAddressRepository(database.GetDB())
	messageRepo := mysql.NewMessageRepository(database.GetDB())

	// Initialize GCS client
	ctx := context.Background()
	gcsClient, err := storage.NewGCSClient(ctx, cfg)
	if err != nil {
		log.Printf("Warning: Failed to initialize GCS client: %v", err)
		// GCS初期化失敗は致命的ではないため、継続
	} else {
		defer func() {
			if err := gcsClient.Close(); err != nil {
				log.Printf("Error closing GCS client: %v", err)
			}
		}()
	}

	// Initialize Gemini client
	geminiClient := gemini.NewClient(cfg.GeminiAPIKey, cfg.GeminiModel)

	// Initialize use cases
	userUseCase := usecase.NewUserUseCase(userRepo)
	notificationUseCase := usecase.NewNotificationUseCase(notificationRepo)
	addressUseCase := usecase.NewAddressUseCase(addressRepo)
	itemUseCase := usecase.NewItemUseCase(itemRepo, commentRepo, transactionRepo, walletRepo, notificationRepo, addressRepo, notificationUseCase)
	likeUseCase := usecase.NewLikeUseCase(likeRepo, itemRepo)
	commentUseCase := usecase.NewCommentUseCase(commentRepo, itemRepo)
	followUseCase := usecase.NewFollowUseCase(followRepo, userRepo)
	messageUseCase := usecase.NewMessageUsecase(messageRepo, transactionRepo)
	generationUseCase := usecase.NewGenerationUseCase(geminiClient)

	// Initialize WebSocket Hub
	hub := websocket.NewHub()
	go hub.Run()

	// Initialize controllers
	userController := controller.NewUserController(userUseCase)
	addressController := controller.NewAddressController(addressUseCase, userUseCase)
	itemController := controller.NewItemController(itemUseCase, userUseCase)
	uploadController := controller.NewUploadController(gcsClient)
	likeController := controller.NewLikeController(likeUseCase)
	commentController := controller.NewCommentController(commentUseCase)
	followController := controller.NewFollowController(followUseCase)
	messageController := controller.NewMessageController(messageUseCase, userRepo, hub)
	generationController := controller.NewGenerationController(generationUseCase)

	// Setup router
	gin.SetMode(cfg.LogLevel)
	r := gin.Default()

	// Custom logger to include request body for debugging
	r.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("[%s] %s %s %d %s %s\n",
			param.TimeStamp.Format("2006/01/02 15:04:05"),
			param.Method,
			param.Path,
			param.StatusCode,
			param.Latency,
			param.Request.UserAgent(),
		)
	}))

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

	// Request body logger for debugging
	r.Use(middleware.RequestBodyLogger())

	// Public routes
	r.GET("/", handleRoot)
	r.GET("/health", handleHealth)
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Public item routes (認証不要)
	r.GET("/api/items", itemController.ListItems)
	r.GET("/api/items/search", itemController.SearchItems) // searchは:idより前に定義
	r.GET("/api/items/seller/:seller_id", itemController.ListItemsBySeller)
	r.GET("/api/items/category/:category_id", itemController.ListItemsByCategory)
	r.GET("/api/items/:id", itemController.GetItem)

	// WebSocket routes (接続後に認証)
	r.GET("/api/ws/transactions/:id", messageController.HandleWebSocket)

	// Protected routes
	api := r.Group("/api")
	api.Use(middleware.FirebaseAuth(userRepo))
	{
		// 認証関連
		api.POST("/auth/login", userController.Login)
		api.GET("/auth/me", userController.GetMe)

		// ユーザー管理
		api.PUT("/users/profile", userController.UpdateProfile)
		api.DELETE("/users", userController.DeleteUser)
		api.GET("/users/me/items", itemController.GetMyItems)
		api.GET("/users/me/purchases", itemController.GetMyPurchases)

		// 配送先管理
		api.GET("/addresses", addressController.GetAddresses)
		api.POST("/addresses", addressController.CreateAddress)
		api.PUT("/addresses/:id", addressController.UpdateAddress)
		api.DELETE("/addresses/:id", addressController.DeleteAddress)

		// アイテム管理（認証必須）
		api.POST("/items", itemController.CreateItem)
		api.PUT("/items/:id", itemController.UpdateItem)
		api.DELETE("/items/:id", itemController.DeleteItem)
		api.POST("/items/:id/purchase", itemController.PurchaseItem)
		api.GET("/items/:id/transaction", itemController.GetItemTransaction)

		// 画像アップロード
		api.POST("/upload/signed-url", uploadController.GenerateSignedURL)

		// いいね管理
		api.POST("/items/:id/likes", likeController.AddLike)
		api.DELETE("/items/:id/likes", likeController.RemoveLike)
		api.GET("/items/:id/likes/status", likeController.GetLikeStatus)

		// コメント管理
		api.POST("/items/:id/comments", commentController.AddComment)
		api.GET("/items/:id/comments", commentController.GetComments)
		api.DELETE("/comments/:comment_id", commentController.DeleteComment)

		// フォロー管理
		api.POST("/users/:id/follow", followController.AddFollow)
		api.DELETE("/users/:id/follow", followController.RemoveFollow)

		// メッセージ管理
		api.POST("/transactions/:id/messages", messageController.SendMessage)
		api.GET("/transactions/:id/messages", messageController.GetMessages)
		// トランザクション情報取得（取引当事者のみ）
		api.GET("/transactions/:id", messageController.GetTransaction)
		api.GET("/messages/unread", messageController.GetUnreadCount)

		// AI生成機能
		api.POST("/items/description-suggestions", generationController.GenerateDescription)
	}

	log.Printf("Server listening on port %s (env: %s)", cfg.Port, cfg.Environment)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
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
