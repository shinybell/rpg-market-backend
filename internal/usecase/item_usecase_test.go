package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/shinybell/rpg-market-backend/internal/domain/entity"
)

// MockItemRepository はItemRepositoryのモック実装
type MockItemRepository struct {
	items map[int64]*entity.Item
}

func NewMockItemRepository() *MockItemRepository {
	return &MockItemRepository{
		items: make(map[int64]*entity.Item),
	}
}

// MockCommentRepository はCommentRepositoryのモック実装
type MockCommentRepository struct {
	comments map[int64][]*entity.Comment
}

func NewMockCommentRepository() *MockCommentRepository {
	return &MockCommentRepository{
		comments: make(map[int64][]*entity.Comment),
	}
}

func (r *MockCommentRepository) CountByItemID(ctx context.Context, itemID int64) (int, error) {
	comments, exists := r.comments[itemID]
	if !exists {
		return 0, nil
	}
	return len(comments), nil
}

func (r *MockCommentRepository) Create(ctx context.Context, comment *entity.Comment) error {
	if r.comments[comment.ItemID] == nil {
		r.comments[comment.ItemID] = []*entity.Comment{}
	}
	comment.ID = int64(len(r.comments[comment.ItemID]) + 1)
	r.comments[comment.ItemID] = append(r.comments[comment.ItemID], comment)
	return nil
}

func (r *MockCommentRepository) FindByItemID(ctx context.Context, itemID int64, limit, offset int) ([]*entity.Comment, error) {
	comments, exists := r.comments[itemID]
	if !exists {
		return []*entity.Comment{}, nil
	}
	return comments, nil
}

func (r *MockCommentRepository) Delete(ctx context.Context, commentID, userID int64) error {
	// Simple implementation: just remove from all items
	for itemID, comments := range r.comments {
		for i, comment := range comments {
			if comment.ID == commentID && comment.UserID == userID {
				r.comments[itemID] = append(comments[:i], comments[i+1:]...)
				return nil
			}
		}
	}
	return nil
}

func (r *MockItemRepository) Create(ctx context.Context, item *entity.Item) error {
	item.ID = int64(len(r.items) + 1)
	r.items[item.ID] = item
	return nil
}

func (r *MockItemRepository) FindByID(ctx context.Context, id int64) (*entity.Item, error) {
	item, exists := r.items[id]
	if !exists {
		return nil, nil
	}
	return item, nil
}

func (r *MockItemRepository) FindBySeller(ctx context.Context, sellerID int64, limit, offset int) ([]*entity.Item, error) {
	var items []*entity.Item
	for _, item := range r.items {
		if item.SellerID == sellerID && item.DeletedAt == nil {
			items = append(items, item)
		}
	}
	return items, nil
}

func (r *MockItemRepository) FindByCategory(ctx context.Context, categoryID int64, limit, offset int) ([]*entity.Item, error) {
	var items []*entity.Item
	for _, item := range r.items {
		if item.CategoryID == categoryID && item.Status == entity.ItemStatusOnSale && item.DeletedAt == nil {
			items = append(items, item)
		}
	}
	return items, nil
}

func (r *MockItemRepository) FindByStatus(ctx context.Context, status entity.ItemStatus, limit, offset int) ([]*entity.Item, error) {
	var items []*entity.Item
	for _, item := range r.items {
		if item.Status == status && item.DeletedAt == nil {
			items = append(items, item)
		}
	}
	return items, nil
}

func (r *MockItemRepository) Search(ctx context.Context, keyword string, limit, offset int) ([]*entity.Item, error) {
	var items []*entity.Item
	for _, item := range r.items {
		if item.Status == entity.ItemStatusOnSale && item.DeletedAt == nil {
			items = append(items, item)
		}
	}
	return items, nil
}

func (r *MockItemRepository) Update(ctx context.Context, item *entity.Item) error {
	if _, exists := r.items[item.ID]; !exists {
		return errors.New("item not found")
	}
	r.items[item.ID] = item
	return nil
}

