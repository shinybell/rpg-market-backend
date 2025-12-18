package repository

import (
	"context"

	"github.com/shinybell/rpg-market-backend/internal/domain/entity"
)

// CommentRepository はコメントのリポジトリインターフェース
type CommentRepository interface {
	// Create はコメントを作成する
	Create(ctx context.Context, comment *entity.Comment) error

	// FindByItemID はアイテムIDでコメントを検索する
	FindByItemID(ctx context.Context, itemID int64, limit, offset int) ([]*entity.Comment, error)

	// CountByItemID はアイテムIDのコメント数をカウントする
	CountByItemID(ctx context.Context, itemID int64) (int, error)

	// Delete はコメントを削除する
	Delete(ctx context.Context, commentID, userID int64) error
}
