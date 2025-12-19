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
