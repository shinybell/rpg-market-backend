#!/bin/bash

# Docker経由でCloud SQLマイグレーションを実行
# MySQL 9.xでの認証プラグイン問題を回避

set -e

# カラー出力
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}=== Cloud SQL Migration (Docker) ===${NC}"

# 環境変数の読み込み
if [ -f .env.cloudsql ]; then
    export $(cat .env.cloudsql | grep -v '^#' | xargs)
    echo -e "${GREEN}✓ Loaded .env.cloudsql${NC}"
else
    echo -e "${RED}✗ .env.cloudsql not found!${NC}"
    exit 1
fi

# 必須変数のチェック
if [ -z "$CLOUD_SQL_CONNECTION_NAME" ] || [ -z "$DB_USER" ] || [ -z "$DB_PASSWORD" ] || [ -z "$DB_NAME" ]; then
    echo -e "${RED}✗ Missing required environment variables${NC}"
    exit 1
fi

# Dockerイメージを使用してマイグレーション実行
echo -e "${YELLOW}Running migration via Docker with MySQL 8 client...${NC}"

docker run --rm \
  --network host \
  -v "$(pwd)/migrations:/migrations" \
  -e "DB_USER=$DB_USER" \
  -e "DB_PASSWORD=$DB_PASSWORD" \
  -e "DB_HOST=${DB_HOST:-127.0.0.1}" \
  -e "DB_PORT=${DB_PORT:-3306}" \
  -e "DB_NAME=$DB_NAME" \
  migrate/migrate:v4.17.0 \
  -path=/migrations \
  -database "mysql://${DB_USER}:${DB_PASSWORD}@tcp(${DB_HOST:-127.0.0.1}:${DB_PORT:-3306})/${DB_NAME}?tls=skip-verify&multiStatements=true" \
  ${1:-up}

echo -e "${GREEN}✓ Migration completed${NC}"
