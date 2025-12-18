package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"cloud.google.com/go/storage"
	"google.golang.org/api/option"

	"github.com/shinybell/rpg-market-backend/config"
)

// GCSClient はGoogle Cloud Storageのクライアント
type GCSClient struct {
	client              *storage.Client
	bucketName          string
	credentialsPath     string
	serviceAccountEmail string
}

// NewGCSClient はGCSクライアントを初期化する
func NewGCSClient(ctx context.Context, cfg *config.Config) (*GCSClient, error) {
	credentialsPath := cfg.GoogleCloudCredentialsPath
	if credentialsPath == "" {
		return nil, fmt.Errorf("GoogleCloudCredentialsPath is not set")
	}

	var client *storage.Client
	var err error

	client, err = storage.NewClient(ctx, option.WithCredentialsFile(credentialsPath))
	if err != nil {
		return nil, fmt.Errorf("failed to create storage client: %w", err)
	}

	// サービスアカウントのメールアドレスを取得
	var serviceAccountEmail string
	data, err := os.ReadFile(credentialsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read credentials file: %w", err)
	}
	var creds struct {
		ClientEmail string `json:"client_email"`
	}
	if err := json.Unmarshal(data, &creds); err != nil {
		return nil, fmt.Errorf("failed to parse credentials: %w", err)
	}
	serviceAccountEmail = creds.ClientEmail

	return &GCSClient{
		client:              client,
		bucketName:          cfg.GCSBucketName,
		credentialsPath:     credentialsPath,
		serviceAccountEmail: serviceAccountEmail,
	}, nil
}

// Close はクライアントをクローズする
func (g *GCSClient) Close() error {
	return g.client.Close()
}

// GenerateSignedUploadURL は署名付きアップロードURLを生成する
func (g *GCSClient) GenerateSignedUploadURL(ctx context.Context, objectName, contentType string) (string, error) {
	// 署名付きURLのオプション
	opts := &storage.SignedURLOptions{
		Scheme: storage.SigningSchemeV4,
		Method: "PUT",
		Headers: []string{
			fmt.Sprintf("Content-Type:%s", contentType),
		},
		Expires: time.Now().Add(15 * time.Minute), // 15分間有効
	}

	// クライアントの認証情報を使用して署名
	url, err := g.client.Bucket(g.bucketName).SignedURL(objectName, opts)
	if err != nil {
		return "", fmt.Errorf("failed to generate signed URL: %w", err)
	}

	return url, nil
}

// GetObjectName はファイル名からオブジェクト名を生成する
func (g *GCSClient) GetObjectName(filename string) string {
	timestamp := time.Now().Unix()
	ext := filepath.Ext(filename)
	baseName := filename[:len(filename)-len(ext)]
	return fmt.Sprintf("items/%s_%d%s", baseName, timestamp, ext)
}

// BucketName はバケット名を返す
func (g *GCSClient) BucketName() string {
	return g.bucketName
}
