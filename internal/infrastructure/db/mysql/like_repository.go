package mysql

import (
	"context"

	"gorm.io/gorm"

	"github.com/shinybell/rpg-market-backend/internal/domain/entity"
	"github.com/shinybell/rpg-market-backend/internal/domain/repository"
)

type likeRepository struct {
	db *gorm.DB
}

// NewLikeRepository はLikeRepositoryの実装を返す
func NewLikeRepository(db *gorm.DB) repository.LikeRepository {
	return &likeRepository{db: db}
}

// Create はいいねを作成する
func (r *likeRepository) Create(ctx context.Context, like *entity.Like) error {
	return r.db.WithContext(ctx).Create(like).Error
}

// Delete はいいねを削除する
func (r *likeRepository) Delete(ctx context.Context, userID, itemID int64) error {
	return r.db.WithContext(ctx).Where("user_id = ? AND item_id = ?", userID, itemID).Delete(&entity.Like{}).Error
}

// ExistsByUserAndItem は指定されたユーザーとアイテムのいいねが存在するかを確認する
func (r *likeRepository) ExistsByUserAndItem(ctx context.Context, userID, itemID int64) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entity.Like{}).Where("user_id = ? AND item_id = ?", userID, itemID).Count(&count).Error
	return count > 0, err
}
