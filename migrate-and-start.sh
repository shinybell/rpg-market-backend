#!/bin/bash
set -e

echo "Waiting for MySQL to be ready..."

# シンプルなループで待機
for i in {1..30}; do
  echo "Attempt $i/30 - Checking MySQL connection..."
  # migrate version は初回は "no migration" エラーを返すが、それは接続OKの証拠
  if /usr/local/bin/migrate -path ./migrations -database "mysql://${DB_USER}:${DB_PASSWORD}@tcp(${DB_HOST}:${DB_PORT})/${DB_NAME}" version 2>&1 | grep -qE "(^[0-9]+$|no migration)"; then
    echo "MySQL is ready!"
    break
  fi

  if [ $i -eq 30 ]; then
    echo "MySQL did not become ready in time"
    exit 1
  fi

  sleep 2
done

echo "Running migrations..."

# Run migrations
/usr/local/bin/migrate -path ./migrations -database "mysql://${DB_USER}:${DB_PASSWORD}@tcp(${DB_HOST}:${DB_PORT})/${DB_NAME}" up

echo "Migrations completed successfully"

# Start the application
echo "Starting application..."
echo "Checking for main binary..."
ls -la /app/ || true
ls -la ./main || true

if [ -f "./main" ]; then
    exec ./main
elif [ -f "/app/main" ]; then
    exec /app/main
else
    echo "Error: main binary not found!"
    exit 1
fi
