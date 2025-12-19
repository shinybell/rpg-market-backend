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

	// Update は取引を更新する
	Update(ctx context.Context, tx *entity.Transaction) error
}
