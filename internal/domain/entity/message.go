package entity

import "time"

// Message はメッセージのエンティティ
type Message struct {
	ID            int64     `json:"id" gorm:"primaryKey"`
	TransactionID int64     `json:"transaction_id" gorm:"not null"`
	SenderID      int64     `json:"sender_id" gorm:"not null"`
	Content       string    `json:"content" gorm:"type:text;not null"`
	IsRead        bool      `json:"is_read" gorm:"default:false"`
	CreatedAt     time.Time `json:"created_at"`

	// リレーション
	Transaction *Transaction `json:"transaction,omitempty" gorm:"foreignKey:TransactionID"`
	Sender      *User        `json:"sender,omitempty" gorm:"foreignKey:SenderID"`
}

// TableName はテーブル名を指定
func (Message) TableName() string {
	return "messages"
}
