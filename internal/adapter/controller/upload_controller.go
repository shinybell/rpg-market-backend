package controller

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/shinybell/rpg-market-backend/internal/infrastructure/storage"
)

// UploadController は画像アップロード関連のコントローラー
type UploadController struct {
	gcsClient *storage.GCSClient
}

// NewUploadController はUploadControllerを初期化する
func NewUploadController(gcsClient *storage.GCSClient) *UploadController {
	return &UploadController{
		gcsClient: gcsClient,
	}
}

// GenerateSignedURLRequest は署名付きURL生成のリクエスト
type GenerateSignedURLRequest struct {
	Filename    string `json:"filename" binding:"required"`
	ContentType string `json:"content_type" binding:"required"`
}

// GenerateSignedURLResponse は署名付きURL生成のレスポンス
type GenerateSignedURLResponse struct {
	UploadURL  string `json:"upload_url"`  // フロントエンドがアップロードに使用するURL
	ObjectName string `json:"object_name"` // GCS上のオブジェクト名
	BucketName string `json:"bucket_name"` // GCSバケット名
}

// GenerateSignedURL は署名付きURLを生成する
// @Summary 画像アップロード用の署名付きURL生成
// @Description 画像アップロード用の署名付きURLとGCSのURLを生成
// @Tags upload
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param request body GenerateSignedURLRequest true "ファイル名とContent-Type"
// @Success 200 {object} GenerateSignedURLResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/upload/signed-url [post]
// @Security BearerAuth
func (ctrl *UploadController) GenerateSignedURL(c *gin.Context) {
	var req GenerateSignedURLRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// オブジェクト名を生成（この時点でタイムスタンプが決まる）
	objectName := ctrl.gcsClient.GetObjectName(req.Filename)

	// 署名付きURLを生成
	signedURL, err := ctrl.gcsClient.GenerateSignedUploadURL(c.Request.Context(), objectName, req.ContentType)
	if err != nil {
		log.Printf("Failed to generate signed URL: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate signed URL"})
		return
	}

	c.JSON(http.StatusOK, GenerateSignedURLResponse{
		UploadURL:  signedURL,
		ObjectName: objectName,
		BucketName: ctrl.gcsClient.BucketName(),
	})
}
