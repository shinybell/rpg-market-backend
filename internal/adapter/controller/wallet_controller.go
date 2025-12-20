package controller

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/shinybell/rpg-market-backend/internal/usecase"
)

// WalletController はウォレット関連のコントローラー
type WalletController struct {
	walletUseCase usecase.WalletUseCase
	userUseCase   *usecase.UserUseCase
}

// NewWalletController はWalletControllerを生成する
func NewWalletController(walletUseCase usecase.WalletUseCase, userUseCase *usecase.UserUseCase) *WalletController {
	return &WalletController{
		walletUseCase: walletUseCase,
		userUseCase:   userUseCase,
	}
}

// ChargeBalanceRequest はチャージリクエスト
type ChargeBalanceRequest struct {
	Amount      int64  `json:"amount" binding:"required,min=1"`
	Description string `json:"description"`
}

// WalletResponse はウォレットレスポンス
type WalletResponse struct {
	UserID  int64 `json:"user_id"`
	Balance int64 `json:"balance"`
	Points  int64 `json:"points"`
}

// TransactionHistoryResponse は取引履歴レスポンス
type TransactionHistoryResponse struct {
	ID              int64  `json:"id"`
	UserID          int64  `json:"user_id"`
	Type            string `json:"type"`
	Amount          int64  `json:"amount"`
	BalanceSnapshot int64  `json:"balance_snapshot"`
	Description     string `json:"description"`
	CreatedAt       string `json:"created_at"`
}

// ChargeBalance はウォレットに残高をチャージする
// @Summary ウォレットチャージ
// @Description 指定した金額をウォレットにチャージする（テスト用）
// @Tags wallet
// @Accept json
// @Produce json
// @Param request body ChargeBalanceRequest true "チャージリクエスト"
// @Success 200 {object} WalletResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /api/wallet/charge [post]
func (ctrl *WalletController) ChargeBalance(c *gin.Context) {
	// リクエストパース
	var req ChargeBalanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// ユーザーID取得
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// チャージ実行
	wallet, err := ctrl.walletUseCase.ChargeBalance(c.Request.Context(), userID.(int64), req.Amount, req.Description)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// レスポンス返却
	c.JSON(http.StatusOK, WalletResponse{
		UserID:  wallet.UserID,
		Balance: wallet.Balance,
		Points:  wallet.Points,
	})
}

// GetTransactionHistory はウォレット取引履歴を取得する
// @Summary ウォレット取引履歴取得
// @Description ユーザーのウォレット取引履歴を取得する
// @Tags wallet
// @Produce json
// @Param limit query int false "取得件数" default(20)
// @Param offset query int false "オフセット" default(0)
// @Success 200 {array} TransactionHistoryResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /api/wallet/transactions [get]
func (ctrl *WalletController) GetTransactionHistory(c *gin.Context) {
	// ユーザーID取得
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// クエリパラメータ取得
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	// 取引履歴取得
	transactions, err := ctrl.walletUseCase.GetTransactionHistory(c.Request.Context(), userID.(int64), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// レスポンス変換
	response := make([]TransactionHistoryResponse, len(transactions))
	for i, tx := range transactions {
		response[i] = TransactionHistoryResponse{
			ID:              tx.ID,
			UserID:          tx.UserID,
			Type:            tx.Type,
			Amount:          tx.Amount,
			BalanceSnapshot: tx.BalanceSnapshot,
			Description:     tx.Description,
			CreatedAt:       tx.CreatedAt.Format("2006-01-02 15:04:05"),
		}
	}

	c.JSON(http.StatusOK, response)
}
