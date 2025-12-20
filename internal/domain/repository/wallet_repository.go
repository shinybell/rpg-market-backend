package repository

import (
	"context"

	"github.com/shinybell/rpg-market-backend/internal/domain/entity"
)

// WalletRepository はウォレットのリポジトリインターフェース
type WalletRepository interface {
	// FindByUserID はユーザーIDでウォレットを検索する
	FindByUserID(ctx context.Context, userID int64) (*entity.Wallet, error)

	// Update はウォレットを更新する
	Update(ctx context.Context, wallet *entity.Wallet) error

	// CreateTransaction はウォレット取引履歴を作成する
	CreateTransaction(ctx context.Context, tx *entity.WalletTransaction) error

	// GetTransactionHistory はウォレット取引履歴を取得する
	GetTransactionHistory(ctx context.Context, userID int64, limit, offset int) ([]*entity.WalletTransaction, error)
}
