#!/bin/bash

# マイグレーション管理スクリプト

set -e

# 環境変数の読み込み
if [ -f .env ]; then
    export $(cat .env | grep -v '^#' | xargs)
fi

# データベース接続文字列の構築
DB_URL="mysql://${DB_USER}:${DB_PASSWORD}@tcp(${DB_HOST}:${DB_PORT})/${DB_NAME}"

case "$1" in
    up)
        echo "Running migrations up..."
        migrate -path ./migrations -database "${DB_URL}" up
        ;;
    down)
        echo "Running migrations down..."
        migrate -path ./migrations -database "${DB_URL}" down
        ;;
    create)
        if [ -z "$2" ]; then
            echo "Usage: ./scripts/migrate.sh create <migration_name>"
            exit 1
        fi
        echo "Creating migration: $2"
        migrate create -ext sql -dir ./migrations -seq "$2"
        ;;
    force)
        if [ -z "$2" ]; then
            echo "Usage: ./scripts/migrate.sh force <version>"
            exit 1
        fi
        echo "Forcing migration version to: $2"
        migrate -path ./migrations -database "${DB_URL}" force "$2"
        ;;
    version)
        echo "Current migration version:"
        migrate -path ./migrations -database "${DB_URL}" version
        ;;
    *)
        echo "Usage: ./scripts/migrate.sh {up|down|create|force|version}"
        echo ""
        echo "Commands:"
        echo "  up                    - Apply all pending migrations"
        echo "  down                  - Rollback last migration"
        echo "  create <name>         - Create new migration files"
        echo "  force <version>       - Force migration to specific version"
        echo "  version               - Show current migration version"
        exit 1
        ;;
esac
