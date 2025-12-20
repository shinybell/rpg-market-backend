package controller

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/shinybell/rpg-market-backend/internal/adapter/presenter"
	"github.com/shinybell/rpg-market-backend/internal/domain/entity"
	"github.com/shinybell/rpg-market-backend/internal/usecase"
)

type ItemController struct {
	itemUseCase *usecase.ItemUseCase
	userUseCase *usecase.UserUseCase
}

// NewItemController はItemControllerを生成する
func NewItemController(itemUseCase *usecase.ItemUseCase, userUseCase *usecase.UserUseCase) *ItemController {
	return &ItemController{
		itemUseCase: itemUseCase,
		userUseCase: userUseCase,
	}
}

type CreateItemRequest struct {
	CategoryID       int64                `json:"category_id" binding:"required"`
	BrandID          *int64               `json:"brand_id,omitempty"`
	Name             string               `json:"name" binding:"required,min=1,max=255"`
	Description      string               `json:"description" binding:"required,min=10"`
	Price            int64                `json:"price" binding:"required,gte=0"`
	Stock            int                  `json:"stock" binding:"required,gte=1"`
	Condition        entity.ItemCondition `json:"condition" binding:"required,oneof=new like_new very_good good acceptable"`
	ShippingPayer    entity.ShippingPayer `json:"shipping_payer" binding:"required,oneof=buyer seller"`
	ShippingMethodID *int                 `json:"shipping_method_id,omitempty"`
	ShippingDays     entity.ShippingDays  `json:"shipping_days" binding:"required,oneof=1-2 2-3 4-7"`
	PrefectureID     *int                 `json:"prefecture_id,omitempty"`
	Status           entity.ItemStatus    `json:"status" binding:"required,oneof=draft on_sale"`
	ImageURL         string               `json:"image_url,omitempty"`
	Images           []CreateItemImageReq `json:"images,omitempty"`
	RPGName          *string              `json:"rpg_name,omitempty" binding:"omitempty,max=255"`
	RPGDescription   *string              `json:"rpg_description,omitempty"`
}

type CreateItemImageReq struct {
	ImageURL     string `json:"image_url" binding:"required"`
	DisplayOrder int    `json:"display_order" binding:"gte=0"`
}

type UpdateItemRequest struct {
	Name             *string               `json:"name,omitempty" binding:"omitempty,min=1,max=255"`
	Description      *string               `json:"description,omitempty" binding:"omitempty,min=10"`
	Price            *int64                `json:"price,omitempty" binding:"omitempty,gte=0"`
	Stock            *int                  `json:"stock,omitempty" binding:"omitempty,gte=0"`
	Condition        *entity.ItemCondition `json:"condition,omitempty"`
	ShippingPayer    *entity.ShippingPayer `json:"shipping_payer,omitempty"`
	ShippingMethodID *int                  `json:"shipping_method_id,omitempty"`
	ShippingDays     *entity.ShippingDays  `json:"shipping_days,omitempty"`
	PrefectureID     *int                  `json:"prefecture_id,omitempty"`
	Status           *entity.ItemStatus    `json:"status,omitempty"`
	Images           []CreateItemImageReq  `json:"images,omitempty"`
	RPGName          *string               `json:"rpg_name,omitempty" binding:"omitempty,max=255"`
	RPGDescription   *string               `json:"rpg_description,omitempty"`
}

