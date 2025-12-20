package mysql

import (
	"context"

	"gorm.io/gorm"

	"github.com/shinybell/rpg-market-backend/internal/domain/entity"
	"github.com/shinybell/rpg-market-backend/internal/domain/repository"
)

type walletRepository struct {
	db *gorm.DB
}

// NewWalletRepository はWalletRepositoryの実装を返す
func NewWalletRepository(db *gorm.DB) repository.WalletRepository {
	return &walletRepository{db: db}
}

// FindByUserID はユーザーIDでウォレットを検索する
func (r *walletRepository) FindByUserID(ctx context.Context, userID int64) (*entity.Wallet, error) {
	var wallet entity.Wallet
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&wallet).Error
	if err != nil {
		return nil, err
	}
	return &wallet, nil
}

// Update はウォレットを更新する
func (r *walletRepository) Update(ctx context.Context, wallet *entity.Wallet) error {
	return r.db.WithContext(ctx).Save(wallet).Error
}

// CreateTransaction はウォレット取引履歴を作成する
func (r *walletRepository) CreateTransaction(ctx context.Context, tx *entity.WalletTransaction) error {
	return r.db.WithContext(ctx).Create(tx).Error
}

// GetTransactionHistory はウォレット取引履歴を取得する
func (r *walletRepository) GetTransactionHistory(ctx context.Context, userID int64, limit, offset int) ([]*entity.WalletTransaction, error) {
	var transactions []*entity.WalletTransaction
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&transactions).Error
	return transactions, err
}
