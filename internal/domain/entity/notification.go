package entity

import "time"

// Notification は通知のエンティティ
type Notification struct {
	ID        int64      `json:"id" gorm:"primaryKey"`
	UserID    int64      `json:"user_id" gorm:"not null"`
	Type      string     `json:"type" gorm:"size:50;not null"`
	Title     string     `json:"title" gorm:"size:255;not null"`
	Content   string     `json:"content" gorm:"type:text"`
	RelatedID *int64     `json:"related_id,omitempty"`
	IsRead    bool       `json:"is_read" gorm:"default:false"`
	CreatedAt time.Time  `json:"created_at"`
	ReadAt    *time.Time `json:"read_at,omitempty"`

	// リレーション
	User *User `json:"-" gorm:"foreignKey:UserID"`
}

// TableName はテーブル名を指定
func (Notification) TableName() string {
	return "notifications"
}
