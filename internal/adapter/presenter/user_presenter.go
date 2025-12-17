package presenter

import (
	"github.com/shinybell/rpg-market-backend/internal/domain/entity"
)

// UserResponse はユーザーのレスポンス
type UserResponse struct {
	ID          int64                `json:"id"`
	FirebaseUID string               `json:"firebase_uid"`
	Email       string               `json:"email"`
	Status      entity.UserStatus    `json:"status"`
	KYCStatus   entity.KYCStatus     `json:"kyc_status"`
	Profile     *UserProfileResponse `json:"profile,omitempty"`
	Wallet      *WalletResponse      `json:"wallet,omitempty"`
	CreatedAt   string               `json:"created_at"`
	UpdatedAt   string               `json:"updated_at"`
}

// UserProfileResponse はユーザープロフィールのレスポンス
type UserProfileResponse struct {
	Nickname  string `json:"nickname"`
	Bio       string `json:"bio"`
	AvatarURL string `json:"avatar_url"`
}

// WalletResponse はウォレットのレスポンス
type WalletResponse struct {
	Balance float64 `json:"balance"`
	Points  float64 `json:"points"`
}

// ToUserResponse はエンティティをレスポンスに変換する
func ToUserResponse(user *entity.User) UserResponse {
	resp := UserResponse{
		ID:          user.ID,
		FirebaseUID: user.FirebaseUID,
		Email:       user.Email,
		Status:      user.Status,
		KYCStatus:   user.KYCStatus,
		CreatedAt:   user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   user.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	// プロフィール情報
	if user.Profile != nil {
		profile := &UserProfileResponse{
			Nickname: user.Profile.Nickname,
		}
		if user.Profile.Bio != nil {
			profile.Bio = *user.Profile.Bio
		}
		if user.Profile.AvatarURL != nil {
			profile.AvatarURL = *user.Profile.AvatarURL
		}
		resp.Profile = profile
	}

	// ウォレット情報
	if user.Wallet != nil {
		resp.Wallet = &WalletResponse{
			Balance: user.Wallet.Balance,
			Points:  user.Wallet.Points,
		}
	}

	return resp
}
