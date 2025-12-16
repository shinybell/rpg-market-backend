package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/shinybell/rpg-market-backend/internal/usecase"
)

type UserController struct {
	userUseCase *usecase.UserUseCase
}

// NewUserController はUserControllerを生成する
func NewUserController(userUseCase *usecase.UserUseCase) *UserController {
	return &UserController{
		userUseCase: userUseCase,
	}
}

type LoginRequest struct {
	Nickname string `json:"nickname" binding:"required,min=2,max=50"`
}

type UpdateProfileRequest struct {
	Nickname  *string `json:"nickname,omitempty" binding:"omitempty,min=2,max=50"`
	Bio       *string `json:"bio,omitempty"`
	AvatarURL *string `json:"avatar_url,omitempty"`
}

// @Summary ユーザーログイン/登録
// @Description Firebase認証後、バックエンドのDBにユーザー情報を同期する
// @Tags auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "ニックネーム"
// @Success 200 {object} entity.User
// @Success 201 {object} entity.User
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /api/auth/login [post]
func (ctrl *UserController) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// ミドルウェアで検証済みのFirebase情報を取得
	firebaseUID, exists := c.Get("firebase_uid")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Firebase UID not found"})
		return
	}

	email, _ := c.Get("email")
	emailStr, ok := email.(string)
	if !ok {
		emailStr = ""
	}

	// ユーザー登録 or ログイン
	user, err := ctrl.userUseCase.LoginOrRegister(
		c.Request.Context(),
		firebaseUID.(string),
		emailStr,
		req.Nickname,
	)
	if err != nil {
		if err == usecase.ErrInvalidNickname {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to login: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, user)
}

// @Summary ユーザー情報取得
// @Description 認証されたユーザーのプロフィール情報を取得
// @Tags auth
// @Produce json
// @Success 200 {object} entity.User
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /api/auth/me [get]
func (ctrl *UserController) GetMe(c *gin.Context) {
	firebaseUID, exists := c.Get("firebase_uid")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Firebase UID not found"})
		return
	}

	user, err := ctrl.userUseCase.GetUserByFirebaseUID(c.Request.Context(), firebaseUID.(string))
	if err != nil {
		if err == usecase.ErrUserNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, user)
}

// @Summary プロフィール更新
// @Description ユーザーのプロフィール情報を更新
// @Tags user
// @Accept json
// @Produce json
// @Param request body UpdateProfileRequest true "更新する情報"
// @Success 200 {object} entity.User
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /api/users/profile [put]
func (ctrl *UserController) UpdateProfile(c *gin.Context) {
	firebaseUID, exists := c.Get("firebase_uid")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Firebase UID not found"})
		return
	}

	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	user, err := ctrl.userUseCase.UpdateProfile(
		c.Request.Context(),
		firebaseUID.(string),
		req.Nickname,
		req.Bio,
		req.AvatarURL,
	)
	if err != nil {
		if err == usecase.ErrUserNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		if err == usecase.ErrInvalidNickname {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, user)
}

// @Summary ユーザー削除
// @Description ユーザーアカウントを論理削除
// @Tags user
// @Produce json
// @Success 204
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /api/users [delete]
func (ctrl *UserController) DeleteUser(c *gin.Context) {
	firebaseUID, exists := c.Get("firebase_uid")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Firebase UID not found"})
		return
	}

	err := ctrl.userUseCase.DeleteUser(c.Request.Context(), firebaseUID.(string))
	if err != nil {
		if err == usecase.ErrUserNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}
