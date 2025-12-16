package repository

import (
	"context"

	"github.com/shinybell/rpg-market-backend/internal/domain/entity"
)

// UserRepository はユーザーのリポジトリインターフェース
type UserRepository interface {
	// Create はユーザーを作成する
	Create(ctx context.Context, user *entity.User) error

	// FindByID はIDでユーザーを検索する
	FindByID(ctx context.Context, id int64) (*entity.User, error)

	// FindByFirebaseUID はFirebase UIDでユーザーを検索する
	FindByFirebaseUID(ctx context.Context, firebaseUID string) (*entity.User, error)

	// FindByEmail はメールアドレスでユーザーを検索する
	FindByEmail(ctx context.Context, email string) (*entity.User, error)

	// Update はユーザー情報を更新する
	Update(ctx context.Context, user *entity.User) error

	// UpdateLastLogin は最終ログイン時刻を更新する
	UpdateLastLogin(ctx context.Context, userID int64) error

	// Delete はユーザーを論理削除する
	Delete(ctx context.Context, userID int64) error

	// UpdateProfile はプロフィールを更新する
	UpdateProfile(ctx context.Context, profile *entity.UserProfile) error

	// UpdateWallet はウォレットを更新する
	UpdateWallet(ctx context.Context, wallet *entity.Wallet) error
}
