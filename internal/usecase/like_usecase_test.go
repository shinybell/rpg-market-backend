package usecase

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/shinybell/rpg-market-backend/internal/domain/entity"
)

// MockLikeRepository はLikeRepositoryのモック
type MockLikeRepository struct {
	mock.Mock
}

func (m *MockLikeRepository) Create(ctx context.Context, like *entity.Like) error {
	args := m.Called(ctx, like)
	return args.Error(0)
}

func (m *MockLikeRepository) Delete(ctx context.Context, userID, itemID int64) error {
	args := m.Called(ctx, userID, itemID)
	return args.Error(0)
}

func (m *MockLikeRepository) ExistsByUserAndItem(ctx context.Context, userID, itemID int64) (bool, error) {
	args := m.Called(ctx, userID, itemID)
	return args.Bool(0), args.Error(1)
}

// MockItemRepository はItemRepositoryのモック（簡易版）
type MockItemRepositoryForLike struct {
	mock.Mock
}

func (m *MockItemRepositoryForLike) Create(ctx context.Context, item *entity.Item) error {
	args := m.Called(ctx, item)
	return args.Error(0)
}

func (m *MockItemRepositoryForLike) FindByID(ctx context.Context, id int64) (*entity.Item, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Item), args.Error(1)
}

func (m *MockItemRepositoryForLike) FindBySeller(ctx context.Context, sellerID int64, limit, offset int) ([]*entity.Item, error) {
	args := m.Called(ctx, sellerID, limit, offset)
	return args.Get(0).([]*entity.Item), args.Error(1)
}

func (m *MockItemRepositoryForLike) FindByCategory(ctx context.Context, categoryID int64, limit, offset int) ([]*entity.Item, error) {
	args := m.Called(ctx, categoryID, limit, offset)
	return args.Get(0).([]*entity.Item), args.Error(1)
}

func (m *MockItemRepositoryForLike) FindByStatus(ctx context.Context, status entity.ItemStatus, limit, offset int) ([]*entity.Item, error) {
	args := m.Called(ctx, status, limit, offset)
	return args.Get(0).([]*entity.Item), args.Error(1)
}

func (m *MockItemRepositoryForLike) Search(ctx context.Context, keyword string, limit, offset int) ([]*entity.Item, error) {
	args := m.Called(ctx, keyword, limit, offset)
	return args.Get(0).([]*entity.Item), args.Error(1)
}

func (m *MockItemRepositoryForLike) SearchByKeywords(ctx context.Context, keywords []string, limit, offset int) ([]*entity.Item, error) {
	args := m.Called(ctx, keywords, limit, offset)
	return args.Get(0).([]*entity.Item), args.Error(1)
}

func (m *MockItemRepositoryForLike) Update(ctx context.Context, item *entity.Item) error {
	args := m.Called(ctx, item)
	return args.Error(0)
}

func (m *MockItemRepositoryForLike) Delete(ctx context.Context, itemID int64) error {
	args := m.Called(ctx, itemID)
	return args.Error(0)
}

func (m *MockItemRepositoryForLike) IncrementViewCount(ctx context.Context, itemID int64) error {
	args := m.Called(ctx, itemID)
	return args.Error(0)
}

func TestLikeUseCase_AddLike(t *testing.T) {
	mockLikeRepo := new(MockLikeRepository)
	mockItemRepo := new(MockItemRepositoryForLike)
	uc := NewLikeUseCase(mockLikeRepo, mockItemRepo)

	ctx := context.Background()
	userID := int64(1)
	itemID := int64(1)
	item := &entity.Item{ID: itemID, LikesCount: 0}

	// モック設定
	mockItemRepo.On("FindByID", ctx, itemID).Return(item, nil)
	mockLikeRepo.On("ExistsByUserAndItem", ctx, userID, itemID).Return(false, nil)
	mockLikeRepo.On("Create", ctx, mock.AnythingOfType("*entity.Like")).Return(nil)
	mockItemRepo.On("Update", ctx, item).Return(nil)

	// 実行
	err := uc.AddLike(ctx, userID, itemID)

	// アサート
	assert.NoError(t, err)
	assert.Equal(t, 1, item.LikesCount)
	mockLikeRepo.AssertExpectations(t)
	mockItemRepo.AssertExpectations(t)
}

func TestLikeUseCase_RemoveLike(t *testing.T) {
	mockLikeRepo := new(MockLikeRepository)
	mockItemRepo := new(MockItemRepositoryForLike)
	uc := NewLikeUseCase(mockLikeRepo, mockItemRepo)

	ctx := context.Background()
	userID := int64(1)
	itemID := int64(1)
	item := &entity.Item{ID: itemID, LikesCount: 1}

	// モック設定
	mockLikeRepo.On("ExistsByUserAndItem", ctx, userID, itemID).Return(true, nil)
	mockLikeRepo.On("Delete", ctx, userID, itemID).Return(nil)
	mockItemRepo.On("FindByID", ctx, itemID).Return(item, nil)
	mockItemRepo.On("Update", ctx, item).Return(nil)

	// 実行
	err := uc.RemoveLike(ctx, userID, itemID)

	// アサート
	assert.NoError(t, err)
	assert.Equal(t, 0, item.LikesCount)
	mockLikeRepo.AssertExpectations(t)
	mockItemRepo.AssertExpectations(t)
}
