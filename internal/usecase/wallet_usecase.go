package usecase

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/shinybell/rpg-market-backend/internal/domain/entity"
	"github.com/shinybell/rpg-market-backend/internal/domain/repository"
)

// WalletUseCase はウォレットのユースケースインターフェース
type WalletUseCase interface {
	// ChargeBalance はウォレットに残高をチャージする
	ChargeBalance(ctx context.Context, userID int64, amount int64, description string) (*entity.Wallet, error)

	// GetTransactionHistory はウォレット取引履歴を取得する
	GetTransactionHistory(ctx context.Context, userID int64, limit, offset int) ([]*entity.WalletTransaction, error)
}

type walletUseCase struct {
	walletRepo repository.WalletRepository
	db         *gorm.DB
}

// NewWalletUseCase はWalletUseCaseの実装を返す
func NewWalletUseCase(walletRepo repository.WalletRepository, db *gorm.DB) WalletUseCase {
	return &walletUseCase{
		walletRepo: walletRepo,
		db:         db,
	}
}

// ChargeBalance はウォレットに残高をチャージする
func (uc *walletUseCase) ChargeBalance(ctx context.Context, userID int64, amount int64, description string) (*entity.Wallet, error) {
	// バリデーション
	if amount <= 0 {
		return nil, errors.New("charge amount must be positive")
	}
	if amount > 100000 {
		return nil, errors.New("charge amount exceeds maximum limit of 100,000")
	}

	var wallet *entity.Wallet
	var err error

	// トランザクション処理
	err = uc.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. ウォレット取得
		wallet, err = uc.walletRepo.FindByUserID(ctx, userID)
		if err != nil {
			return err
		}

		// 2. 残高更新
		if err := wallet.Deposit(amount); err != nil {
			return err
		}

		// 3. ウォレット保存
		if err := uc.walletRepo.Update(ctx, wallet); err != nil {
			return err
		}

		// 4. 取引履歴作成
		transaction := &entity.WalletTransaction{
			UserID:          userID,
			Type:            entity.WalletTransactionTypeCharge,
			Amount:          amount,
			BalanceSnapshot: wallet.Balance,
			Description:     description,
		}
		if err := uc.walletRepo.CreateTransaction(ctx, transaction); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return wallet, nil
}

// GetTransactionHistory はウォレット取引履歴を取得する
func (uc *walletUseCase) GetTransactionHistory(ctx context.Context, userID int64, limit, offset int) ([]*entity.WalletTransaction, error) {
	return uc.walletRepo.GetTransactionHistory(ctx, userID, limit, offset)
}
