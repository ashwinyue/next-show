// Package store 提供数据访问层.
package store

import (
	"context"
	"time"

	"gorm.io/gorm"

	"github.com/ashwinyue/next-show/internal/model"
)

// MessageStore Message 存储接口.
type MessageStore interface {
	Create(ctx context.Context, message *model.Message) error
	Get(ctx context.Context, id string) (*model.Message, error)
	Update(ctx context.Context, message *model.Message) error
	ListBySession(ctx context.Context, sessionID string) ([]*model.Message, error)
	ListBySessionWithFilter(ctx context.Context, sessionID string, beforeTime time.Time, limit int) ([]*model.Message, error)
}

type messageStore struct {
	db *gorm.DB
}

func newMessageStore(db *gorm.DB) MessageStore {
	return &messageStore{db: db}
}

func (s *messageStore) Create(ctx context.Context, message *model.Message) error {
	return s.db.WithContext(ctx).Create(message).Error
}

func (s *messageStore) Get(ctx context.Context, id string) (*model.Message, error) {
	var message model.Message
	if err := s.db.WithContext(ctx).Where("id = ?", id).First(&message).Error; err != nil {
		return nil, err
	}
	return &message, nil
}

func (s *messageStore) Update(ctx context.Context, message *model.Message) error {
	return s.db.WithContext(ctx).Save(message).Error
}

func (s *messageStore) ListBySession(ctx context.Context, sessionID string) ([]*model.Message, error) {
	var messages []*model.Message
	if err := s.db.WithContext(ctx).Where("session_id = ?", sessionID).Order("created_at ASC").Find(&messages).Error; err != nil {
		return nil, err
	}
	return messages, nil
}

// ListBySessionWithFilter 获取会话消息（支持时间过滤和限制，对齐 WeKnora）.
func (s *messageStore) ListBySessionWithFilter(ctx context.Context, sessionID string, beforeTime time.Time, limit int) ([]*model.Message, error) {
	query := s.db.WithContext(ctx).Where("session_id = ?", sessionID)

	// 时间过滤（获取指定时间之前的消息）
	if !beforeTime.IsZero() {
		query = query.Where("created_at < ?", beforeTime)
	}

	// 限制数量（倒序获取最新的 limit 条，再倒回来）
	query = query.Order("created_at DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}

	var messages []*model.Message
	if err := query.Find(&messages).Error; err != nil {
		return nil, err
	}

	// 倒序返回（旧 -> 新）
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, nil
}
