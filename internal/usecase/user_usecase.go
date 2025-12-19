package usecase

import (
	"context"
	"errors"
	"log"

	"github.com/shinybell/rpg-market-backend/internal/domain/entity"
	"github.com/shinybell/rpg-market-backend/internal/domain/repository"
	"github.com/shinybell/rpg-market-backend/internal/infrastructure/auth"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrInvalidNickname   = errors.New("invalid nickname")
	ErrUserAlreadyExists = errors.New("user already exists")
)

type UserUseCase struct {
	userRepo repository.UserRepository
}

// NewUserUseCase はUserUseCaseを生成する
func NewUserUseCase(userRepo repository.UserRepository) *UserUseCase {
	return &UserUseCase{
		userRepo: userRepo,
	}
}

// LoginOrRegister はFirebase認証後にユーザーを登録またはログインする
func (uc *UserUseCase) LoginOrRegister(ctx context.Context, firebaseUID, email, nickname string) (*entity.User, error) {
	// ニックネームのバリデーション
	if len(nickname) < 2 || len(nickname) > 50 {
		return nil, ErrInvalidNickname
	}

	// 既存ユーザーを検索
	user, err := uc.userRepo.FindByFirebaseUID(ctx, firebaseUID)
	if err != nil {
		return nil, err
	}

	// ユーザーが存在する場合
	if user != nil {
		// 削除済みユーザーのチェック
		if user.DeletedAt != nil {
			return nil, errors.New("このアカウントは削除されています。復活を希望する場合はサポートにお問い合わせください")
		}

		// 最終ログイン時刻を更新
		if err := uc.userRepo.UpdateLastLogin(ctx, user.ID); err != nil {
			log.Printf("Failed to update last login: %v", err)
		}
		// Set custom claims to indicate user is registered
		client := auth.GetClient()
		if client != nil {
			err = client.GetAuthClient().SetCustomUserClaims(ctx, firebaseUID, map[string]interface{}{
				"registered": true,
			})
			if err != nil {
				log.Printf("Failed to set custom claims for user %s: %v", firebaseUID, err)
				// Don't fail the login, just log the error
			}
		}
		log.Printf("User logged in: %s (ID: %d)", email, user.ID)
		return user, nil
	}

	// 新規ユーザー作成
	newUser := &entity.User{
		FirebaseUID: firebaseUID,
		Email:       email,
		Status:      entity.UserStatusActive,
		KYCStatus:   entity.KYCStatusUnverified,
		Profile: &entity.UserProfile{
			Nickname: nickname,
		},
	}

	if err := uc.userRepo.Create(ctx, newUser); err != nil {
		return nil, err
	}

	// Set custom claims to indicate user is registered
	client := auth.GetClient()
	if client != nil {
		err = client.GetAuthClient().SetCustomUserClaims(ctx, firebaseUID, map[string]interface{}{
			"registered": true,
		})
		if err != nil {
			log.Printf("Failed to set custom claims for user %s: %v", firebaseUID, err)
			// Don't fail the registration, just log the error
		}
	}

	log.Printf("New user registered: %s (ID: %d, Firebase UID: %s)", email, newUser.ID, firebaseUID)
	return newUser, nil
}

// GetUserByFirebaseUID はFirebase UIDでユーザー情報を取得する
func (uc *UserUseCase) GetUserByFirebaseUID(ctx context.Context, firebaseUID string) (*entity.User, error) {
	user, err := uc.userRepo.FindByFirebaseUID(ctx, firebaseUID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}
	// 削除済みユーザーのチェック
	if user.DeletedAt != nil {
		return nil, errors.New("このアカウントは削除されています")
	}
	return user, nil
}

// GetUserByID はIDでユーザー情報を取得する
func (uc *UserUseCase) GetUserByID(ctx context.Context, id int64) (*entity.User, error) {
	user, err := uc.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}
	return user, nil
}

// UpdateProfile はプロフィールを更新する
func (uc *UserUseCase) UpdateProfile(ctx context.Context, firebaseUID string, nickname *string, bio *string, avatarURL *string) (*entity.User, error) {
	// ユーザー取得
	user, err := uc.userRepo.FindByFirebaseUID(ctx, firebaseUID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	// プロフィール更新
	if user.Profile != nil {
		if nickname != nil {
			if len(*nickname) < 2 || len(*nickname) > 50 {
				return nil, ErrInvalidNickname
			}
			user.Profile.UpdateNickname(*nickname)
		}
		if bio != nil {
			user.Profile.UpdateBio(*bio)
		}
		if avatarURL != nil {
			user.Profile.UpdateAvatar(*avatarURL)
		}

		if err := uc.userRepo.UpdateProfile(ctx, user.Profile); err != nil {
			return nil, err
		}
	}

	// 最新の情報を取得して返す
	return uc.userRepo.FindByFirebaseUID(ctx, firebaseUID)
}

// DeleteUser はユーザーを論理削除する
func (uc *UserUseCase) DeleteUser(ctx context.Context, firebaseUID string) error {
	user, err := uc.userRepo.FindByFirebaseUID(ctx, firebaseUID)
	if err != nil {
		return err
	}
	if user == nil {
		return ErrUserNotFound
	}

	return uc.userRepo.Delete(ctx, user.ID)
}
