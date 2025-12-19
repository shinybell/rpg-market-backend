#!/bin/bash

# Cloud SQLへのマイグレーション実行スクリプト
# Cloud SQL Proxyを使用してローカルから安全に接続

set -e

# カラー出力
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}=== Cloud SQL Migration Script ===${NC}"

# 環境変数の読み込み
if [ -f .env.cloudsql ]; then
    export $(cat .env.cloudsql | grep -v '^#' | xargs)
    echo -e "${GREEN}✓ Loaded .env.cloudsql${NC}"
else
    echo -e "${RED}✗ .env.cloudsql not found!${NC}"
    echo "Please create .env.cloudsql with your Cloud SQL connection details"
    exit 1
fi

# 必須変数のチェック
if [ -z "$CLOUD_SQL_CONNECTION_NAME" ] || [ -z "$DB_USER" ] || [ -z "$DB_PASSWORD" ] || [ -z "$DB_NAME" ]; then
    echo -e "${RED}✗ Missing required environment variables${NC}"
    echo "Required: CLOUD_SQL_CONNECTION_NAME, DB_USER, DB_PASSWORD, DB_NAME"
    exit 1
fi

# Cloud SQL Proxyのチェック
if ! command -v cloud-sql-proxy &> /dev/null; then
    echo -e "${YELLOW}⚠ Cloud SQL Proxy not found${NC}"
    echo "Installing Cloud SQL Proxy..."

    # OS判定
    OS=$(uname -s)
    ARCH=$(uname -m)

    if [ "$OS" = "Darwin" ]; then
        if [ "$ARCH" = "arm64" ]; then
            PROXY_URL="https://storage.googleapis.com/cloud-sql-connectors/cloud-sql-proxy/v2.15.0/cloud-sql-proxy.darwin.arm64"
        else
            PROXY_URL="https://storage.googleapis.com/cloud-sql-connectors/cloud-sql-proxy/v2.15.0/cloud-sql-proxy.darwin.amd64"
        fi
    elif [ "$OS" = "Linux" ]; then
        PROXY_URL="https://storage.googleapis.com/cloud-sql-connectors/cloud-sql-proxy/v2.15.0/cloud-sql-proxy.linux.amd64"
    else
        echo -e "${RED}✗ Unsupported OS: $OS${NC}"
        exit 1
    fi

    curl -o cloud-sql-proxy "$PROXY_URL"
    chmod +x cloud-sql-proxy
    sudo mv cloud-sql-proxy /usr/local/bin/

    echo -e "${GREEN}✓ Cloud SQL Proxy installed${NC}"
fi

# migrate-cliのチェック
if ! command -v migrate &> /dev/null; then
    echo -e "${RED}✗ migrate-cli not found${NC}"
    echo "Please install migrate-cli:"
    echo "  macOS: brew install golang-migrate"
    echo "  Linux: See https://github.com/golang-migrate/migrate/tree/master/cmd/migrate#installation"
    exit 1
fi

# Cloud SQL Proxyを起動
echo -e "${YELLOW}Starting Cloud SQL Proxy...${NC}"
cloud-sql-proxy "$CLOUD_SQL_CONNECTION_NAME" --port=$CLOUDSQL_PROXY_PORT &
PROXY_PID=$!

# Proxyの起動を待機
echo "Waiting for Cloud SQL Proxy to be ready..."
sleep 3

# トラップ設定（スクリプト終了時にProxyを停止）
trap "echo -e '\n${YELLOW}Stopping Cloud SQL Proxy...${NC}'; kill $PROXY_PID 2>/dev/null || true; exit" EXIT INT TERM

# 接続確認
echo "Testing connection to Cloud SQL..."
for i in {1..10}; do
    if mysql -h $DB_HOST -P $DB_PORT -u $DB_USER -p$DB_PASSWORD --default-auth=caching_sha2_password -e "SELECT 1" &> /dev/null; then
        echo -e "${GREEN}✓ Connected to Cloud SQL${NC}"
        break
    fi

    if [ $i -eq 10 ]; then
        echo -e "${RED}✗ Failed to connect to Cloud SQL${NC}"
        echo "Please check:"
        echo "  1. Your IP is allowed in Cloud SQL authorized networks"
        echo "  2. Database credentials are correct"
        echo "  3. Cloud SQL instance is running"
        exit 1
    fi

    echo "Attempt $i/10 - Retrying..."
    sleep 2
done

# データベース接続文字列
# tls=skip-verify を追加してTLS検証をスキップ（Cloud SQL Proxy使用時は安全）
DB_URL="mysql://${DB_USER}:${DB_PASSWORD}@tcp(${DB_HOST}:${DB_PORT})/${DB_NAME}?tls=skip-verify&multiStatements=true"

# マイグレーションコマンド
case "${1:-up}" in
    up)
        echo -e "${GREEN}Running migrations UP...${NC}"
        migrate -path ./migrations -database "$DB_URL" up
        echo -e "${GREEN}✓ Migrations completed successfully${NC}"
        ;;
    down)
        echo -e "${YELLOW}Running migrations DOWN...${NC}"
        migrate -path ./migrations -database "$DB_URL" down 1
        echo -e "${GREEN}✓ Rollback completed${NC}"
        ;;
    version)
        echo -e "${GREEN}Current migration version:${NC}"
        migrate -path ./migrations -database "$DB_URL" version
        ;;
    force)
        if [ -z "$2" ]; then
            echo -e "${RED}✗ Usage: $0 force <version>${NC}"
            exit 1
        fi
        echo -e "${YELLOW}Forcing migration to version: $2${NC}"
        migrate -path ./migrations -database "$DB_URL" force "$2"
        echo -e "${GREEN}✓ Migration version forced${NC}"
        ;;
    goto)
        if [ -z "$2" ]; then
            echo -e "${RED}✗ Usage: $0 goto <version>${NC}"
            exit 1
        fi
        echo -e "${YELLOW}Migrating to version: $2${NC}"
        migrate -path ./migrations -database "$DB_URL" goto "$2"
        echo -e "${GREEN}✓ Migration completed${NC}"
        ;;
    *)
        echo "Usage: $0 {up|down|version|force|goto}"
        echo ""
        echo "Commands:"
        echo "  up              - Apply all pending migrations (default)"
        echo "  down            - Rollback last migration"
        echo "  version         - Show current migration version"
        echo "  force <version> - Force migration to specific version"
        echo "  goto <version>  - Migrate to specific version"
        exit 1
        ;;
esac

echo -e "${GREEN}Done!${NC}"
