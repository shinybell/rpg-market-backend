package mysql

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"github.com/shinybell/rpg-market-backend/internal/domain/entity"
	"github.com/shinybell/rpg-market-backend/internal/domain/repository"
)

type userRepository struct {
	db *gorm.DB
}

// NewUserRepository はUserRepositoryの実装を返す
func NewUserRepository(db *gorm.DB) repository.UserRepository {
	return &userRepository{db: db}
}

// Create はユーザーを作成する（トランザクション付き）
func (r *userRepository) Create(ctx context.Context, user *entity.User) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// ユーザー作成
		if err := tx.Create(user).Error; err != nil {
			return err
		}

		// プロフィール作成（既に存在する場合はスキップ）
		if user.Profile != nil {
			user.Profile.UserID = user.ID

			// 既存のプロフィールを確認
			var existingProfile entity.UserProfile
			err := tx.Where("user_id = ?", user.ID).First(&existingProfile).Error

			if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}

			// プロフィールが存在しない場合のみ作成
			if errors.Is(err, gorm.ErrRecordNotFound) {
				if err := tx.Create(user.Profile).Error; err != nil {
					return err
				}
			} else {
				// 既存のプロフィールを使用
				user.Profile = &existingProfile
			}
		}

		// ウォレット作成（既に存在する場合はスキップ）
		var existingWallet entity.Wallet
		err := tx.Where("user_id = ?", user.ID).First(&existingWallet).Error

		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		// ウォレットが存在しない場合のみ作成
		if errors.Is(err, gorm.ErrRecordNotFound) {
			wallet := &entity.Wallet{
				UserID:  user.ID,
				Balance: 0.00,
				Points:  0.00,
			}
			if err := tx.Create(wallet).Error; err != nil {
				return err
			}
			user.Wallet = wallet
		} else {
			// 既存のウォレットを使用
			user.Wallet = &existingWallet
		}

		return nil
	})
}

// FindByID はIDでユーザーを検索する
func (r *userRepository) FindByID(ctx context.Context, id int64) (*entity.User, error) {
	var user entity.User
	err := r.db.WithContext(ctx).
		Preload("Profile").
		Preload("Wallet").
		Where("id = ? AND deleted_at IS NULL", id).
		First(&user).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// FindByFirebaseUID はFirebase UIDでユーザーを検索する
func (r *userRepository) FindByFirebaseUID(ctx context.Context, firebaseUID string) (*entity.User, error) {
	var user entity.User
	err := r.db.WithContext(ctx).
		Preload("Profile").
		Preload("Wallet").
		Where("firebase_uid = ? AND deleted_at IS NULL", firebaseUID).
		First(&user).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// FindByEmail はメールアドレスでユーザーを検索する
func (r *userRepository) FindByEmail(ctx context.Context, email string) (*entity.User, error) {
	var user entity.User
	err := r.db.WithContext(ctx).
		Preload("Profile").
		Preload("Wallet").
		Where("email = ? AND deleted_at IS NULL", email).
		First(&user).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// Update はユーザー情報を更新する
func (r *userRepository) Update(ctx context.Context, user *entity.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

// UpdateLastLogin は最終ログイン時刻を更新する
func (r *userRepository) UpdateLastLogin(ctx context.Context, userID int64) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&entity.User{}).
		Where("id = ?", userID).
		Update("last_login_at", now).Error
}

// Delete はユーザーを論理削除する
func (r *userRepository) Delete(ctx context.Context, userID int64) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&entity.User{}).
		Where("id = ?", userID).
		Updates(map[string]interface{}{
			"deleted_at": now,
			"status":     entity.UserStatusDeleted,
		}).Error
}

// UpdateProfile はプロフィールを更新する
func (r *userRepository) UpdateProfile(ctx context.Context, profile *entity.UserProfile) error {
	return r.db.WithContext(ctx).Save(profile).Error
}

// UpdateWallet はウォレットを更新する
func (r *userRepository) UpdateWallet(ctx context.Context, wallet *entity.Wallet) error {
	return r.db.WithContext(ctx).Save(wallet).Error
}
