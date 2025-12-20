package usecase

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/shinybell/rpg-market-backend/internal/domain/entity"
	"github.com/shinybell/rpg-market-backend/internal/domain/repository"
)

var (
	ErrItemNotFound   = errors.New("item not found")
	ErrItemNotForSale = errors.New("item not for sale")
)

type ItemUseCase struct {
	itemRepo            repository.ItemRepository
	commentRepo         repository.CommentRepository
	transactionRepo     repository.TransactionRepository
	walletRepo          repository.WalletRepository
	notificationRepo    repository.NotificationRepository
	addressRepo         repository.AddressRepository
	notificationUseCase *NotificationUseCase
}

// NewItemUseCase はItemUseCaseを生成する
func NewItemUseCase(itemRepo repository.ItemRepository, commentRepo repository.CommentRepository, transactionRepo repository.TransactionRepository, walletRepo repository.WalletRepository, notificationRepo repository.NotificationRepository, addressRepo repository.AddressRepository, notificationUseCase *NotificationUseCase) *ItemUseCase {
	return &ItemUseCase{
		itemRepo:            itemRepo,
		commentRepo:         commentRepo,
		transactionRepo:     transactionRepo,
		walletRepo:          walletRepo,
		notificationRepo:    notificationRepo,
		addressRepo:         addressRepo,
		notificationUseCase: notificationUseCase,
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

	// コメント数をカウントして設定
	commentsCount, err := uc.commentRepo.CountByItemID(ctx, itemID)
	if err != nil {
		log.Printf("Failed to count comments for item %d: %v", itemID, err)
		// エラーが発生しても処理を続行（コメント数0として扱う）
		item.CommentsCount = 0
	} else {
		item.CommentsCount = commentsCount
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

// SearchItemsByKeywords は複数キーワードでアイテムを検索する
func (uc *ItemUseCase) SearchItemsByKeywords(ctx context.Context, keywords []string, limit, offset int) ([]*entity.Item, error) {
	return uc.itemRepo.SearchByKeywords(ctx, keywords, limit, offset)
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

// MockPayment は決済をモックする（常時成功）
func (uc *ItemUseCase) MockPayment(ctx context.Context, amount int64) error {
	log.Printf("Mock payment processed for amount: %d", amount)
	return nil
}

// PurchaseItem はアイテムを購入する
func (uc *ItemUseCase) PurchaseItem(ctx context.Context, itemID, buyerID, addressID int64, paymentMethod string, pointsUsed int64) error {
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

	// 配送先を取得
	address, err := uc.addressRepo.FindByID(ctx, addressID)
	if err != nil {
		return err
	}
	if address == nil || address.UserID != buyerID {
		return errors.New("invalid address")
	}

	// ウォレットを取得
	buyerWallet, err := uc.walletRepo.FindByUserID(ctx, buyerID)
	if err != nil {
		return err
	}
	if buyerWallet == nil {
		return errors.New("buyer wallet not found")
	}

	sellerWallet, err := uc.walletRepo.FindByUserID(ctx, item.SellerID)
	if err != nil {
		return err
	}
	if sellerWallet == nil {
		return errors.New("seller wallet not found")
	}

	// ポイント使用チェック
	if pointsUsed < 0 || buyerWallet.Points < pointsUsed {
		return entity.ErrInsufficientPoints
	}

	// 最終価格計算
	finalPrice := item.Price - pointsUsed
	if finalPrice < 0 {
		finalPrice = 0
	}

	// 残高チェック
	if buyerWallet.Balance < finalPrice {
		return entity.ErrInsufficientBalance
	}

	// 決済モック
	err = uc.MockPayment(ctx, finalPrice)
	if err != nil {
		return err
	}

	// トランザクション作成
	tx := &entity.Transaction{
		ItemID:            itemID,
		BuyerID:           buyerID,
		SellerID:          item.SellerID,
		Price:             item.Price,
		FeeAmount:         0, // TODO: 手数料計算
		ProfitAmount:      finalPrice,
		PaymentStatus:     entity.PaymentStatusCaptured,
		TransactionStatus: entity.TransactionStatusAwaitingShip,
		PaymentMethod:     paymentMethod,
	}
	err = uc.transactionRepo.Create(ctx, tx)
	if err != nil {
		return err
	}

	// ウォレット更新
	err = buyerWallet.Withdraw(finalPrice)
	if err != nil {
		return err
	}
	buyerWallet.Points -= pointsUsed
	err = uc.walletRepo.Update(ctx, buyerWallet)
	if err != nil {
		return err
	}

	err = sellerWallet.Deposit(finalPrice)
	if err != nil {
		return err
	}
	err = uc.walletRepo.Update(ctx, sellerWallet)
	if err != nil {
		return err
	}

	// アイテム更新
	err = item.DecrementStock()
	if err != nil {
		return err
	}
	if item.Stock == 0 {
		item.Status = entity.ItemStatusSoldOut
	}
	err = uc.itemRepo.Update(ctx, item)
	if err != nil {
		return err
	}

	// 通知作成
	buyerNotification := &entity.Notification{
		UserID:    buyerID,
		Type:      "purchase",
		Title:     "購入完了",
		Content:   fmt.Sprintf("アイテム %s を購入しました", item.Name),
		RelatedID: &tx.ID,
	}
	err = uc.notificationRepo.Create(ctx, buyerNotification)
	if err != nil {
		log.Printf("Failed to create buyer notification: %v", err)
	}

	sellerNotification := &entity.Notification{
		UserID:    item.SellerID,
		Type:      "sale",
		Title:     "販売完了",
		Content:   fmt.Sprintf("アイテム %s が売れました", item.Name),
		RelatedID: &tx.ID,
	}
	err = uc.notificationRepo.Create(ctx, sellerNotification)
	if err != nil {
		log.Printf("Failed to create seller notification: %v", err)
	}

	// メール通知モック TODO
	uc.notificationUseCase.SendPurchaseEmail(ctx, buyerID, item.Name)
	uc.notificationUseCase.SendSaleEmail(ctx, item.SellerID, item.Name)

	log.Printf("Purchase completed: Item %d, Buyer %d, Seller %d", itemID, buyerID, item.SellerID)
	return nil
}

// GetItemTransactionByUserID はアイテムの取引情報を取得（購入者または出品者のみ）
func (uc *ItemUseCase) GetItemTransactionByUserID(ctx context.Context, itemID, userID int64) (*entity.Transaction, error) {
	// アイテム情報を取得
	item, err := uc.itemRepo.FindByID(ctx, itemID)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, ErrItemNotFound
	}

	// このアイテムに関連する取引を検索
	transaction, err := uc.transactionRepo.FindByItemID(ctx, itemID)
	if err != nil {
		return nil, errors.New("transaction not found")
	}

	// ユーザーが購入者または出品者であることを確認
	if transaction.BuyerID != userID && transaction.SellerID != userID {
		return nil, errors.New("access denied: you are not part of this transaction")
	}

	return transaction, nil
}

// GetPurchasesByBuyerID は購入者の購入履歴を取得する
func (uc *ItemUseCase) GetPurchasesByBuyerID(ctx context.Context, buyerID int64) ([]*entity.Transaction, error) {
	return uc.transactionRepo.FindByBuyerID(ctx, buyerID)
}

// GetTransactionByIDForUser はトランザクションIDで取引を取得し、ユーザーが関係者か検証する
func (uc *ItemUseCase) GetTransactionByIDForUser(ctx context.Context, txID, userID int64) (*entity.Transaction, error) {
	tx, err := uc.transactionRepo.FindByID(ctx, txID)
	if err != nil {
		return nil, errors.New("transaction not found")
	}

	if tx.BuyerID != userID && tx.SellerID != userID {
		return nil, errors.New("access denied: you are not part of this transaction")
	}

	return tx, nil
}
