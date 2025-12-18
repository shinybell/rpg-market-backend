package entity

import "time"

// Like はいいねのエンティティ
type Like struct {
	ID        int64     `json:"id" gorm:"primaryKey"`
	UserID    int64     `json:"user_id" gorm:"not null"`
	ItemID    int64     `json:"item_id" gorm:"not null"`
	CreatedAt time.Time `json:"created_at"`
}

// TableName はテーブル名を指定
func (Like) TableName() string {
	return "item_likes"
}
