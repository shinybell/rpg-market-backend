package entity

import "time"

// Address は配送先のエンティティ
type Address struct {
	ID         int64     `json:"id" gorm:"primaryKey"`
	UserID     int64     `json:"user_id" gorm:"not null"`
	Name       string    `json:"name" gorm:"size:100;not null"`
	PostalCode string    `json:"postal_code" gorm:"size:10;not null"`
	Address    string    `json:"address" gorm:"type:text;not null"`
	Phone      string    `json:"phone" gorm:"size:20;not null"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`

	// リレーション
	User *User `json:"user,omitempty" gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}

// TableName はテーブル名を指定
func (Address) TableName() string {
	return "addresses"
}
