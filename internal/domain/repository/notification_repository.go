package repository

import (
	"context"

	"github.com/shinybell/rpg-market-backend/internal/domain/entity"
)

// NotificationRepository は通知のリポジトリインターフェース
type NotificationRepository interface {
	// Create は通知を作成する
	Create(ctx context.Context, n *entity.Notification) error
}
