package mysql

import (
	"context"

	"gorm.io/gorm"

	"github.com/shinybell/rpg-market-backend/internal/domain/entity"
	"github.com/shinybell/rpg-market-backend/internal/domain/repository"
)

type commentRepository struct {
	db *gorm.DB
}

// NewCommentRepository はCommentRepositoryの実装を返す
func NewCommentRepository(db *gorm.DB) repository.CommentRepository {
	return &commentRepository{db: db}
}

// Create はコメントを作成する
func (r *commentRepository) Create(ctx context.Context, comment *entity.Comment) error {
	return r.db.WithContext(ctx).Create(comment).Error
}

// FindByItemID はアイテムIDでコメントを検索する
func (r *commentRepository) FindByItemID(ctx context.Context, itemID int64, limit, offset int) ([]*entity.Comment, error) {
	var comments []*entity.Comment
	err := r.db.WithContext(ctx).
		Preload("User").
		Preload("User.Profile").
		Where("item_id = ? AND is_deleted = false", itemID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&comments).Error
	return comments, err
}

// Delete はコメントを削除する
func (r *commentRepository) Delete(ctx context.Context, commentID, userID int64) error {
	return r.db.WithContext(ctx).Where("id = ? AND user_id = ?", commentID, userID).Delete(&entity.Comment{}).Error
}
