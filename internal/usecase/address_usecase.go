package usecase

import (
	"context"
	"errors"

	"github.com/shinybell/rpg-market-backend/internal/domain/entity"
	"github.com/shinybell/rpg-market-backend/internal/domain/repository"
)

var (
	ErrAddressNotFound = errors.New("address not found")
)

type AddressUseCase struct {
	addressRepo repository.AddressRepository
}

// NewAddressUseCase はAddressUseCaseを生成する
func NewAddressUseCase(addressRepo repository.AddressRepository) *AddressUseCase {
	return &AddressUseCase{
		addressRepo: addressRepo,
	}
}

// CreateAddress は配送先を作成する
func (uc *AddressUseCase) CreateAddress(ctx context.Context, addr *entity.Address) error {
	return uc.addressRepo.Create(ctx, addr)
}

// GetAddressesByUserID はユーザーIDで配送先一覧を取得する
func (uc *AddressUseCase) GetAddressesByUserID(ctx context.Context, userID int64) ([]*entity.Address, error) {
	return uc.addressRepo.FindByUserID(ctx, userID)
}

// UpdateAddress は配送先を更新する
func (uc *AddressUseCase) UpdateAddress(ctx context.Context, addr *entity.Address, userID int64) error {
	existing, err := uc.addressRepo.FindByID(ctx, addr.ID)
	if err != nil {
		return err
	}
	if existing == nil {
		return ErrAddressNotFound
	}
	if existing.UserID != userID {
		return errors.New("unauthorized")
	}
	return uc.addressRepo.Update(ctx, addr)
}

// DeleteAddress は配送先を削除する
func (uc *AddressUseCase) DeleteAddress(ctx context.Context, addressID, userID int64) error {
	addr, err := uc.addressRepo.FindByID(ctx, addressID)
	if err != nil {
		return err
	}
	if addr == nil {
		return ErrAddressNotFound
	}
	if addr.UserID != userID {
		return errors.New("unauthorized")
	}
	return uc.addressRepo.Delete(ctx, addressID)
}
