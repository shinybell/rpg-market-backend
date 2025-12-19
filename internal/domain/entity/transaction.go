package entity

import "time"

// PaymentStatus は支払いステータス
type PaymentStatus string

// TransactionStatus は取引ステータス
type TransactionStatus string

const (
	PaymentStatusPending   PaymentStatus = "pending"
	PaymentStatusCaptured  PaymentStatus = "captured"
	PaymentStatusFailed    PaymentStatus = "failed"
	PaymentStatusRefunded  PaymentStatus = "refunded"
	PaymentStatusCancelled PaymentStatus = "cancelled"

	TransactionStatusAwaitingPay  TransactionStatus = "awaiting_pay"
	TransactionStatusAwaitingShip TransactionStatus = "awaiting_ship"
	TransactionStatusShipped      TransactionStatus = "shipped"
	TransactionStatusDelivered    TransactionStatus = "delivered"
	TransactionStatusCompleted    TransactionStatus = "completed"
	TransactionStatusCancelled    TransactionStatus = "cancelled"
)

// Transaction は取引のエンティティ
type Transaction struct {
	ID                   int64             `json:"id" gorm:"primaryKey"`
	ItemID               int64             `json:"item_id" gorm:"not null"`
	BuyerID              int64             `json:"buyer_id" gorm:"not null"`
	SellerID             int64             `json:"seller_id" gorm:"not null"`
	Price                int64             `json:"price" gorm:"not null;check:price >= 0"`
	FeeAmount            int64             `json:"fee_amount" gorm:"default:0"`
	ProfitAmount         int64             `json:"profit_amount" gorm:"not null"`
	PaymentStatus        PaymentStatus     `json:"payment_status" gorm:"type:enum('pending','captured','failed','refunded','cancelled');default:'pending'"`
	TransactionStatus    TransactionStatus `json:"transaction_status" gorm:"type:enum('awaiting_pay','awaiting_ship','shipped','delivered','completed','cancelled');default:'awaiting_pay'"`
	PaymentMethod        string            `json:"payment_method" gorm:"size:50"`
	PaymentTransactionID string            `json:"payment_transaction_id" gorm:"size:255"`
	CreatedAt            time.Time         `json:"created_at"`
	CompletedAt          *time.Time        `json:"completed_at,omitempty"`
	CancelledAt          *time.Time        `json:"cancelled_at,omitempty"`

	// リレーション
	Item   *Item `json:"item,omitempty" gorm:"foreignKey:ItemID"`
	Buyer  *User `json:"buyer,omitempty" gorm:"foreignKey:BuyerID"`
	Seller *User `json:"seller,omitempty" gorm:"foreignKey:SellerID"`
}

// TableName はテーブル名を指定
func (Transaction) TableName() string {
	return "transactions"
}
