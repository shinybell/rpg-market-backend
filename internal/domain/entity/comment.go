package entity

import "time"

// Comment はコメントのエンティティ
type Comment struct {
	ID        int64      `json:"id" gorm:"primaryKey"`
	ItemID    int64      `json:"item_id" gorm:"not null"`
	UserID    int64      `json:"user_id" gorm:"not null"`
	Comment   string     `json:"comment" gorm:"type:text;not null"`
	IsDeleted bool       `json:"is_deleted" gorm:"default:false"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty" gorm:"index"`

	// リレーション
	User *User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// TableName はテーブル名を指定
func (Comment) TableName() string {
	return "item_comments"
}