// @Summary アイテム作成
// @Description 新しいアイテムを出品する
// @Tags items
// @Accept json
// @Produce json
// @Param request body CreateItemRequest true "アイテム情報"
// @Success 201 {object} presenter.ItemResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /api/items [post]
func (ctrl *ItemController) CreateItem(c *gin.Context) {
	var req CreateItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// Firebase UIDを取得
	firebaseUID, exists := c.Get("firebase_uid")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Firebase UID not found"})
		return
	}

	// Firebase UIDからユーザー情報を取得
	user, err := ctrl.userUseCase.GetUserByFirebaseUID(c.Request.Context(), firebaseUID.(string))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	// アイテムエンティティを作成
	item := &entity.Item{
		SellerID:         user.ID,
		CategoryID:       req.CategoryID,
		BrandID:          req.BrandID,
		Name:             req.Name,
		Description:      req.Description,
		Price:            req.Price,
		Stock:            req.Stock,
		Condition:        req.Condition,
		ShippingPayer:    req.ShippingPayer,
		ShippingMethodID: req.ShippingMethodID,
		ShippingDays:     req.ShippingDays,
		PrefectureID:     req.PrefectureID,
		Status:           req.Status,
		RPGName:          req.RPGName,
		RPGDescription:   req.RPGDescription,
	}

	// 画像を追加
	if req.ImageURL != "" {
		item.Images = []entity.ItemImage{
			{
				ImageURL:     req.ImageURL,
				DisplayOrder: 0,
			},
		}
	} else if len(req.Images) > 0 {
		images := make([]entity.ItemImage, len(req.Images))
		for i, img := range req.Images {
			images[i] = entity.ItemImage{
				ImageURL:     img.ImageURL,
				DisplayOrder: img.DisplayOrder,
			}
		}
		item.Images = images
	}

	// アイテムを作成
	if err := ctrl.itemUseCase.CreateItem(c.Request.Context(), item); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create item: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, presenter.ToItemResponse(item))
}

// @Summary アイテム詳細取得
// @Description アイテムの詳細情報を取得する
// @Tags items
// @Produce json
// @Param id path int true "アイテムID"
// @Success 200 {object} presenter.ItemResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/items/{id} [get]
func (ctrl *ItemController) GetItem(c *gin.Context) {
	itemID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid item ID"})
		return
	}

	item, err := ctrl.itemUseCase.GetItemByID(c.Request.Context(), itemID)
	if err != nil {
		if err == usecase.ErrItemNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Item not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get item: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, presenter.ToItemResponse(item))
}

// @Summary アイテム一覧取得
// @Description 販売中のアイテム一覧を取得する
// @Tags items
// @Produce json
// @Param limit query int false "取得件数" default(20)
// @Param offset query int false "オフセット" default(0)
// @Success 200 {array} presenter.ItemResponse
// @Failure 500 {object} map[string]string
// @Router /api/items [get]
func (ctrl *ItemController) ListItems(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	items, err := ctrl.itemUseCase.GetItemsOnSale(c.Request.Context(), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get items: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, presenter.ToItemListResponse(items))
}

