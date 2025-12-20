// internal/domain/entity/item.go
package entity

import (
	"errors"
	"time"
)

// Item関連のエラー
var (
	ErrOutOfStock          = errors.New("out of stock")
	ErrInvalidPrice        = errors.New("invalid price: must be non-negative")
	ErrPriceTooHigh        = errors.New("price too high: maximum is 999999999")
	ErrInvalidItemName     = errors.New("invalid item name: must be 1-255 characters")
	ErrDescriptionTooShort = errors.New("description too short: minimum 10 characters")
	ErrUnauthorizedEdit    = errors.New("unauthorized to edit this item")
)

// ItemCondition はアイテムの状態を表す
type ItemCondition string

// ShippingPayer は送料負担者を表す
type ShippingPayer string

// ShippingDays は発送日数を表す
type ShippingDays string

// ItemStatus はアイテムの販売状態を表す
type ItemStatus string

const (
	ItemConditionNew        ItemCondition = "new"
	ItemConditionLikeNew    ItemCondition = "like_new"
	ItemConditionVeryGood   ItemCondition = "very_good"
	ItemConditionGood       ItemCondition = "good"
	ItemConditionAcceptable ItemCondition = "acceptable"

	ShippingPayerBuyer  ShippingPayer = "buyer"
	ShippingPayerSeller ShippingPayer = "seller"

	ShippingDays1to2 ShippingDays = "1-2"
	ShippingDays2to3 ShippingDays = "2-3"
	ShippingDays4to7 ShippingDays = "4-7"

	ItemStatusDraft     ItemStatus = "draft"
	ItemStatusOnSale    ItemStatus = "on_sale"
	ItemStatusTrading   ItemStatus = "trading"
	ItemStatusSoldOut   ItemStatus = "sold_out"
	ItemStatusSuspended ItemStatus = "suspended"
)

// Item はアイテムのエンティティ
type Item struct {
	ID               int64         `json:"id" gorm:"primaryKey"`
	SellerID         int64         `json:"seller_id" gorm:"not null"`
	CategoryID       int64         `json:"category_id" gorm:"not null"`
	BrandID          *int64        `json:"brand_id,omitempty"`
	Name             string        `json:"name" gorm:"size:255;not null"`
	Description      string        `json:"description" gorm:"type:text;not null"`
	RPGName          *string       `json:"rpg_name,omitempty" gorm:"size:255"`
	RPGDescription   *string       `json:"rpg_description,omitempty" gorm:"type:text"`
	Price            int64         `json:"price" gorm:"type:bigint;not null"`
	Stock            int           `json:"stock" gorm:"default:1;not null"`
	Condition        ItemCondition `json:"condition" gorm:"type:enum('new','like_new','very_good','good','acceptable');not null"`
	ShippingPayer    ShippingPayer `json:"shipping_payer" gorm:"type:enum('buyer','seller');default:'seller'"`
	ShippingMethodID *int          `json:"shipping_method_id,omitempty"`
	ShippingDays     ShippingDays  `json:"shipping_days" gorm:"type:enum('1-2','2-3','4-7');default:'2-3'"`
	PrefectureID     *int          `json:"prefecture_id,omitempty"`
	Status           ItemStatus    `json:"status" gorm:"type:enum('draft','on_sale','trading','sold_out','suspended');default:'on_sale'"`
	LikesCount       int           `json:"likes_count" gorm:"default:0"`
	CommentsCount    int           `json:"comments_count" gorm:"default:0"`
	ViewCount        int           `json:"view_count" gorm:"default:0"`
	CreatedAt        time.Time     `json:"created_at"`
	UpdatedAt        time.Time     `json:"updated_at"`
	DeletedAt        *time.Time    `json:"deleted_at,omitempty" gorm:"index"`

	// リレーション
	Seller *User       `json:"seller,omitempty" gorm:"foreignKey:SellerID"`
	Images []ItemImage `json:"images,omitempty" gorm:"foreignKey:ItemID"`
}

// TableName はテーブル名を指定
func (Item) TableName() string {
	return "items"
}

// IsAvailable はアイテムが購入可能かどうかを返す
func (i *Item) IsAvailable() bool {
	return i.Status == ItemStatusOnSale && i.Stock > 0 && i.DeletedAt == nil
}

// CanEdit は出品者が編集可能かどうかを返す
func (i *Item) CanEdit(userID int64) bool {
	return i.SellerID == userID && i.Status != ItemStatusSoldOut
}

// DecrementStock は在庫を1減らす
func (i *Item) DecrementStock() error {
	if i.Stock <= 0 {
		return ErrOutOfStock
	}
	i.Stock--
	if i.Stock == 0 {
		i.Status = ItemStatusSoldOut
	}
	return nil
}

// IncrementViewCount は閲覧数を増やす
func (i *Item) IncrementViewCount() {
	i.ViewCount++
}

// IncrementLikesCount はいいね数を増やす
func (i *Item) IncrementLikesCount() {
	i.LikesCount++
}

// DecrementLikesCount はいいね数を減らす
func (i *Item) DecrementLikesCount() {
	if i.LikesCount > 0 {
		i.LikesCount--
	}
}

// ValidatePrice は価格が有効かどうかを検証
func (i *Item) ValidatePrice() error {
	if i.Price < 0 {
		return ErrInvalidPrice
	}
	if i.Price > 999999999 {
		return ErrPriceTooHigh
	}
	return nil
}

// ValidateName は商品名が有効かどうかを検証
func (i *Item) ValidateName() error {
	if len(i.Name) < 1 || len(i.Name) > 255 {
		return ErrInvalidItemName
	}
	return nil
}

// ValidateDescription は説明文が有効かどうかを検証
func (i *Item) ValidateDescription() error {
	if len(i.Description) < 10 {
		return ErrDescriptionTooShort
	}
	return nil
}
