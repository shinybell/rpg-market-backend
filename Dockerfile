# syntax=docker/dockerfile:1

FROM golang:1.25 AS base

ARG ENV=production
ARG TARGET_STAGE=runtime

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

FROM base AS development

# credentialsディレクトリは開発用のみにCOPY
COPY ./credentials ./credentials

EXPOSE 8080

# 開発時はgo runでホットリロード
CMD ["go", "run", "./cmd/api"]

FROM base AS builder

# credentialsディレクトリはビルド用のみにCOPY
COPY ./credentials ./credentials

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/api

FROM alpine:latest AS runtime

RUN apk --no-cache add ca-certificates libc6-compat

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/main .

# Copy swagger docs
COPY --from=builder /app/docs ./docs

# 本番環境ではcredentialsコピーしない

EXPOSE 8080

CMD ["./main"]