// @Summary 出品者のアイテム一覧取得
// @Description 指定した出品者のアイテム一覧を取得する
// @Tags items
// @Produce json
// @Param seller_id path int true "出品者ID"
// @Param limit query int false "取得件数" default(20)
// @Param offset query int false "オフセット" default(0)
// @Success 200 {array} presenter.ItemResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/items/seller/{seller_id} [get]
func (ctrl *ItemController) ListItemsBySeller(c *gin.Context) {
	sellerID, err := strconv.ParseInt(c.Param("seller_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid seller ID"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	items, err := ctrl.itemUseCase.GetItemsBySeller(c.Request.Context(), sellerID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get items: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, presenter.ToItemListResponse(items))
}

// @Summary カテゴリ別アイテム一覧取得
// @Description 指定したカテゴリのアイテム一覧を取得する
// @Tags items
// @Produce json
// @Param category_id path int true "カテゴリID"
// @Param limit query int false "取得件数" default(20)
// @Param offset query int false "オフセット" default(0)
// @Success 200 {array} presenter.ItemResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/items/category/{category_id} [get]
func (ctrl *ItemController) ListItemsByCategory(c *gin.Context) {
	categoryID, err := strconv.ParseInt(c.Param("category_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid category ID"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	items, err := ctrl.itemUseCase.GetItemsByCategory(c.Request.Context(), categoryID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get items: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, presenter.ToItemListResponse(items))
}

// @Summary アイテム検索
// @Description キーワードでアイテムを検索する
// @Tags items
// @Produce json
// @Param q query string true "検索キーワード"
// @Param limit query int false "取得件数" default(20)
// @Param offset query int false "オフセット" default(0)
// @Success 200 {array} presenter.ItemResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/items/search [get]
func (ctrl *ItemController) SearchItems(c *gin.Context) {
	keyword := c.Query("q")
	if keyword == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Search keyword is required"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	items, err := ctrl.itemUseCase.SearchItems(c.Request.Context(), keyword, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to search items: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, presenter.ToItemListResponse(items))
}

// @Summary アイテム更新
// @Description アイテムの情報を更新する（出品者のみ）
// @Tags items
// @Accept json
// @Produce json
// @Param id path int true "アイテムID"
// @Param request body UpdateItemRequest true "更新情報"
// @Success 200 {object} presenter.ItemResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /api/items/{id} [put]
func (ctrl *ItemController) UpdateItem(c *gin.Context) {
	itemID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid item ID"})
		return
	}

	var req UpdateItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// Firebase UIDを取得
	firebaseUID, exists := c.Get("firebase_uid")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Firebase UID not found"})
		return
	}

	// Firebase UIDからユーザー情報を取得
	user, err := ctrl.userUseCase.GetUserByFirebaseUID(c.Request.Context(), firebaseUID.(string))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	// 既存のアイテムを取得
	item, err := ctrl.itemUseCase.GetItemByID(c.Request.Context(), itemID)
	if err != nil {
		if err == usecase.ErrItemNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Item not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get item: " + err.Error()})
		return
	}

	// 更新内容を反映
	if req.Name != nil {
		item.Name = *req.Name
	}
	if req.Description != nil {
		item.Description = *req.Description
	}
	if req.Price != nil {
		item.Price = *req.Price
	}
	if req.Stock != nil {
		item.Stock = *req.Stock
	}
	if req.Condition != nil {
		item.Condition = *req.Condition
	}
	if req.ShippingPayer != nil {
		item.ShippingPayer = *req.ShippingPayer
	}
	if req.ShippingMethodID != nil {
		item.ShippingMethodID = req.ShippingMethodID
	}
	if req.ShippingDays != nil {
		item.ShippingDays = *req.ShippingDays
	}
	if req.PrefectureID != nil {
		item.PrefectureID = req.PrefectureID
	}
	if req.Status != nil {
		item.Status = *req.Status
	}
	if req.RPGName != nil {
		item.RPGName = req.RPGName
	}
	if req.RPGDescription != nil {
		item.RPGDescription = req.RPGDescription
	}
	if req.Images != nil {
		images := make([]entity.ItemImage, len(req.Images))
		for i, img := range req.Images {
			images[i] = entity.ItemImage{
				ItemID:       itemID,
				ImageURL:     img.ImageURL,
				DisplayOrder: img.DisplayOrder,
			}
		}
		item.Images = images
	}

	// アイテムを更新
	if err := ctrl.itemUseCase.UpdateItem(c.Request.Context(), item, user.ID); err != nil {
		if err == entity.ErrUnauthorizedEdit {
			c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized to edit this item"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update item: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, presenter.ToItemResponse(item))
}

// @Summary アイテム削除
// @Description アイテムを削除する（出品者のみ）
// @Tags items
// @Produce json
// @Param id path int true "アイテムID"
// @Success 204 "No Content"
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /api/items/{id} [delete]
func (ctrl *ItemController) DeleteItem(c *gin.Context) {
	itemID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid item ID"})
		return
	}

	// Firebase UIDを取得
	firebaseUID, exists := c.Get("firebase_uid")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Firebase UID not found"})
		return
	}

	// Firebase UIDからユーザー情報を取得
	user, err := ctrl.userUseCase.GetUserByFirebaseUID(c.Request.Context(), firebaseUID.(string))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	// アイテムを削除
	if err := ctrl.itemUseCase.DeleteItem(c.Request.Context(), itemID, user.ID); err != nil {
		if err == usecase.ErrItemNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Item not found"})
			return
		}
		if err == entity.ErrUnauthorizedEdit {
			c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized to delete this item"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete item: " + err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

type PurchaseItemRequest struct {
	AddressID     int64  `json:"address_id" binding:"required"`
	PaymentMethod string `json:"payment_method" binding:"required"`
	PointsUsed    int64  `json:"points_used" binding:"gte=0"`
}

// PurchaseItem はアイテムを購入する
func (ctrl *ItemController) PurchaseItem(c *gin.Context) {
	// アイテムIDを取得
	itemIDStr := c.Param("id")
	itemID, err := strconv.ParseInt(itemIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid item ID"})
		return
	}

	// リクエストボディをパース
	var req PurchaseItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Firebase UIDを取得
	firebaseUID, exists := c.Get("firebase_uid")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Firebase UID not found"})
		return
	}

	// Firebase UIDからユーザー情報を取得
	user, err := ctrl.userUseCase.GetUserByFirebaseUID(c.Request.Context(), firebaseUID.(string))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	// 購入処理
	if err := ctrl.itemUseCase.PurchaseItem(c.Request.Context(), itemID, user.ID, req.AddressID, req.PaymentMethod, req.PointsUsed); err != nil {
		if err == usecase.ErrItemNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Item not found"})
			return
		}
		if err == usecase.ErrItemNotForSale {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Item not for sale"})
			return
		}
		if err == entity.ErrInsufficientBalance {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Insufficient balance"})
			return
		}
		if err == entity.ErrInsufficientPoints {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Insufficient points"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to purchase item: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Purchase successful"})
}

// GetMyItems は自分の出品アイテムを取得
// @Summary 自分の出品アイテム取得
// @Description 自分が出品したアイテムの一覧を取得
// @Tags items
// @Produce json
// @Success 200 {array} entity.Item
// @Router /api/users/me/items [get]
func (ic *ItemController) GetMyItems(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	items, err := ic.itemUseCase.GetItemsBySeller(c.Request.Context(), userID.(int64), 100, 0)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get items"})
		return
	}

	c.JSON(http.StatusOK, items)
}

// GetMyPurchases は自分の購入アイテムを取得
// @Summary 自分の購入アイテム取得
// @Description 自分が購入したアイテムの一覧を取得
// @Tags items
// @Produce json
// @Success 200 {array} entity.Transaction
// @Router /api/users/me/purchases [get]
func (ic *ItemController) GetMyPurchases(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	purchases, err := ic.itemUseCase.GetPurchasesByBuyerID(c.Request.Context(), userID.(int64))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get purchases"})
		return
	}

	c.JSON(http.StatusOK, purchases)
}

// GetItemTransaction は商品の取引情報を取得
// @Summary 商品の取引情報取得
// @Description 自分が関わる取引情報を取得（購入者または出品者のみ）
// @Tags items
// @Produce json
// @Param id path int true "Item ID"
// @Success 200 {object} entity.Transaction
// @Router /api/items/{id}/transaction [get]
func (ic *ItemController) GetItemTransaction(c *gin.Context) {
	itemIDStr := c.Param("id")
	itemID, err := strconv.ParseInt(itemIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid item ID"})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	transaction, err := ic.itemUseCase.GetItemTransactionByUserID(c.Request.Context(), itemID, userID.(int64))
	if err != nil {
		// item_id による検索で見つからない場合はその旨を正しく返す（IDの混同を許容しない）
		log.Printf("GetItemTransaction failed (by item_id): item_id=%d user_id=%v err=%v", itemID, userID, err)
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	log.Printf("GetItemTransaction success: item_id=%d user_id=%v tx_id=%d buyer=%d seller=%d", itemID, userID, transaction.ID, transaction.BuyerID, transaction.SellerID)

	c.JSON(http.StatusOK, transaction)
}
