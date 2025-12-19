package usecase

import (
	"context"
	"log"

	"github.com/shinybell/rpg-market-backend/internal/domain/entity"
	"github.com/shinybell/rpg-market-backend/internal/domain/repository"
)

type NotificationUseCase struct {
	notificationRepo repository.NotificationRepository
}

// NewNotificationUseCase はNotificationUseCaseを生成する
func NewNotificationUseCase(notificationRepo repository.NotificationRepository) *NotificationUseCase {
	return &NotificationUseCase{
		notificationRepo: notificationRepo,
	}
}

// CreateNotification は通知を作成する
func (uc *NotificationUseCase) CreateNotification(ctx context.Context, n *entity.Notification) error {
	return uc.notificationRepo.Create(ctx, n)
}

// SendPurchaseEmail は購入完了メールを送信する（モック）
func (uc *NotificationUseCase) SendPurchaseEmail(ctx context.Context, userID int64, itemName string) {
	log.Printf("TODO: Send purchase email to user %d for item %s", userID, itemName)
}

// SendSaleEmail は販売完了メールを送信する（モック）
func (uc *NotificationUseCase) SendSaleEmail(ctx context.Context, userID int64, itemName string) {
	log.Printf("TODO: Send sale email to user %d for item %s", userID, itemName)
}
