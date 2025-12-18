package usecase

import (
	"context"
	"errors"
	"log"

	"github.com/shinybell/rpg-market-backend/internal/domain/entity"
	"github.com/shinybell/rpg-market-backend/internal/domain/repository"
)

var (
	ErrItemNotFound   = errors.New("item not found")
	ErrItemNotForSale = errors.New("item not for sale")
)

type ItemUseCase struct {
	itemRepo repository.ItemRepository
}

// NewItemUseCase はItemUseCaseを生成する
func NewItemUseCase(itemRepo repository.ItemRepository) *ItemUseCase {
	return &ItemUseCase{
		itemRepo: itemRepo,
	}
}

// CreateItem はアイテムを作成する
func (uc *ItemUseCase) CreateItem(ctx context.Context, item *entity.Item) error {
	// バリデーション
	if err := item.ValidateName(); err != nil {
		return err
	}
	if err := item.ValidateDescription(); err != nil {
		return err
	}
	if err := item.ValidatePrice(); err != nil {
		return err
	}

	// 作成
	if err := uc.itemRepo.Create(ctx, item); err != nil {
		log.Printf("Failed to create item: %v", err)
		return err
	}

	log.Printf("Item created: %s (ID: %d, SellerID: %d)", item.Name, item.ID, item.SellerID)
	return nil
}

// GetItemByID はアイテムを取得し、閲覧数を増やす
func (uc *ItemUseCase) GetItemByID(ctx context.Context, itemID int64) (*entity.Item, error) {
	item, err := uc.itemRepo.FindByID(ctx, itemID)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, ErrItemNotFound
	}

	// 閲覧数を増やす（非同期で行う）
	go func() {
		if err := uc.itemRepo.IncrementViewCount(context.Background(), itemID); err != nil {
			log.Printf("Failed to increment view count for item %d: %v", itemID, err)
		}
	}()

	return item, nil
}

// GetItemsBySeller は出品者のアイテム一覧を取得する
func (uc *ItemUseCase) GetItemsBySeller(ctx context.Context, sellerID int64, limit, offset int) ([]*entity.Item, error) {
	return uc.itemRepo.FindBySeller(ctx, sellerID, limit, offset)
}

// GetItemsByCategory はカテゴリ別のアイテム一覧を取得する
func (uc *ItemUseCase) GetItemsByCategory(ctx context.Context, categoryID int64, limit, offset int) ([]*entity.Item, error) {
	return uc.itemRepo.FindByCategory(ctx, categoryID, limit, offset)
}

// GetItemsOnSale は販売中のアイテム一覧を取得する
func (uc *ItemUseCase) GetItemsOnSale(ctx context.Context, limit, offset int) ([]*entity.Item, error) {
	return uc.itemRepo.FindByStatus(ctx, entity.ItemStatusOnSale, limit, offset)
}

// SearchItems はキーワード検索でアイテムを取得する
func (uc *ItemUseCase) SearchItems(ctx context.Context, keyword string, limit, offset int) ([]*entity.Item, error) {
	return uc.itemRepo.Search(ctx, keyword, limit, offset)
}

// UpdateItem はアイテムを更新する（出品者のみ）
func (uc *ItemUseCase) UpdateItem(ctx context.Context, item *entity.Item, userID int64) error {
	// 既存のアイテムを取得
	existingItem, err := uc.itemRepo.FindByID(ctx, item.ID)
	if err != nil {
		return err
	}
	if existingItem == nil {
		return ErrItemNotFound
	}

	// 権限チェック
	if !existingItem.CanEdit(userID) {
		return entity.ErrUnauthorizedEdit
	}

	// バリデーション
	if err := item.ValidateName(); err != nil {
		return err
	}
	if err := item.ValidateDescription(); err != nil {
		return err
	}
	if err := item.ValidatePrice(); err != nil {
		return err
	}

	// 更新
	if err := uc.itemRepo.Update(ctx, item); err != nil {
		log.Printf("Failed to update item: %v", err)
		return err
	}

	log.Printf("Item updated: %s (ID: %d)", item.Name, item.ID)
	return nil
}

// DeleteItem はアイテムを削除する（出品者のみ）
func (uc *ItemUseCase) DeleteItem(ctx context.Context, itemID, userID int64) error {
	// 既存のアイテムを取得
	item, err := uc.itemRepo.FindByID(ctx, itemID)
	if err != nil {
		return err
	}
	if item == nil {
		return ErrItemNotFound
	}

	// 権限チェック
	if !item.CanEdit(userID) {
		return entity.ErrUnauthorizedEdit
	}

	// 削除
	if err := uc.itemRepo.Delete(ctx, itemID); err != nil {
		log.Printf("Failed to delete item: %v", err)
		return err
	}

	log.Printf("Item deleted: ID %d", itemID)
	return nil
}

// PurchaseItem はアイテムを購入する
func (uc *ItemUseCase) PurchaseItem(ctx context.Context, itemID, buyerID int64) error {
	// アイテムを取得
	item, err := uc.itemRepo.FindByID(ctx, itemID)
	if err != nil {
		return err
	}
	if item == nil {
		return ErrItemNotFound
	}

	// 購入可能かチェック
	if !item.IsAvailable() {
		return ErrItemNotForSale
	}

	// 在庫を減らす
	if err := item.DecrementStock(); err != nil {
		return err
	}

	// 在庫を更新
	if err := uc.itemRepo.Update(ctx, item); err != nil {
		log.Printf("Failed to update item stock: %v", err)
		return err
	}

	log.Printf("Item purchased: ID %d, Buyer: %d, Remaining stock: %d", itemID, buyerID, item.Stock)
	return nil
}
