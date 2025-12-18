package repository

import (
	"context"

	"github.com/shinybell/rpg-market-backend/internal/domain/entity"
)

// FollowRepository はフォローのリポジトリインターフェース
type FollowRepository interface {
	// Create はフォローを作成する
	Create(ctx context.Context, follow *entity.Follow) error

	// Delete はフォローを削除する
	Delete(ctx context.Context, followerID, followeeID int64) error

	// ExistsByFollowerAndFollowee は指定されたフォロワーとフォロー対象のフォローが存在するかを確認する
	ExistsByFollowerAndFollowee(ctx context.Context, followerID, followeeID int64) (bool, error)
}
