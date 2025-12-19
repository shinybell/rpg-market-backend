package repository

import (
	"context"

	"github.com/shinybell/rpg-market-backend/internal/domain/entity"
)

// TransactionRepository は取引のリポジトリインターフェース
type TransactionRepository interface {
	// Create は取引を作成する
	Create(ctx context.Context, tx *entity.Transaction) error

	// FindByID はIDで取引を検索する
	FindByID(ctx context.Context, id int64) (*entity.Transaction, error)

	// FindByItemID はアイテムIDで取引を検索する
	FindByItemID(ctx context.Context, itemID int64) (*entity.Transaction, error)

	// FindByBuyerID は購入者IDで取引一覧を検索する
	FindByBuyerID(ctx context.Context, buyerID int64) ([]*entity.Transaction, error)

	// Update は取引を更新する
	Update(ctx context.Context, tx *entity.Transaction) error
}
