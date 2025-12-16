# Docker環境構築ガイド

## 🚀 クイックスタート

```bash
# 開発環境の起動
make up

# ログを確認
make logs

# 停止
make down
```

## 📦 利用可能なコマンド

```bash
make help          # 全コマンド一覧を表示
make up            # Docker環境を起動（バックグラウンド）
make dev           # Docker環境を起動（ログ表示）
make down          # Docker環境を停止
make build         # Dockerイメージを再ビルド
make restart       # 再起動
make logs          # バックエンドのログを表示
make logs-db       # MySQLのログを表示
make db-shell      # MySQLに接続
make backend-shell # バックエンドコンテナに接続
make clean         # 完全クリーンアップ（ボリューム含む）
```

## 🗄️ マイグレーション

```bash
# マイグレーション実行（通常は自動実行されます）
make migrate-up

# ロールバック
make migrate-down

# 新規マイグレーション作成
make migrate-create name=add_users_table
```

## 🔧 初回セットアップ

1. **環境変数の確認**
   `.env` ファイルが存在することを確認してください。

2. **Docker環境の起動**
   ```bash
   make up
   ```

3. **データベースの確認**
   ```bash
   make db-shell
   # MySQLシェル内で
   SHOW TABLES;
   ```

4. **アプリケーションの確認**
   ブラウザで `http://localhost:8080` にアクセス

## 📊 データベース接続情報

- **ホスト**: localhost
- **ポート**: 3306
- **ユーザー**: root
- **パスワード**: password
- **データベース名**: rpg_market

## 🐛 トラブルシューティング

### ポートが既に使用されている

```bash
# 使用中のプロセスを確認
lsof -i :3306
lsof -i :8080

# または別のポートを使用（docker-compose.ymlを編集）
```

### マイグレーションエラー

```bash
# コンテナ内でマイグレーション状態を確認
make backend-shell
migrate -path ./migrations -database "mysql://root:password@tcp(mysql:3306)/rpg_market" version

# 強制的にバージョンを設定
migrate -path ./migrations -database "mysql://root:password@tcp(mysql:3306)/rpg_market" force <version>
```

### 完全リセット

```bash
make clean
make build
make up
```

## 📁 プロジェクト構造

```
.
├── cmd/
│   └── api/
│       └── main.go           # アプリケーションエントリーポイント
├── config/
│   ├── config.go             # 設定管理
│   └── database.go           # DB接続管理
├── migrations/               # SQLマイグレーションファイル
├── docker-compose.yml        # Docker構成
├── Dockerfile                # アプリケーションイメージ
├── Makefile                  # 便利コマンド集
└── migrate-and-start.sh      # 起動スクリプト
```

## 🌐 本番環境デプロイ（Cloud Run + Cloud SQL）

開発環境との違い:
- DB接続: Cloud SQL Proxyまたはプライベート接続
- マイグレーション: CI/CDパイプラインで実行
- 環境変数: Secret Managerで管理

```bash
# Cloud Run用のビルド
docker build -t gcr.io/PROJECT_ID/rpg-market-backend .
docker push gcr.io/PROJECT_ID/rpg-market-backend

# Cloud Runへデプロイ
gcloud run deploy rpg-market-backend \
  --image gcr.io/PROJECT_ID/rpg-market-backend \
  --add-cloudsql-instances PROJECT_ID:REGION:INSTANCE_NAME
```
