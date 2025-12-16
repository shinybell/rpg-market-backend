package entity

import "time"

// UserProfile はユーザープロフィールのエンティティ
type UserProfile struct {
	UserID           int64     `json:"user_id" gorm:"primaryKey"`
	Nickname         string    `json:"nickname" gorm:"size:50;not null"`
	Bio              *string   `json:"bio,omitempty" gorm:"type:text"`
	AvatarURL        *string   `json:"avatar_url,omitempty" gorm:"size:500"`
	IsSellerVerified bool      `json:"is_seller_verified" gorm:"default:false"`
	FollowersCount   int       `json:"followers_count" gorm:"default:0"`
	FollowingCount   int       `json:"following_count" gorm:"default:0"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`

	// リレーション
	User *User `json:"-" gorm:"foreignKey:UserID"`
}

// TableName はテーブル名を指定
func (UserProfile) TableName() string {
	return "user_profiles"
}

// HasAvatar はアバター画像が設定されているかを返す
func (up *UserProfile) HasAvatar() bool {
	return up.AvatarURL != nil && *up.AvatarURL != ""
}

// UpdateNickname はニックネームを更新
func (up *UserProfile) UpdateNickname(nickname string) {
	up.Nickname = nickname
}

// UpdateBio は自己紹介を更新
func (up *UserProfile) UpdateBio(bio string) {
	up.Bio = &bio
}

// UpdateAvatar はアバター画像URLを更新
func (up *UserProfile) UpdateAvatar(url string) {
	up.AvatarURL = &url
}

// IncrementFollowers はフォロワー数を増やす
func (up *UserProfile) IncrementFollowers() {
	up.FollowersCount++
}

// DecrementFollowers はフォロワー数を減らす
func (up *UserProfile) DecrementFollowers() {
	if up.FollowersCount > 0 {
		up.FollowersCount--
	}
}

// IncrementFollowing はフォロー数を増やす
func (up *UserProfile) IncrementFollowing() {
	up.FollowingCount++
}

// DecrementFollowing はフォロー数を減らす
func (up *UserProfile) DecrementFollowing() {
	if up.FollowingCount > 0 {
		up.FollowingCount--
	}
}
