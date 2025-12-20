package entity

import (
	"time"
)

// WalletTransaction はウォレット取引のエンティティ
type WalletTransaction struct {
	ID              int64     `json:"id" gorm:"primaryKey"`
	UserID          int64     `json:"user_id" gorm:"not null;index:idx_user_id"`
	RelatedID       *int64    `json:"related_id"`
	Type            string    `json:"type" gorm:"type:varchar(50);not null;index:idx_type"`
	Amount          int64     `json:"amount" gorm:"not null"`
	BalanceSnapshot int64     `json:"balance_snapshot" gorm:"not null"`
	Description     string    `json:"description" gorm:"type:varchar(255)"`
	CreatedAt       time.Time `json:"created_at" gorm:"autoCreateTime;index:idx_created_at"`

	// リレーション
	User *User `json:"-" gorm:"foreignKey:UserID"`
}

// TableName はテーブル名を指定
func (WalletTransaction) TableName() string {
	return "wallet_transactions"
}

// 取引タイプの定数
const (
	WalletTransactionTypeSalesDeposit = "sales_deposit"
	WalletTransactionTypePurchase     = "purchase"
	WalletTransactionTypeWithdrawal   = "withdrawal"
	WalletTransactionTypePointGrant   = "point_grant"
	WalletTransactionTypePointExpire  = "point_expire"
	WalletTransactionTypeRefund       = "refund"
	WalletTransactionTypeCharge       = "charge"
)
