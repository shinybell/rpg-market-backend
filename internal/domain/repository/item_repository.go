package repository

import (
	"context"

	"github.com/shinybell/rpg-market-backend/internal/domain/entity"
)

// ItemRepository はアイテムのリポジトリインターフェース
type ItemRepository interface {
	// Create はアイテムを作成する
	Create(ctx context.Context, item *entity.Item) error

	// FindByID はIDでアイテムを検索する
	FindByID(ctx context.Context, id int64) (*entity.Item, error)

	// FindBySeller は出品者IDでアイテムを検索する
	FindBySeller(ctx context.Context, sellerID int64, limit, offset int) ([]*entity.Item, error)

	// FindByCategory はカテゴリIDでアイテムを検索する
	FindByCategory(ctx context.Context, categoryID int64, limit, offset int) ([]*entity.Item, error)

	// FindByStatus はステータスでアイテムを検索する
	FindByStatus(ctx context.Context, status entity.ItemStatus, limit, offset int) ([]*entity.Item, error)

	// Search はキーワードでアイテムを検索する（単一キーワード）
	Search(ctx context.Context, keyword string, limit, offset int) ([]*entity.Item, error)

	// SearchByKeywords は複数キーワードでアイテムを検索する
	SearchByKeywords(ctx context.Context, keywords []string, limit, offset int) ([]*entity.Item, error)

	// Update はアイテム情報を更新する
	Update(ctx context.Context, item *entity.Item) error

	// Delete はアイテムを論理削除する
	Delete(ctx context.Context, itemID int64) error

	// IncrementViewCount は閲覧数を増やす
	IncrementViewCount(ctx context.Context, itemID int64) error
}
