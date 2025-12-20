package usecase

import (
	"context"
	"errors"

	"github.com/shinybell/rpg-market-backend/internal/domain/entity"
	"github.com/shinybell/rpg-market-backend/internal/domain/repository"
)

// MessageUsecase はメッセージのユースケース
type MessageUsecase interface {
	SendMessage(ctx context.Context, transactionID, senderID int64, content string) (*entity.Message, error)
	GetMessages(ctx context.Context, transactionID, userID int64) ([]*entity.Message, error)
	ValidateAccess(ctx context.Context, transactionID, userID int64) error
	GetUnreadCount(userID int64) (int, error)
	// GetTransactionByIDForUser はトランザクションIDで取引を取得し、ユーザーが関係者か検証する
	GetTransactionByIDForUser(ctx context.Context, txID, userID int64) (*entity.Transaction, error)
}

type messageUsecase struct {
	messageRepo     repository.MessageRepository
	transactionRepo repository.TransactionRepository
}

// NewMessageUsecase はMessageUsecaseを生成
func NewMessageUsecase(
	messageRepo repository.MessageRepository,
	transactionRepo repository.TransactionRepository,
) MessageUsecase {
	return &messageUsecase{
		messageRepo:     messageRepo,
		transactionRepo: transactionRepo,
	}
}

// ValidateAccess は取引へのアクセス権を検証（購入者または出品者のみ）
func (u *messageUsecase) ValidateAccess(ctx context.Context, transactionID, userID int64) error {
	transaction, err := u.transactionRepo.FindByID(ctx, transactionID)
	if err != nil {
		return err
	}

	if transaction.BuyerID != userID && transaction.SellerID != userID {
		return errors.New("access denied: you are not part of this transaction")
	}

	return nil
}

// SendMessage はメッセージを送信
func (u *messageUsecase) SendMessage(ctx context.Context, transactionID, senderID int64, content string) (*entity.Message, error) {
	// アクセス権チェック
	if err := u.ValidateAccess(ctx, transactionID, senderID); err != nil {
		return nil, err
	}

	// メッセージ作成
	message := &entity.Message{
		TransactionID: transactionID,
		SenderID:      senderID,
		Content:       content,
		IsRead:        false,
	}

	if err := u.messageRepo.Create(message); err != nil {
		return nil, err
	}

	// Senderプロフィールを取得して返す
	messages, err := u.messageRepo.GetByTransactionID(transactionID)
	if err != nil {
		return message, nil // メッセージ自体は作成されているのでエラーは無視
	}

	// 作成したメッセージを探す
	for _, msg := range messages {
		if msg.ID == message.ID {
			return msg, nil
		}
	}

	return message, nil
}

// GetMessages は取引に紐づくメッセージ一覧を取得
func (u *messageUsecase) GetMessages(ctx context.Context, transactionID, userID int64) ([]*entity.Message, error) {
	// アクセス権チェック
	if err := u.ValidateAccess(ctx, transactionID, userID); err != nil {
		return nil, err
	}

	return u.messageRepo.GetByTransactionID(transactionID)
}

// GetUnreadCount は未読メッセージ数を取得
func (u *messageUsecase) GetUnreadCount(userID int64) (int, error) {
	return u.messageRepo.GetUnreadCount(userID)
}

// GetTransactionByIDForUser はトランザクションIDで取引を取得し、ユーザーが関係者か検証する
func (u *messageUsecase) GetTransactionByIDForUser(ctx context.Context, txID, userID int64) (*entity.Transaction, error) {
	tx, err := u.transactionRepo.FindByID(ctx, txID)
	if err != nil {
		return nil, errors.New("transaction not found")
	}

	if tx.BuyerID != userID && tx.SellerID != userID {
		return nil, errors.New("access denied: you are not part of this transaction")
	}

	return tx, nil
}
