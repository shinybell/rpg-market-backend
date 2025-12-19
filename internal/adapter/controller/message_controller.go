package controller

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/shinybell/rpg-market-backend/internal/domain/repository"
	"github.com/shinybell/rpg-market-backend/internal/infrastructure/auth"
	"github.com/shinybell/rpg-market-backend/internal/infrastructure/websocket"
	"github.com/shinybell/rpg-market-backend/internal/usecase"

	"github.com/gin-gonic/gin"
	ws "github.com/gorilla/websocket"
)

// MessageController はメッセージのコントローラー
type MessageController struct {
	messageUsecase usecase.MessageUsecase
	userRepo       repository.UserRepository
	hub            *websocket.Hub
}

// NewMessageController はMessageControllerを生成
func NewMessageController(messageUsecase usecase.MessageUsecase, userRepo repository.UserRepository, hub *websocket.Hub) *MessageController {
	return &MessageController{
		messageUsecase: messageUsecase,
		userRepo:       userRepo,
		hub:            hub,
	}
}

var upgrader = ws.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// 本番環境では適切なCORS設定が必要
		return true
	},
}

// HandleWebSocket はWebSocket接続を処理
// @Summary WebSocket接続
// @Description 取引に紐づくメッセージをリアルタイムで送受信
// @Tags messages
// @Param id path int true "Transaction ID"
// @Router /ws/transactions/{id} [get]
func (c *MessageController) HandleWebSocket(ctx *gin.Context) {
	transactionIDStr := ctx.Param("id")
	transactionID, err := strconv.ParseInt(transactionIDStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid transaction id"})
		return
	}

	// WebSocketにアップグレード
	conn, err := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		log.Printf("failed to upgrade connection: %v", err)
		return
	}

	// 最初のメッセージで認証
	_, message, err := conn.ReadMessage()
	if err != nil {
		log.Printf("failed to read auth message: %v", err)
		return
	}

	// 認証メッセージをパース
	var authMsg struct {
		Type  string `json:"type"`
		Token string `json:"token"`
	}
	if err := json.Unmarshal(message, &authMsg); err != nil {
		log.Printf("failed to parse auth message: %v", err)
		conn.WriteJSON(map[string]string{"error": "invalid auth message"})
		return
	}

	if authMsg.Type != "auth" || authMsg.Token == "" {
		log.Println("invalid auth message type or empty token")
		conn.WriteJSON(map[string]string{"error": "authentication required"})
		return
	}

	// Firebaseトークン検証
	client := auth.GetClient()
	if client == nil {
		log.Println("Firebase client not initialized")
		conn.WriteJSON(map[string]string{"error": "server error"})
		return
	}

	token, err := client.VerifyIDToken(context.Background(), authMsg.Token)
	if err != nil {
		log.Printf("failed to verify token: %v", err)
		conn.WriteJSON(map[string]string{"error": "invalid token"})
		return
	}

	// ユーザー情報を取得
	user, err := c.userRepo.FindByFirebaseUID(context.Background(), token.UID)
	if err != nil || user == nil {
		log.Printf("user not found: %v", err)
		conn.WriteJSON(map[string]string{"error": "user not found"})
		return
	}

	// アクセス権チェック
	if err := c.messageUsecase.ValidateAccess(context.Background(), transactionID, user.ID); err != nil {
		log.Printf("access denied: %v", err)
		conn.WriteJSON(map[string]string{"error": "access denied"})
		return
	}

	// 認証成功を通知
	conn.WriteJSON(map[string]string{"type": "auth_success"})

	// クライアント作成
	wsClient := websocket.NewClient(c.hub, conn, user.ID, transactionID)

	// メッセージ受信時の処理
	onMessage := func(content string) {
		msg, err := c.messageUsecase.SendMessage(context.Background(), transactionID, user.ID, content)
		if err != nil {
			log.Printf("failed to send message: %v", err)
			return
		}

		// 同じ取引のクライアントにブロードキャスト
		c.hub.Broadcast(transactionID, msg)
	}

	// クライアント開始
	wsClient.Start(onMessage)
}

// SendMessage はメッセージを送信（REST API）
// @Summary メッセージ送信
// @Description 取引に紐づくメッセージを送信
// @Tags messages
// @Accept json
// @Produce json
// @Param id path int true "Transaction ID"
// @Param body body object true "Message content"
// @Success 201 {object} entity.Message
// @Router /api/transactions/{id}/messages [post]
func (c *MessageController) SendMessage(ctx *gin.Context) {
	transactionIDStr := ctx.Param("id")
	transactionID, err := strconv.ParseInt(transactionIDStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid transaction id"})
		return
	}

	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req struct {
		Content string `json:"content" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	message, err := c.messageUsecase.SendMessage(context.Background(), transactionID, userID.(int64), req.Content)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// WebSocketでもブロードキャスト
	c.hub.Broadcast(transactionID, message)

	ctx.JSON(http.StatusCreated, message)
}

// GetMessages はメッセージ一覧を取得
// @Summary メッセージ一覧取得
// @Description 取引に紐づくメッセージ一覧を取得
// @Tags messages
// @Produce json
// @Param id path int true "Transaction ID"
// @Success 200 {array} entity.Message
// @Router /api/transactions/{id}/messages [get]
func (c *MessageController) GetMessages(ctx *gin.Context) {
	transactionIDStr := ctx.Param("id")
	transactionID, err := strconv.ParseInt(transactionIDStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid transaction id"})
		return
	}

	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	messages, err := c.messageUsecase.GetMessages(context.Background(), transactionID, userID.(int64))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, messages)
}

// GetTransaction はトランザクション情報を取得（取引の当事者のみ）
// @Summary 取引情報取得
// @Description トランザクションIDで取引情報を取得（購入者または出品者のみ）
// @Tags messages
// @Produce json
// @Param id path int true "Transaction ID"
// @Success 200 {object} entity.Transaction
// @Router /api/transactions/{id} [get]
func (c *MessageController) GetTransaction(ctx *gin.Context) {
	txIDStr := ctx.Param("id")
	txID, err := strconv.ParseInt(txIDStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid transaction id"})
		return
	}

	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	tx, err := c.messageUsecase.GetTransactionByIDForUser(ctx.Request.Context(), txID, userID.(int64))
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, tx)
}

// GetUnreadCount は未読メッセージ数を取得
// @Summary 未読メッセージ数取得
// @Description ユーザーの未読メッセージ数を取得
// @Tags messages
// @Produce json
// @Success 200 {object} object{count=int}
// @Router /api/messages/unread [get]
func (c *MessageController) GetUnreadCount(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	count, err := c.messageUsecase.GetUnreadCount(userID.(int64))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"count": count})
}
