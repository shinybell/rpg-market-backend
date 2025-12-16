package db

import (
	"fmt"
	"log"

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
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBName,
	)

	// ログレベルの設定
	logLevel := logger.Silent
	if cfg.Environment == "development" {
		logLevel = logger.Info
	}

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
	sqlDB.SetMaxOpenConns(100)

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
