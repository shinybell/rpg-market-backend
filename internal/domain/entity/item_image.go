// rpg-market-backend/internal/domain/entity/item_image.go
package entity

import "time"

// ItemImage はアイテム画像のエンティティ
type ItemImage struct {
	ID           int64     `json:"id" gorm:"primaryKey"`
	ItemID       int64     `json:"item_id" gorm:"not null"`
	ImageURL     string    `json:"image_url" gorm:"size:500;not null"`
	DisplayOrder int       `json:"display_order" gorm:"default:0"`
	CreatedAt    time.Time `json:"created_at"`

	// リレーション
	Item *Item `json:"-" gorm:"foreignKey:ItemID"`
}

// TableName はテーブル名を指定
func (ItemImage) TableName() string {
	return "item_images"
}
