package mysql

import (
	"github.com/shinybell/rpg-market-backend/internal/domain/entity"
	"github.com/shinybell/rpg-market-backend/internal/domain/repository"

	"gorm.io/gorm"
)

type messageRepository struct {
	db *gorm.DB
}

// NewMessageRepository はMessageRepositoryを生成
func NewMessageRepository(db *gorm.DB) repository.MessageRepository {
	return &messageRepository{db: db}
}

// Create は新しいメッセージを作成
func (r *messageRepository) Create(message *entity.Message) error {
	return r.db.Create(message).Error
}

// GetByTransactionID は取引IDに紐づくメッセージを取得
func (r *messageRepository) GetByTransactionID(transactionID int64) ([]*entity.Message, error) {
	var messages []*entity.Message
	err := r.db.
		Where("transaction_id = ?", transactionID).
		Order("created_at ASC").
		Preload("Sender").
		Preload("Sender.Profile").
		Find(&messages).Error
	return messages, err
}

// MarkAsRead はメッセージを既読にする
func (r *messageRepository) MarkAsRead(messageID int64) error {
	return r.db.Model(&entity.Message{}).
		Where("id = ?", messageID).
		Update("is_read", true).Error
}

// GetUnreadCount はユーザーの未読メッセージ数を取得
func (r *messageRepository) GetUnreadCount(userID int64) (int, error) {
	var count int64
	err := r.db.Model(&entity.Message{}).
		Joins("JOIN transactions ON messages.transaction_id = transactions.id").
		Where("(transactions.buyer_id = ? OR transactions.seller_id = ?) AND messages.sender_id != ? AND messages.is_read = false", userID, userID, userID).
		Count(&count).Error
	return int(count), err
}
