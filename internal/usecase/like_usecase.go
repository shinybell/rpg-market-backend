package usecase

import (
	"context"
	"errors"

	"github.com/shinybell/rpg-market-backend/internal/domain/entity"
	"github.com/shinybell/rpg-market-backend/internal/domain/repository"
)

var (
	ErrLikeAlreadyExists = errors.New("like already exists")
	ErrLikeNotFound      = errors.New("like not found")
)

type LikeUseCase struct {
	likeRepo repository.LikeRepository
	itemRepo repository.ItemRepository
}

// NewLikeUseCase はLikeUseCaseを生成する
func NewLikeUseCase(likeRepo repository.LikeRepository, itemRepo repository.ItemRepository) *LikeUseCase {
	return &LikeUseCase{
		likeRepo: likeRepo,
		itemRepo: itemRepo,
	}
}

// AddLike はいいねを追加する
func (uc *LikeUseCase) AddLike(ctx context.Context, userID, itemID int64) error {
	// アイテム存在確認
	item, err := uc.itemRepo.FindByID(ctx, itemID)
	if err != nil {
		return err
	}
	if item == nil {
		return ErrItemNotFound
	}

	// すでにいいね済みか確認
	exists, err := uc.likeRepo.ExistsByUserAndItem(ctx, userID, itemID)
	if err != nil {
		return err
	}
	if exists {
		return ErrLikeAlreadyExists
	}

	// いいね作成
	like := &entity.Like{
		UserID: userID,
		ItemID: itemID,
	}
	if err := uc.likeRepo.Create(ctx, like); err != nil {
		return err
	}

	// アイテムのいいね数更新
	item.IncrementLikesCount()
	return uc.itemRepo.Update(ctx, item)
}

// RemoveLike はいいねを削除する
func (uc *LikeUseCase) RemoveLike(ctx context.Context, userID, itemID int64) error {
	// いいね存在確認
	exists, err := uc.likeRepo.ExistsByUserAndItem(ctx, userID, itemID)
	if err != nil {
		return err
	}
	if !exists {
		return ErrLikeNotFound
	}

	// いいね削除
	if err := uc.likeRepo.Delete(ctx, userID, itemID); err != nil {
		return err
	}

	// アイテムのいいね数更新
	item, err := uc.itemRepo.FindByID(ctx, itemID)
	if err != nil {
		return err
	}
	if item != nil {
		item.DecrementLikesCount()
		return uc.itemRepo.Update(ctx, item)
	}
	return nil
}