func (r *MockItemRepository) Delete(ctx context.Context, itemID int64) error {
	item, exists := r.items[itemID]
	if !exists {
		return errors.New("item not found")
	}
	now := item.CreatedAt // Use a time value from the item
	item.DeletedAt = &now
	return nil
}

func (r *MockItemRepository) IncrementViewCount(ctx context.Context, itemID int64) error {
	item, exists := r.items[itemID]
	if !exists {
		return errors.New("item not found")
	}
	item.ViewCount++
	return nil
}

func TestCreateItem(t *testing.T) {
	repo := NewMockItemRepository()
	commentRepo := NewMockCommentRepository()
	uc := NewItemUseCase(repo, commentRepo)

	item := &entity.Item{
		SellerID:    1,
		CategoryID:  1,
		Name:        "Test Item",
		Description: "This is a test item description.",
		Price:       1000.00,
		Stock:       10,
		Condition:   entity.ItemConditionNew,
		Status:      entity.ItemStatusOnSale,
	}

	err := uc.CreateItem(context.Background(), item)
	if err != nil {
		t.Fatalf("Failed to create item: %v", err)
	}

	if item.ID == 0 {
		t.Error("Item ID should be set after creation")
	}
}

func TestCreateItemWithInvalidName(t *testing.T) {
	repo := NewMockItemRepository()
	commentRepo := NewMockCommentRepository()
	uc := NewItemUseCase(repo, commentRepo)

	item := &entity.Item{
		SellerID:    1,
		CategoryID:  1,
		Name:        "", // Invalid: empty name
		Description: "This is a test item description.",
		Price:       1000,
		Stock:       10,
		Condition:   entity.ItemConditionNew,
		Status:      entity.ItemStatusOnSale,
	}

	err := uc.CreateItem(context.Background(), item)
	if err != entity.ErrInvalidItemName {
		t.Errorf("Expected ErrInvalidItemName, got: %v", err)
	}
}

func TestCreateItemWithInvalidPrice(t *testing.T) {
	repo := NewMockItemRepository()
	commentRepo := NewMockCommentRepository()
	uc := NewItemUseCase(repo, commentRepo)

	item := &entity.Item{
		SellerID:    1,
		CategoryID:  1,
		Name:        "Test Item",
		Description: "This is a test item description.",
		Price:       -100, // Invalid: negative price
		Stock:       10,
		Condition:   entity.ItemConditionNew,
		Status:      entity.ItemStatusOnSale,
	}

	err := uc.CreateItem(context.Background(), item)
	if err != entity.ErrInvalidPrice {
		t.Errorf("Expected ErrInvalidPrice, got: %v", err)
	}
}

func TestGetItemByID(t *testing.T) {
	repo := NewMockItemRepository()
	commentRepo := NewMockCommentRepository()
	uc := NewItemUseCase(repo, commentRepo)

	// Create test item
	item := &entity.Item{
		SellerID:    1,
		CategoryID:  1,
		Name:        "Test Item",
		Description: "This is a test item description.",
		Price:       1000,
		Stock:       10,
		Condition:   entity.ItemConditionNew,
		Status:      entity.ItemStatusOnSale,
	}
	_ = uc.CreateItem(context.Background(), item)

	// Get item
	retrieved, err := uc.GetItemByID(context.Background(), item.ID)
	if err != nil {
		t.Fatalf("Failed to get item: %v", err)
	}

	if retrieved.Name != item.Name {
		t.Errorf("Expected name %s, got %s", item.Name, retrieved.Name)
	}
}

func TestGetItemByIDNotFound(t *testing.T) {
	repo := NewMockItemRepository()
	commentRepo := NewMockCommentRepository()
	uc := NewItemUseCase(repo, commentRepo)

	_, err := uc.GetItemByID(context.Background(), 999)
	if err != ErrItemNotFound {
		t.Errorf("Expected ErrItemNotFound, got: %v", err)
	}
}

