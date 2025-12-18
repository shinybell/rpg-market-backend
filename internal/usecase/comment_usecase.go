package usecase

import (
	"context"
	"errors"

	"github.com/shinybell/rpg-market-backend/internal/domain/entity"
	"github.com/shinybell/rpg-market-backend/internal/domain/repository"
)

var (
	ErrCommentNotFound = errors.New("comment not found")
)

type CommentUseCase struct {
	commentRepo repository.CommentRepository
	itemRepo    repository.ItemRepository
}

// NewCommentUseCase はCommentUseCaseを生成する
func NewCommentUseCase(commentRepo repository.CommentRepository, itemRepo repository.ItemRepository) *CommentUseCase {
	return &CommentUseCase{
		commentRepo: commentRepo,
		itemRepo:    itemRepo,
	}
}

// AddComment はコメントを追加する
func (uc *CommentUseCase) AddComment(ctx context.Context, userID, itemID int64, commentText string) error {
	// アイテム存在確認
	item, err := uc.itemRepo.FindByID(ctx, itemID)
	if err != nil {
		return err
	}
	if item == nil {
		return ErrItemNotFound
	}

	// コメント作成
	comment := &entity.Comment{
		ItemID:  itemID,
		UserID:  userID,
		Comment: commentText,
	}
	return uc.commentRepo.Create(ctx, comment)
}

// GetCommentsByItemID はアイテムIDでコメントを取得する
func (uc *CommentUseCase) GetCommentsByItemID(ctx context.Context, itemID int64, limit, offset int) ([]*entity.Comment, error) {
	return uc.commentRepo.FindByItemID(ctx, itemID, limit, offset)
}

// DeleteComment はコメントを削除する
func (uc *CommentUseCase) DeleteComment(ctx context.Context, commentID, userID int64) error {
	// コメント削除（論理削除）
	return uc.commentRepo.Delete(ctx, commentID, userID)
}
