package controller

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/shinybell/rpg-market-backend/internal/usecase"
)

type FollowController struct {
	followUseCase *usecase.FollowUseCase
}

// NewFollowController はFollowControllerを生成する
func NewFollowController(followUseCase *usecase.FollowUseCase) *FollowController {
	return &FollowController{
		followUseCase: followUseCase,
	}
}

// AddFollow はフォローを追加する
func (c *FollowController) AddFollow(ctx *gin.Context) {
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

	followeeIDStr := ctx.Param("id")
	followeeID, err := strconv.ParseInt(followeeIDStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid followee id"})
		return
	}

	if err := c.followUseCase.AddFollow(ctx.Request.Context(), userIDInt, followeeID); err != nil {
		if err == usecase.ErrFollowAlreadyExists {
			ctx.JSON(http.StatusConflict, gin.H{"error": "already following"})
			return
		}
		if err == usecase.ErrSelfFollow {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "cannot follow yourself"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"message": "followed"})
}

// RemoveFollow はフォローを削除する
func (c *FollowController) RemoveFollow(ctx *gin.Context) {
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

	followeeIDStr := ctx.Param("id")
	followeeID, err := strconv.ParseInt(followeeIDStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid followee id"})
		return
	}

	if err := c.followUseCase.RemoveFollow(ctx.Request.Context(), userIDInt, followeeID); err != nil {
		if err == usecase.ErrFollowNotFound {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "follow not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "unfollowed"})
}
