package controller

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/shinybell/rpg-market-backend/internal/usecase"
)

type GenerationController struct {
	generationUseCase *usecase.GenerationUseCase
}

// NewGenerationController は新しいGenerationControllerを生成する
func NewGenerationController(generationUseCase *usecase.GenerationUseCase) *GenerationController {
	return &GenerationController{
		generationUseCase: generationUseCase,
	}
}

// GenerateDescriptionRequest はアイテム説明文生成のリクエスト
type GenerateDescriptionRequest struct {
	ItemName       string `json:"item_name" binding:"required,min=1,max=255"`
	Category       string `json:"category,omitempty"`
	Condition      string `json:"condition,omitempty"`
	NumSuggestions int    `json:"num_suggestions,omitempty"`
}

// GenerateDescriptionResponse はアイテム説明文生成のレスポンス
type GenerateDescriptionResponse struct {
	Suggestions []string `json:"suggestions"`
}

// @Summary アイテム説明文生成
// @Description Gemini APIを使ってアイテムの説明文候補を生成する
// @Tags generation
// @Accept json
// @Produce json
// @Param request body GenerateDescriptionRequest true "生成リクエスト"
// @Success 200 {object} GenerateDescriptionResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /api/items/description-suggestions [post]
func (ctrl *GenerationController) GenerateDescription(c *gin.Context) {
	var req GenerateDescriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// UseCaseのリクエストに変換
	useCaseReq := usecase.GenerateDescriptionRequest{
		ItemName:       req.ItemName,
		Category:       req.Category,
		Condition:      req.Condition,
		NumSuggestions: req.NumSuggestions,
	}

	// 説明文を生成
	result, err := ctrl.generationUseCase.GenerateDescriptions(c.Request.Context(), useCaseReq)
	if err != nil {
		// エラーログを出力
		log.Printf("[GENERATION ERROR] %v", err)

		// エラーの種類に応じてステータスコードを変更
		switch err {
		case usecase.ErrInvalidInput, usecase.ErrExceededMaxLength:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case usecase.ErrContentFiltered:
			c.JSON(http.StatusBadRequest, gin.H{"error": "Content filtered due to safety concerns"})
		case usecase.ErrGenerationFailed:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate descriptions"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		}
		return
	}

	// レスポンスを返す
	c.JSON(http.StatusOK, GenerateDescriptionResponse{
		Suggestions: result.Suggestions,
	})
}

// AppraiseItemRequest はアイテム鑑定のリクエスト
type AppraiseItemRequest struct {
	ItemName    string `json:"item_name" binding:"required,min=1,max=255"`
	Description string `json:"description,omitempty"`
	Category    string `json:"category,omitempty"`
	Condition   string `json:"condition,omitempty"`
}

// AppraiseItemResponse はアイテム鑑定のレスポンス
type AppraiseItemResponse struct {
	RPGName        string `json:"rpg_name"`
	RPGDescription string `json:"rpg_description"`
}

// @Summary アイテム鑑定（RPG風変換）
// @Description Gemini APIを使ってアイテムをRPG風の名前と説明文に変換する
// @Tags generation
// @Accept json
// @Produce json
// @Param request body AppraiseItemRequest true "鑑定リクエスト"
// @Success 200 {object} AppraiseItemResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /api/items/appraise [post]
func (ctrl *GenerationController) AppraiseItem(c *gin.Context) {
	var req AppraiseItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// UseCaseのリクエストに変換
	useCaseReq := usecase.AppraiseItemRequest{
		ItemName:    req.ItemName,
		Description: req.Description,
		Category:    req.Category,
		Condition:   req.Condition,
	}

	// 鑑定実行
	result, err := ctrl.generationUseCase.AppraiseItem(c.Request.Context(), useCaseReq)
	if err != nil {
		// エラーログを出力
		log.Printf("[APPRAISE ERROR] %v", err)

		// エラーの種類に応じてステータスコードを変更
		switch err {
		case usecase.ErrInvalidInput:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case usecase.ErrGenerationFailed:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to appraise item"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		}
		return
	}

	// レスポンスを返す
	c.JSON(http.StatusOK, AppraiseItemResponse{
		RPGName:        result.RPGName,
		RPGDescription: result.RPGDescription,
	})
}
