package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port                       string
	Environment                string
	LogLevel                   string
	DBHost                     string
	DBPort                     string
	DBUser                     string
	DBPassword                 string
	DBName                     string
	JWTSecret                  string
	FirebaseCredentialsPath    string
	VercelDeployURL            string
	GoogleCloudCredentialsPath string
	GCSBucketName              string
}

func Load() *Config {
	// .envファイルを読み込む（開発環境用）
	// 本番環境では環境変数が設定されているため、エラーは無視
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	cfg := &Config{
		Port:                       getEnv("PORT", "8080"),
		Environment:                getEnv("ENV", "development"),
		LogLevel:                   getEnv("LOG_LEVEL", "debug"),
		DBHost:                     getEnv("DB_HOST", "localhost"),
		DBPort:                     getEnv("DB_PORT", "3306"),
		DBUser:                     getEnv("DB_USER", "root"),
		DBPassword:                 getEnv("DB_PASSWORD", ""),
		DBName:                     getEnv("DB_NAME", "rpg_market"),
		JWTSecret:                  getEnv("JWT_SECRET", "your-secret-key"),
		FirebaseCredentialsPath:    getFirebaseCredentialsPath(),
		VercelDeployURL:            getEnv("VERCEL_DEPLOY_URL", ""),
		GoogleCloudCredentialsPath: getGoogleCloudCredentialsPath(),
		GCSBucketName:              getEnv("GCS_BUCKET_NAME", "rpg-market-images"),
	}

	return cfg
}

func getFirebaseCredentialsPath() string {
	env := getEnv("ENV", "development")

	if env == "production" {
		// 本番環境: Secret Managerからマウントされたパス
		return "/secrets/firebase-credentials/firebase-credentials"
	}

	// 開発環境: .envまたは環境変数から取得
	return getEnv("FIREBASE_CREDENTIALS_PATH", "credentials/firebase-credentials.json")
}

func getGoogleCloudCredentialsPath() string {
	env := getEnv("ENV", "development")

	if env == "production" {
		// 本番環境: Secret Managerからマウントされたパス
		return "/secrets/gcs-credentials/gcs_credentials"
	}

	// 開発環境: .envまたは環境変数から取得
	return getEnv("GOOGLE_CLOUD_CREDENTIALS_PATH", "credentials/gcs_credentials.json")
}

// Cloud SQL Unix Socket接続に対応
func (c *Config) GetDSN() string {
	// Cloud SQL Unix Socket接続の判定（本番環境）
	if c.Environment == "production" {
		dsn := fmt.Sprintf("%s:%s@unix(/cloudsql/%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			c.DBUser,
			c.DBPassword,
			c.DBHost, // ここにCloud SQLの接続名を指定
			c.DBName,
		)
		return dsn
	}

	// TCP接続（開発環境）
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		c.DBUser,
		c.DBPassword,
		c.DBHost,
		c.DBPort,
		c.DBName,
	)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
