package presenter

import (
	"github.com/shinybell/rpg-market-backend/internal/domain/entity"
)

// ItemResponse はアイテムのレスポンス
type ItemResponse struct {
	ID               int64                `json:"id"`
	SellerID         int64                `json:"seller_id"`
	Seller           *UserResponse        `json:"seller,omitempty"`
	CategoryID       int64                `json:"category_id"`
	BrandID          *int64               `json:"brand_id,omitempty"`
	Name             string               `json:"name"`
	Description      string               `json:"description"`
	Price            int64                `json:"price"`
	Stock            int                  `json:"stock"`
	Condition        entity.ItemCondition `json:"condition"`
	ShippingPayer    entity.ShippingPayer `json:"shipping_payer"`
	ShippingMethodID *int                 `json:"shipping_method_id,omitempty"`
	ShippingDays     entity.ShippingDays  `json:"shipping_days"`
	PrefectureID     *int                 `json:"prefecture_id,omitempty"`
	Status           entity.ItemStatus    `json:"status"`
	LikesCount       int                  `json:"likes_count"`
	CommentsCount    int                  `json:"comments_count"`
	ViewCount        int                  `json:"view_count"`
	Images           []ItemImageResponse  `json:"images,omitempty"`
	CreatedAt        string               `json:"created_at"`
	UpdatedAt        string               `json:"updated_at"`
}

// ItemImageResponse はアイテム画像のレスポンス
type ItemImageResponse struct {
	ID           int64  `json:"id"`
	ImageURL     string `json:"image_url"`
	DisplayOrder int    `json:"display_order"`
}

// ToItemResponse はエンティティをレスポンスに変換する
func ToItemResponse(item *entity.Item) ItemResponse {
	resp := ItemResponse{
		ID:               item.ID,
		SellerID:         item.SellerID,
		CategoryID:       item.CategoryID,
		BrandID:          item.BrandID,
		Name:             item.Name,
		Description:      item.Description,
		Price:            item.Price,
		Stock:            item.Stock,
		Condition:        item.Condition,
		ShippingPayer:    item.ShippingPayer,
		ShippingMethodID: item.ShippingMethodID,
		ShippingDays:     item.ShippingDays,
		PrefectureID:     item.PrefectureID,
		Status:           item.Status,
		LikesCount:       item.LikesCount,
		CommentsCount:    item.CommentsCount,
		ViewCount:        item.ViewCount,
		CreatedAt:        item.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:        item.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	// 出品者情報
	if item.Seller != nil {
		seller := ToUserResponse(item.Seller)
		resp.Seller = &seller
	}

	// 画像情報
	if len(item.Images) > 0 {
		images := make([]ItemImageResponse, len(item.Images))
		for i, img := range item.Images {
			images[i] = ItemImageResponse{
				ID:           img.ID,
				ImageURL:     img.ImageURL,
				DisplayOrder: img.DisplayOrder,
			}
		}
		resp.Images = images
	}

	return resp
}

// ToItemListResponse はアイテムリストをレスポンスに変換する
func ToItemListResponse(items []*entity.Item) []ItemResponse {
	responses := make([]ItemResponse, len(items))
	for i, item := range items {
		responses[i] = ToItemResponse(item)
	}
	return responses
}
