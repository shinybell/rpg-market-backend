package entity

import (
	"errors"
	"time"
)

// Wallet はウォレットのエンティティ
type Wallet struct {
	UserID    int64     `json:"user_id" gorm:"primaryKey"`
	Balance   float64   `json:"balance" gorm:"type:decimal(10,2);default:0.00;check:balance >= 0"`
	Points    float64   `json:"points" gorm:"type:decimal(10,2);default:0.00;check:points >= 0"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// リレーション
	User *User `json:"-" gorm:"foreignKey:UserID"`
}

// TableName はテーブル名を指定
func (Wallet) TableName() string {
	return "wallets"
}

var (
	ErrInsufficientBalance = errors.New("insufficient balance")
	ErrInsufficientPoints  = errors.New("insufficient points")
	ErrInvalidAmount       = errors.New("invalid amount")
)

// Deposit は残高に入金する
func (w *Wallet) Deposit(amount float64) error {
	if amount <= 0 {
		return ErrInvalidAmount
	}
	w.Balance += amount
	return nil
}

// Withdraw は残高から出金する
func (w *Wallet) Withdraw(amount float64) error {
	if amount <= 0 {
		return ErrInvalidAmount
	}
	if w.Balance < amount {
		return ErrInsufficientBalance
	}
	w.Balance -= amount
	return nil
}

// AddPoints はポイントを追加する
func (w *Wallet) AddPoints(points float64) error {
	if points <= 0 {
		return ErrInvalidAmount
	}
	w.Points += points
	return nil
}

// UsePoints はポイントを使用する
func (w *Wallet) UsePoints(points float64) error {
	if points <= 0 {
		return ErrInvalidAmount
	}
	if w.Points < points {
		return ErrInsufficientPoints
	}
	w.Points -= points
	return nil
}

// CanAfford は指定金額を支払えるかを返す
func (w *Wallet) CanAfford(amount float64) bool {
	return w.Balance >= amount
}

// HasPoints は指定ポイントを持っているかを返す
func (w *Wallet) HasPoints(points float64) bool {
	return w.Points >= points
}

// GetTotalValue は残高とポイントの合計を返す
func (w *Wallet) GetTotalValue() float64 {
	return w.Balance + w.Points
}
