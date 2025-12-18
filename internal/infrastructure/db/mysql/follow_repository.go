package mysql

import (
	"context"

	"gorm.io/gorm"

	"github.com/shinybell/rpg-market-backend/internal/domain/entity"
	"github.com/shinybell/rpg-market-backend/internal/domain/repository"
)

type followRepository struct {
	db *gorm.DB
}

// NewFollowRepository はFollowRepositoryの実装を返す
func NewFollowRepository(db *gorm.DB) repository.FollowRepository {
	return &followRepository{db: db}
}

// Create はフォローを作成する
func (r *followRepository) Create(ctx context.Context, follow *entity.Follow) error {
	return r.db.WithContext(ctx).Create(follow).Error
}

// Delete はフォローを削除する
func (r *followRepository) Delete(ctx context.Context, followerID, followeeID int64) error {
	return r.db.WithContext(ctx).Where("follower_id = ? AND followee_id = ?", followerID, followeeID).Delete(&entity.Follow{}).Error
}

// ExistsByFollowerAndFollowee は指定されたフォロワーとフォロー対象のフォローが存在するかを確認する
func (r *followRepository) ExistsByFollowerAndFollowee(ctx context.Context, followerID, followeeID int64) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entity.Follow{}).Where("follower_id = ? AND followee_id = ?", followerID, followeeID).Count(&count).Error
	return count > 0, err
}
