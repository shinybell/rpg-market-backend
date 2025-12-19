package mysql

import (
	"context"

	"gorm.io/gorm"

	"github.com/shinybell/rpg-market-backend/internal/domain/entity"
	"github.com/shinybell/rpg-market-backend/internal/domain/repository"
)

type transactionRepository struct {
	db *gorm.DB
}

// NewTransactionRepository はTransactionRepositoryの実装を返す
func NewTransactionRepository(db *gorm.DB) repository.TransactionRepository {
	return &transactionRepository{db: db}
}

// Create は取引を作成する
func (r *transactionRepository) Create(ctx context.Context, tx *entity.Transaction) error {
	return r.db.WithContext(ctx).Create(tx).Error
}

// FindByID はIDで取引を検索する
func (r *transactionRepository) FindByID(ctx context.Context, id int64) (*entity.Transaction, error) {
	var tx entity.Transaction
	err := r.db.WithContext(ctx).First(&tx, id).Error
	if err != nil {
		return nil, err
	}
	return &tx, nil
}

// Update は取引を更新する
func (r *transactionRepository) Update(ctx context.Context, tx *entity.Transaction) error {
	return r.db.WithContext(ctx).Save(tx).Error
}