func TestUpdateItem(t *testing.T) {
	repo := NewMockItemRepository()
	commentRepo := NewMockCommentRepository()
	uc := NewItemUseCase(repo, commentRepo)

	// Create test item
	item := &entity.Item{
		SellerID:    1,
		CategoryID:  1,
		Name:        "Test Item",
		Description: "This is a test item description.",
		Price:       1000,
		Stock:       10,
		Condition:   entity.ItemConditionNew,
		Status:      entity.ItemStatusOnSale,
	}
	_ = uc.CreateItem(context.Background(), item)

	// Update item
	item.Name = "Updated Item"
	item.Price = 1500

	err := uc.UpdateItem(context.Background(), item, 1) // Same seller
	if err != nil {
		t.Fatalf("Failed to update item: %v", err)
	}

	// Verify update
	updated, _ := uc.GetItemByID(context.Background(), item.ID)
	if updated.Name != "Updated Item" {
		t.Errorf("Expected name 'Updated Item', got %s", updated.Name)
	}
	if updated.Price != 1500 {
		t.Errorf("Expected price 1500, got %d", updated.Price)
	}
}

func TestUpdateItemUnauthorized(t *testing.T) {
	repo := NewMockItemRepository()
	commentRepo := NewMockCommentRepository()
	uc := NewItemUseCase(repo, commentRepo)

	// Create test item
	item := &entity.Item{
		SellerID:    1,
		CategoryID:  1,
		Name:        "Test Item",
		Description: "This is a test item description.",
		Price:       1000,
		Stock:       10,
		Condition:   entity.ItemConditionNew,
		Status:      entity.ItemStatusOnSale,
	}
	_ = uc.CreateItem(context.Background(), item)

	// Try to update as different user
	item.Name = "Hacked Item"
	err := uc.UpdateItem(context.Background(), item, 2) // Different seller

	if err != entity.ErrUnauthorizedEdit {
		t.Errorf("Expected ErrUnauthorizedEdit, got: %v", err)
	}
}

func TestDeleteItem(t *testing.T) {
	repo := NewMockItemRepository()
	commentRepo := NewMockCommentRepository()
	uc := NewItemUseCase(repo, commentRepo)

	// Create test item
	item := &entity.Item{
		SellerID:    1,
		CategoryID:  1,
		Name:        "Test Item",
		Description: "This is a test item description.",
		Price:       1000,
		Stock:       10,
		Condition:   entity.ItemConditionNew,
		Status:      entity.ItemStatusOnSale,
	}
	_ = uc.CreateItem(context.Background(), item)

	// Delete item
	err := uc.DeleteItem(context.Background(), item.ID, 1) // Same seller
	if err != nil {
		t.Fatalf("Failed to delete item: %v", err)
	}

	// Verify deletion
	deleted, _ := repo.FindByID(context.Background(), item.ID)
	if deleted.DeletedAt == nil {
		t.Error("Item should be marked as deleted")
	}
}

func TestDeleteItemUnauthorized(t *testing.T) {
	repo := NewMockItemRepository()
	commentRepo := NewMockCommentRepository()
	uc := NewItemUseCase(repo, commentRepo)

	// Create test item
	item := &entity.Item{
		SellerID:    1,
		CategoryID:  1,
		Name:        "Test Item",
		Description: "This is a test item description.",
		Price:       1000,
		Stock:       10,
		Condition:   entity.ItemConditionNew,
		Status:      entity.ItemStatusOnSale,
	}
	_ = uc.CreateItem(context.Background(), item)

	// Try to delete as different user
	err := uc.DeleteItem(context.Background(), item.ID, 2) // Different seller

	if err != entity.ErrUnauthorizedEdit {
		t.Errorf("Expected ErrUnauthorizedEdit, got: %v", err)
	}
}

