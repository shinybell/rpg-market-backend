package entity

import (
	"errors"
	"time"
)

// Wallet はウォレットのエンティティ
type Wallet struct {
	UserID    int64     `json:"user_id" gorm:"primaryKey"`
	Balance   int64     `json:"balance" gorm:"type:bigint;default:0;check:balance >= 0"`
	Points    int64     `json:"points" gorm:"type:bigint;default:0;check:points >= 0"`
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
func (w *Wallet) Deposit(amount int64) error {
	if amount <= 0 {
		return ErrInvalidAmount
	}
	w.Balance += amount
	return nil
}

// Withdraw は残高から出金する
func (w *Wallet) Withdraw(amount int64) error {
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
func (w *Wallet) AddPoints(points int64) error {
	if points <= 0 {
		return ErrInvalidAmount
	}
	w.Points += points
	return nil
}

// UsePoints はポイントを使用する
func (w *Wallet) UsePoints(points int64) error {
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
func (w *Wallet) CanAfford(amount int64) bool {
	return w.Balance >= amount
}

// HasPoints は指定ポイントを持っているかを返す
func (w *Wallet) HasPoints(points int64) bool {
	return w.Points >= points
}

// GetTotalValue は残高とポイントの合計を返す
func (w *Wallet) GetTotalValue() int64 {
	return w.Balance + w.Points
}
