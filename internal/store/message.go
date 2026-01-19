// Package store 提供数据访问层.
package store

import (
	"context"

	"gorm.io/gorm"

	"github.com/mervyn/next-show/internal/model"
)

// MessageStore Message 存储接口.
type MessageStore interface {
	Create(ctx context.Context, message *model.Message) error
	Get(ctx context.Context, id string) (*model.Message, error)
	Update(ctx context.Context, message *model.Message) error
	ListBySession(ctx context.Context, sessionID string) ([]*model.Message, error)
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
