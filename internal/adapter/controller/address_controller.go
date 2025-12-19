package controller

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/shinybell/rpg-market-backend/internal/domain/entity"
	"github.com/shinybell/rpg-market-backend/internal/usecase"
)

type AddressController struct {
	addressUseCase *usecase.AddressUseCase
	userUseCase    *usecase.UserUseCase
}

// NewAddressController はAddressControllerを生成する
func NewAddressController(addressUseCase *usecase.AddressUseCase, userUseCase *usecase.UserUseCase) *AddressController {
	return &AddressController{
		addressUseCase: addressUseCase,
		userUseCase:    userUseCase,
	}
}

type CreateAddressRequest struct {
	Name       string `json:"name" binding:"required"`
	PostalCode string `json:"postal_code" binding:"required"`
	Address    string `json:"address" binding:"required"`
	Phone      string `json:"phone" binding:"required"`
}

type UpdateAddressRequest struct {
	Name       *string `json:"name,omitempty"`
	PostalCode *string `json:"postal_code,omitempty"`
	Address    *string `json:"address,omitempty"`
	Phone      *string `json:"phone,omitempty"`
}

// CreateAddress は配送先を作成する
func (ctrl *AddressController) CreateAddress(c *gin.Context) {
	var req CreateAddressRequest
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

	user, err := ctrl.userUseCase.GetUserByFirebaseUID(c.Request.Context(), firebaseUID.(string))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	addr := &entity.Address{
		UserID:     user.ID,
		Name:       req.Name,
		PostalCode: req.PostalCode,
		Address:    req.Address,
		Phone:      req.Phone,
	}

	if err := ctrl.addressUseCase.CreateAddress(c.Request.Context(), addr); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create address: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, addr)
}

// GetAddresses は配送先一覧を取得する
func (ctrl *AddressController) GetAddresses(c *gin.Context) {
	// Firebase UIDを取得
	firebaseUID, exists := c.Get("firebase_uid")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Firebase UID not found"})
		return
	}

	user, err := ctrl.userUseCase.GetUserByFirebaseUID(c.Request.Context(), firebaseUID.(string))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	addresses, err := ctrl.addressUseCase.GetAddressesByUserID(c.Request.Context(), user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get addresses: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, addresses)
}

// UpdateAddress は配送先を更新する
func (ctrl *AddressController) UpdateAddress(c *gin.Context) {
	addressIDStr := c.Param("id")
	addressID, err := strconv.ParseInt(addressIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid address ID"})
		return
	}

	var req UpdateAddressRequest
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

	user, err := ctrl.userUseCase.GetUserByFirebaseUID(c.Request.Context(), firebaseUID.(string))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	addr := &entity.Address{
		ID: addressID,
	}
	if req.Name != nil {
		addr.Name = *req.Name
	}
	if req.PostalCode != nil {
		addr.PostalCode = *req.PostalCode
	}
	if req.Address != nil {
		addr.Address = *req.Address
	}
	if req.Phone != nil {
		addr.Phone = *req.Phone
	}

	if err := ctrl.addressUseCase.UpdateAddress(c.Request.Context(), addr, user.ID); err != nil {
		if err == usecase.ErrAddressNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Address not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update address: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, addr)
}

// DeleteAddress は配送先を削除する
func (ctrl *AddressController) DeleteAddress(c *gin.Context) {
	addressIDStr := c.Param("id")
	addressID, err := strconv.ParseInt(addressIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid address ID"})
		return
	}

	// Firebase UIDを取得
	firebaseUID, exists := c.Get("firebase_uid")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Firebase UID not found"})
		return
	}

	user, err := ctrl.userUseCase.GetUserByFirebaseUID(c.Request.Context(), firebaseUID.(string))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	if err := ctrl.addressUseCase.DeleteAddress(c.Request.Context(), addressID, user.ID); err != nil {
		if err == usecase.ErrAddressNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Address not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete address: " + err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}
