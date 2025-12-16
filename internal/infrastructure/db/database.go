package db

import (
	"fmt"
	"log"
	"time"

	"github.com/shinybell/rpg-market-backend/config"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Database struct {
	db *gorm.DB
}

var globalDB *Database

// NewDatabase はデータベース接続を初期化する
func NewDatabase(cfg *config.Config) (*Database, error) {
	dsn := cfg.GetDSN()

	// ログレベルの設定
	logLevel := logger.Silent
	if cfg.Environment == "development" {
		logLevel = logger.Info
	}

	log.Printf("Connecting to database: %s", maskPassword(dsn))

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// コネクションプールの設定
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	// 最大アイドル接続数
	sqlDB.SetMaxIdleConns(10)
	// 最大オープン接続数
	sqlDB.SetMaxOpenConns(20)
	// 接続の最大寿命
	sqlDB.SetConnMaxLifetime(time.Hour)

	log.Println("Database connection established successfully")

	database := &Database{db: db}
	globalDB = database
	return database, nil
}

// GetDB はGORMのDBインスタンスを返す
func (d *Database) GetDB() *gorm.DB {
	return d.db
}

// Close はデータベース接続を閉じる
func (d *Database) Close() error {
	sqlDB, err := d.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// GetGlobalDB はグローバルなDBインスタンスを返す
func GetGlobalDB() *gorm.DB {
	if globalDB == nil {
		return nil
	}
	return globalDB.db
}

// maskPassword はDSNからパスワードをマスクするユーティリティ関数
func maskPassword(dsn string) string {
	// 例: user:password@tcp(localhost:3306)/dbname
	var maskedDSN string
	n, err := fmt.Sscanf(dsn, "%[^:]:%[^@]@%s", new(string), new(string), new(string))
	if err != nil || n != 3 {
		return dsn // フォーマットが異なる場合はそのまま返す
	}
	var user, password, rest string
	fmt.Sscanf(dsn, "%[^:]:%[^@]@%s", &user, &password, &rest)
	maskedDSN = fmt.Sprintf("%s:****@%s", user, rest)
	return maskedDSN
}
