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
}
