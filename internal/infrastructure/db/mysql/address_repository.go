package mysql

import (
	"context"

	"gorm.io/gorm"

	"github.com/shinybell/rpg-market-backend/internal/domain/entity"
	"github.com/shinybell/rpg-market-backend/internal/domain/repository"
)

type addressRepository struct {
	db *gorm.DB
}

// NewAddressRepository はAddressRepositoryの実装を返す
func NewAddressRepository(db *gorm.DB) repository.AddressRepository {
	return &addressRepository{db: db}
}

// Create は配送先を作成する
func (r *addressRepository) Create(ctx context.Context, addr *entity.Address) error {
	return r.db.WithContext(ctx).Create(addr).Error
}

// FindByID はIDで配送先を検索する
func (r *addressRepository) FindByID(ctx context.Context, id int64) (*entity.Address, error) {
	var addr entity.Address
	err := r.db.WithContext(ctx).First(&addr, id).Error
	if err != nil {
		return nil, err
	}
	return &addr, nil
}

// FindByUserID はユーザーIDで配送先一覧を取得する
func (r *addressRepository) FindByUserID(ctx context.Context, userID int64) ([]*entity.Address, error) {
	var addresses []*entity.Address
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&addresses).Error
	return addresses, err
}

// Update は配送先を更新する
func (r *addressRepository) Update(ctx context.Context, addr *entity.Address) error {
	return r.db.WithContext(ctx).Save(addr).Error
}

// Delete は配送先を削除する
func (r *addressRepository) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&entity.Address{}).Error
}
