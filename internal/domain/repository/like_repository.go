package repository

import (
	"context"

	"github.com/shinybell/rpg-market-backend/internal/domain/entity"
)

// LikeRepository はいいねのリポジトリインターフェース
type LikeRepository interface {
	// Create はいいねを作成する
	Create(ctx context.Context, like *entity.Like) error

	// Delete はいいねを削除する
	Delete(ctx context.Context, userID, itemID int64) error

	// ExistsByUserAndItem は指定されたユーザーとアイテムのいいねが存在するかを確認する
	ExistsByUserAndItem(ctx context.Context, userID, itemID int64) (bool, error)
}
