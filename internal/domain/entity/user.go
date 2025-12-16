package entity

import "time"

// UserStatus はユーザーのステータスを表す
type UserStatus string

// KYCStatus は本人確認のステータスを表す
type KYCStatus string

const (
	UserStatusActive    UserStatus = "active"
	UserStatusSuspended UserStatus = "suspended"
	UserStatusBanned    UserStatus = "banned"
	UserStatusDeleted   UserStatus = "deleted"

	KYCStatusUnverified KYCStatus = "unverified"
	KYCStatusPending    KYCStatus = "pending"
	KYCStatusVerified   KYCStatus = "verified"
	KYCStatusRejected   KYCStatus = "rejected"
)

// User はユーザーのエンティティ
type User struct {
	ID          int64      `json:"id" gorm:"primaryKey"`
	FirebaseUID string     `json:"firebase_uid" gorm:"uniqueIndex;size:128;not null"`
	Email       string     `json:"email" gorm:"uniqueIndex;size:255;not null"`
	PhoneNumber *string    `json:"phone_number,omitempty" gorm:"uniqueIndex;size:20"`
	Status      UserStatus `json:"status" gorm:"type:enum('active','suspended','banned','deleted');default:'active'"`
	KYCStatus   KYCStatus  `json:"kyc_status" gorm:"type:enum('unverified','pending','verified','rejected');default:'unverified'"`
	LastLoginAt *time.Time `json:"last_login_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty" gorm:"index"`

	// リレーション
	Profile *UserProfile `json:"profile,omitempty" gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	Wallet  *Wallet      `json:"wallet,omitempty" gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}

// TableName はテーブル名を指定
func (User) TableName() string {
	return "users"
}

// IsActive はユーザーがアクティブかどうかを返す
func (u *User) IsActive() bool {
	return u.Status == UserStatusActive && u.DeletedAt == nil
}

// IsKYCVerified は本人確認が完了しているかを返す
func (u *User) IsKYCVerified() bool {
	return u.KYCStatus == KYCStatusVerified
}

// CanSell は出品可能かどうかを返す
func (u *User) CanSell() bool {
	return u.IsActive() && u.Profile != nil && u.Profile.IsSellerVerified
}
