package mysql

import (
	"context"

	"gorm.io/gorm"

	"github.com/shinybell/rpg-market-backend/internal/domain/entity"
	"github.com/shinybell/rpg-market-backend/internal/domain/repository"
)

type notificationRepository struct {
	db *gorm.DB
}

// NewNotificationRepository はNotificationRepositoryの実装を返す
func NewNotificationRepository(db *gorm.DB) repository.NotificationRepository {
	return &notificationRepository{db: db}
}

// Create は通知を作成する
func (r *notificationRepository) Create(ctx context.Context, n *entity.Notification) error {
	return r.db.WithContext(ctx).Create(n).Error
}
