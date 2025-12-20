package repository

import "github.com/shinybell/rpg-market-backend/internal/domain/entity"

// MessageRepository はメッセージのリポジトリインターフェース
type MessageRepository interface {
	Create(message *entity.Message) error
	GetByTransactionID(transactionID int64) ([]*entity.Message, error)
	MarkAsRead(messageID int64) error
	GetUnreadCount(userID int64) (int, error)
}
