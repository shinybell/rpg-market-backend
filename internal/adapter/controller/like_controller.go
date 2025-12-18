package controller

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/shinybell/rpg-market-backend/internal/usecase"
)

type LikeController struct {
	likeUseCase *usecase.LikeUseCase
}

// NewLikeController はLikeControllerを生成する
func NewLikeController(likeUseCase *usecase.LikeUseCase) *LikeController {
	return &LikeController{
		likeUseCase: likeUseCase,
	}
}

// AddLike はいいねを追加する
func (c *LikeController) AddLike(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}
	userIDInt, ok := userID.(int64)
	if !ok {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user id type"})
		return
	}

	itemIDStr := ctx.Param("id")
	itemID, err := strconv.ParseInt(itemIDStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid item id"})
		return
	}

	if err := c.likeUseCase.AddLike(ctx.Request.Context(), userIDInt, itemID); err != nil {
		if err == usecase.ErrLikeAlreadyExists {
			ctx.JSON(http.StatusConflict, gin.H{"error": "already liked"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"message": "liked"})
}

// RemoveLike はいいねを削除する
func (c *LikeController) RemoveLike(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}
	userIDInt, ok := userID.(int64)
	if !ok {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user id type"})
		return
	}

	itemIDStr := ctx.Param("id")
	itemID, err := strconv.ParseInt(itemIDStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid item id"})
		return
	}

	if err := c.likeUseCase.RemoveLike(ctx.Request.Context(), userIDInt, itemID); err != nil {
		if err == usecase.ErrLikeNotFound {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "like not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "unliked"})
}

// GetLikeStatus はいいね状態を取得する
func (c *LikeController) GetLikeStatus(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}
	userIDInt, ok := userID.(int64)
	if !ok {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user id type"})
		return
	}

	itemIDStr := ctx.Param("id")
	itemID, err := strconv.ParseInt(itemIDStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid item id"})
		return
	}

	liked, err := c.likeUseCase.GetLikeStatus(ctx.Request.Context(), userIDInt, itemID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"liked": liked})
}
