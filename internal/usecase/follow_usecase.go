package usecase

import (
	"context"
	"errors"

	"github.com/shinybell/rpg-market-backend/internal/domain/entity"
	"github.com/shinybell/rpg-market-backend/internal/domain/repository"
)

var (
	ErrFollowAlreadyExists = errors.New("follow already exists")
	ErrFollowNotFound      = errors.New("follow not found")
	ErrSelfFollow          = errors.New("cannot follow yourself")
)

type FollowUseCase struct {
	followRepo repository.FollowRepository
	userRepo   repository.UserRepository
}

// NewFollowUseCase はFollowUseCaseを生成する
func NewFollowUseCase(followRepo repository.FollowRepository, userRepo repository.UserRepository) *FollowUseCase {
	return &FollowUseCase{
		followRepo: followRepo,
		userRepo:   userRepo,
	}
}

// AddFollow はフォローを追加する
func (uc *FollowUseCase) AddFollow(ctx context.Context, followerID, followeeID int64) error {
	if followerID == followeeID {
		return ErrSelfFollow
	}

	// ユーザー存在確認
	follower, err := uc.userRepo.FindByID(ctx, followerID)
	if err != nil {
		return err
	}
	if follower == nil {
		return errors.New("follower not found")
	}
	followee, err := uc.userRepo.FindByID(ctx, followeeID)
	if err != nil {
		return err
	}
	if followee == nil {
		return errors.New("followee not found")
	}

	// すでにフォロー済みか確認
	exists, err := uc.followRepo.ExistsByFollowerAndFollowee(ctx, followerID, followeeID)
	if err != nil {
		return err
	}
	if exists {
		return ErrFollowAlreadyExists
	}

	// フォロー作成
	follow := &entity.Follow{
		FollowerID: followerID,
		FolloweeID: followeeID,
	}
	if err := uc.followRepo.Create(ctx, follow); err != nil {
		return err
	}

	// プロフィールのカウント更新
	if follower.Profile != nil {
		follower.Profile.IncrementFollowing()
		if err := uc.userRepo.UpdateProfile(ctx, follower.Profile); err != nil {
			return err
		}
	}
	if followee.Profile != nil {
		followee.Profile.IncrementFollowers()
		if err := uc.userRepo.UpdateProfile(ctx, followee.Profile); err != nil {
			return err
		}
	}
	return nil
}

// RemoveFollow はフォローを削除する
func (uc *FollowUseCase) RemoveFollow(ctx context.Context, followerID, followeeID int64) error {
	// フォロー存在確認
	exists, err := uc.followRepo.ExistsByFollowerAndFollowee(ctx, followerID, followeeID)
	if err != nil {
		return err
	}
	if !exists {
		return ErrFollowNotFound
	}

	// フォロー削除
	if err := uc.followRepo.Delete(ctx, followerID, followeeID); err != nil {
		return err
	}

	// プロフィールのカウント更新
	follower, _ := uc.userRepo.FindByID(ctx, followerID)
	followee, _ := uc.userRepo.FindByID(ctx, followeeID)
	if follower != nil && follower.Profile != nil {
		follower.Profile.DecrementFollowing()
		if err := uc.userRepo.UpdateProfile(ctx, follower.Profile); err != nil {
			return err
		}
	}
	if followee != nil && followee.Profile != nil {
		followee.Profile.DecrementFollowers()
		if err := uc.userRepo.UpdateProfile(ctx, followee.Profile); err != nil {
			return err
		}
	}
	return nil
}