func TestPurchaseItem(t *testing.T) {
	repo := NewMockItemRepository()
	commentRepo := NewMockCommentRepository()
	uc := NewItemUseCase(repo, commentRepo)

	// Create test item
	item := &entity.Item{
		SellerID:    1,
		CategoryID:  1,
		Name:        "Test Item",
		Description: "This is a test item description.",
		Price:       1000,
		Stock:       5,
		Condition:   entity.ItemConditionNew,
		Status:      entity.ItemStatusOnSale,
	}
	_ = uc.CreateItem(context.Background(), item)

	// Purchase item
	err := uc.PurchaseItem(context.Background(), item.ID, 2)
	if err != nil {
		t.Fatalf("Failed to purchase item: %v", err)
	}

	// Verify stock decreased
	updated, _ := uc.GetItemByID(context.Background(), item.ID)
	if updated.Stock != 4 {
		t.Errorf("Expected stock 4, got %d", updated.Stock)
	}
}

func TestPurchaseItemOutOfStock(t *testing.T) {
	repo := NewMockItemRepository()
	commentRepo := NewMockCommentRepository()
	uc := NewItemUseCase(repo, commentRepo)

	// Create test item with 1 stock
	item := &entity.Item{
		SellerID:    1,
		CategoryID:  1,
		Name:        "Test Item",
		Description: "This is a test item description.",
		Price:       1000,
		Stock:       1,
		Condition:   entity.ItemConditionNew,
		Status:      entity.ItemStatusOnSale,
	}
	_ = uc.CreateItem(context.Background(), item)

	// Purchase first time (should succeed)
	_ = uc.PurchaseItem(context.Background(), item.ID, 2)

	// Try to purchase again (should fail with out of stock)
	err := uc.PurchaseItem(context.Background(), item.ID, 3)
	if err != ErrItemNotForSale && err != entity.ErrOutOfStock {
		t.Errorf("Expected ErrItemNotForSale or ErrOutOfStock, got: %v", err)
	}
}

func TestGetItemsBySeller(t *testing.T) {
	repo := NewMockItemRepository()
	commentRepo := NewMockCommentRepository()
	uc := NewItemUseCase(repo, commentRepo)

	// Create multiple items for seller 1
	for i := 0; i < 3; i++ {
		item := &entity.Item{
			SellerID:    1,
			CategoryID:  1,
			Name:        "Test Item",
			Description: "This is a test item description.",
			Price:       1000,
			Stock:       10,
			Condition:   entity.ItemConditionNew,
			Status:      entity.ItemStatusOnSale,
		}
		_ = uc.CreateItem(context.Background(), item)
	}

	// Get items by seller
	items, err := uc.GetItemsBySeller(context.Background(), 1, 10, 0)
	if err != nil {
		t.Fatalf("Failed to get items by seller: %v", err)
	}

	if len(items) != 3 {
		t.Errorf("Expected 3 items, got %d", len(items))
	}
}

func TestGetItemsByCategory(t *testing.T) {
	repo := NewMockItemRepository()
	commentRepo := NewMockCommentRepository()
	uc := NewItemUseCase(repo, commentRepo)

	// Create items in category 1
	for i := 0; i < 2; i++ {
		item := &entity.Item{
			SellerID:    1,
			CategoryID:  1,
			Name:        "Test Item",
			Description: "This is a test item description.",
			Price:       1000,
			Stock:       10,
			Condition:   entity.ItemConditionNew,
			Status:      entity.ItemStatusOnSale,
		}
		_ = uc.CreateItem(context.Background(), item)
	}

	// Create item in category 2
	item := &entity.Item{
		SellerID:    1,
		CategoryID:  2,
		Name:        "Test Item",
		Description: "This is a test item description.",
		Price:       1000,
		Stock:       10,
		Condition:   entity.ItemConditionNew,
		Status:      entity.ItemStatusOnSale,
	}
	_ = uc.CreateItem(context.Background(), item)

	// Get items by category 1
	items, err := uc.GetItemsByCategory(context.Background(), 1, 10, 0)
	if err != nil {
		t.Fatalf("Failed to get items by category: %v", err)
	}

	if len(items) != 2 {
		t.Errorf("Expected 2 items in category 1, got %d", len(items))
	}
}
