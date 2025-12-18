package mysql

import (
	"context"
	"time"

	"gorm.io/gorm"

	"github.com/shinybell/rpg-market-backend/internal/domain/entity"
	"github.com/shinybell/rpg-market-backend/internal/domain/repository"
)

type itemRepository struct {
	db *gorm.DB
}

// NewItemRepository はItemRepositoryの実装を返す
func NewItemRepository(db *gorm.DB) repository.ItemRepository {
	return &itemRepository{db: db}
}

// Create はアイテムを作成する
func (r *itemRepository) Create(ctx context.Context, item *entity.Item) error {
	tx := r.db.WithContext(ctx).Begin()
	if err := tx.Create(item).Error; err != nil {
		tx.Rollback()
		return err
	}
	if len(item.Images) > 0 {
		if err := tx.Model(item).Association("Images").Append(item.Images); err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit().Error
}

// FindByID はIDでアイテムを検索する
func (r *itemRepository) FindByID(ctx context.Context, id int64) (*entity.Item, error) {
	var item entity.Item
	err := r.db.WithContext(ctx).
		Preload("Seller").
		Preload("Seller.Profile").
		Preload("Images", func(db *gorm.DB) *gorm.DB {
			return db.Order("display_order ASC")
		}).
		Where("id = ? AND deleted_at IS NULL", id).
		First(&item).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}

	return &item, nil
}

// FindBySeller は出品者IDでアイテムを検索する
func (r *itemRepository) FindBySeller(ctx context.Context, sellerID int64, limit, offset int) ([]*entity.Item, error) {
	var items []*entity.Item
	err := r.db.WithContext(ctx).
		Preload("Images", func(db *gorm.DB) *gorm.DB {
			return db.Order("display_order ASC")
		}).
		Where("seller_id = ? AND deleted_at IS NULL", sellerID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&items).Error

	if err != nil {
		return nil, err
	}

	return items, nil
}

// FindByCategory はカテゴリIDでアイテムを検索する
func (r *itemRepository) FindByCategory(ctx context.Context, categoryID int64, limit, offset int) ([]*entity.Item, error) {
	var items []*entity.Item
	err := r.db.WithContext(ctx).
		Preload("Seller").
		Preload("Images", func(db *gorm.DB) *gorm.DB {
			return db.Order("display_order ASC")
		}).
		Where("category_id = ? AND status = ? AND deleted_at IS NULL", categoryID, entity.ItemStatusOnSale).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&items).Error

	if err != nil {
		return nil, err
	}

	return items, nil
}

// FindByStatus はステータスでアイテムを検索する
func (r *itemRepository) FindByStatus(ctx context.Context, status entity.ItemStatus, limit, offset int) ([]*entity.Item, error) {
	var items []*entity.Item
	err := r.db.WithContext(ctx).
		Preload("Seller").
		Preload("Images", func(db *gorm.DB) *gorm.DB {
			return db.Order("display_order ASC")
		}).
		Where("status = ? AND deleted_at IS NULL", status).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&items).Error

	if err != nil {
		return nil, err
	}

	return items, nil
}

// Search はキーワードでアイテムを検索する
func (r *itemRepository) Search(ctx context.Context, keyword string, limit, offset int) ([]*entity.Item, error) {
	var items []*entity.Item
	likeKeyword := "%" + keyword + "%"

	err := r.db.WithContext(ctx).
		Preload("Seller").
		Preload("Images", func(db *gorm.DB) *gorm.DB {
			return db.Order("display_order ASC")
		}).
		Where("(name LIKE ? OR description LIKE ?) AND status = ? AND deleted_at IS NULL",
			likeKeyword, likeKeyword, entity.ItemStatusOnSale).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&items).Error

	if err != nil {
		return nil, err
	}

	return items, nil
}

// Update はアイテム情報を更新する
func (r *itemRepository) Update(ctx context.Context, item *entity.Item) error {
	return r.db.WithContext(ctx).
		Model(item).
		Updates(map[string]interface{}{
			"name":               item.Name,
			"description":        item.Description,
			"price":              item.Price,
			"stock":              item.Stock,
			"condition":          item.Condition,
			"shipping_payer":     item.ShippingPayer,
			"shipping_method_id": item.ShippingMethodID,
			"shipping_days":      item.ShippingDays,
			"prefecture_id":      item.PrefectureID,
			"status":             item.Status,
			"updated_at":         time.Now(),
		}).Error
}

// Delete はアイテムを論理削除する
func (r *itemRepository) Delete(ctx context.Context, itemID int64) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&entity.Item{}).
		Where("id = ?", itemID).
		Update("deleted_at", now).Error
}

// IncrementViewCount は閲覧数を増やす
func (r *itemRepository) IncrementViewCount(ctx context.Context, itemID int64) error {
	return r.db.WithContext(ctx).
		Model(&entity.Item{}).
		Where("id = ?", itemID).
		UpdateColumn("view_count", gorm.Expr("view_count + 1")).Error
}
