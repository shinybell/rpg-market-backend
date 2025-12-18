package entity

import "time"

// Follow はフォローのエンティティ
type Follow struct {
	ID         int64     `json:"id" gorm:"primaryKey"`
	FollowerID int64     `json:"follower_id" gorm:"not null"`
	FolloweeID int64     `json:"followee_id" gorm:"not null"`
	CreatedAt  time.Time `json:"created_at"`
}

// TableName はテーブル名を指定
func (Follow) TableName() string {
	return "user_follows"
}
