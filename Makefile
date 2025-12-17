.PHONY: help up down build logs db-shell migrate-up migrate-down migrate-create clean restart rebuild

help: ## このヘルプを表示
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

up: ## Docker環境を起動
	docker-compose up -d

down: ## Docker環境を停止
	docker-compose down

build: ## Dockerイメージを再ビルド
	docker-compose build --no-cache

rebuild: clean build up ## 完全再ビルドして起動
	@echo "Rebuild complete!"

softrebuild: down build up ## イメージを再ビルドして起動
	@echo "Soft rebuild complete!"

logs: ## ログを表示（バックエンド）
	docker-compose logs -f backend

logs-all: ## すべてのログを表示
	docker-compose logs -f

logs-db: ## ログを表示（MySQL）
	docker-compose logs -f mysql

db-shell: ## MySQLに接続
	docker-compose exec mysql mysql -uroot -ppassword rpg_market

backend-shell: ## バックエンドコンテナに接続
	docker-compose exec backend sh

migrate-up: ## マイグレーションを実行（手動）
	docker-compose up migrate

migrate-down: ## マイグレーションをロールバック
	docker run --rm -v $(PWD)/migrations:/migrations --network rpg-market-backend_default \
	migrate/migrate -path=/migrations -database "mysql://root:password@tcp(mysql:3306)/rpg_market" down 1

migrate-create: ## 新しいマイグレーションを作成 (例: make migrate-create name=add_users)
	@if [ -z "$(name)" ]; then \
	echo "Usage: make migrate-create name=migration_name"; \
	exit 1; \
	fi
	migrate create -ext sql -dir ./migrations -seq $(name)

clean: ## Docker環境とボリュームを完全削除
	docker-compose down -v
	docker system prune -f

restart: down up ## Docker環境を再起動

dev: ## 開発環境を起動してログを表示
	docker-compose up

test: ## テストを実行
	go test ./...

fmt: ## コードフォーマット
	go fmt ./...

lint: ## リント実行
	golangci-lint run

ps: ## コンテナの状態を確認
	docker-compose ps

test-coverage: ## カバレッジ付きテストを実行
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

fmt-check: ## フォーマットチェック（CI用）
	@if [ -n "$$(gofmt -l .)" ]; then \
		echo "以下のファイルがフォーマットされていません:"; \
		gofmt -l .; \
		exit 1; \
	fi

imports: ## import文を整理
	goimports -w .

.DEFAULT_GOAL := help
