package controller

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/shinybell/rpg-market-backend/internal/usecase"
)

type CommentController struct {
	commentUseCase *usecase.CommentUseCase
}

// NewCommentController はCommentControllerを生成する
func NewCommentController(commentUseCase *usecase.CommentUseCase) *CommentController {
	return &CommentController{
		commentUseCase: commentUseCase,
	}
}

type AddCommentRequest struct {
	Comment string `json:"comment" binding:"required,min=1,max=500"`
}

// AddComment はコメントを追加する
func (c *CommentController) AddComment(ctx *gin.Context) {
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

	var req AddCommentRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		log.Printf("Bind error: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.Printf("Adding comment: userID=%d, itemID=%d, comment=%s", userIDInt, itemID, req.Comment)

	if err := c.commentUseCase.AddComment(ctx.Request.Context(), userIDInt, itemID, req.Comment); err != nil {
		log.Printf("AddComment error: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"message": "comment added"})
}

// GetComments はコメントを取得する
func (c *CommentController) GetComments(ctx *gin.Context) {
	itemIDStr := ctx.Param("id")
	itemID, err := strconv.ParseInt(itemIDStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid item id"})
		return
	}

	limitStr := ctx.DefaultQuery("limit", "10")
	offsetStr := ctx.DefaultQuery("offset", "0")
	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)

	comments, err := c.commentUseCase.GetCommentsByItemID(ctx.Request.Context(), itemID, limit, offset)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, comments)
}

// DeleteComment はコメントを削除する
func (c *CommentController) DeleteComment(ctx *gin.Context) {
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

	commentIDStr := ctx.Param("comment_id")
	commentID, err := strconv.ParseInt(commentIDStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid comment id"})
		return
	}

	if err := c.commentUseCase.DeleteComment(ctx.Request.Context(), commentID, userIDInt); err != nil {
		if err == usecase.ErrCommentNotFound {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "comment not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "comment deleted"})
}
