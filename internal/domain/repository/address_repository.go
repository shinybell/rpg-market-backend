package repository

import (
	"context"

	"github.com/shinybell/rpg-market-backend/internal/domain/entity"
)

// AddressRepository は配送先のリポジトリインターフェース
type AddressRepository interface {
	// Create は配送先を作成する
	Create(ctx context.Context, addr *entity.Address) error

	// FindByID はIDで配送先を検索する
	FindByID(ctx context.Context, id int64) (*entity.Address, error)

	// FindByUserID はユーザーIDで配送先一覧を取得する
	FindByUserID(ctx context.Context, userID int64) ([]*entity.Address, error)

	// Update は配送先を更新する
	Update(ctx context.Context, addr *entity.Address) error

	// Delete は配送先を削除する
	Delete(ctx context.Context, id int64) error
}
