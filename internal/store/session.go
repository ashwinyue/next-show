// Package store 提供数据访问层.
package store

import (
	"context"

	"gorm.io/gorm"

	"github.com/mervyn/next-show/internal/model"
)

// SessionStore Session 存储接口.
type SessionStore interface {
	Create(ctx context.Context, session *model.Session) error
	Get(ctx context.Context, id string) (*model.Session, error)
	GetWithAgent(ctx context.Context, id string) (*model.Session, error)
	Update(ctx context.Context, session *model.Session) error
	UpdateTitle(ctx context.Context, id, title string) error
	UpdateStatus(ctx context.Context, id string, status model.SessionStatus) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, userID string, offset, limit int) ([]*model.Session, int64, error)
	ListByAgent(ctx context.Context, agentID string, offset, limit int) ([]*model.Session, int64, error)
}

type sessionStore struct {
	db *gorm.DB
}

func newSessionStore(db *gorm.DB) SessionStore {
	return &sessionStore{db: db}
}

func (s *sessionStore) Create(ctx context.Context, session *model.Session) error {
	return s.db.WithContext(ctx).Create(session).Error
}

func (s *sessionStore) Get(ctx context.Context, id string) (*model.Session, error) {
	var session model.Session
	if err := s.db.WithContext(ctx).Where("id = ?", id).First(&session).Error; err != nil {
		return nil, err
	}
	return &session, nil
}

func (s *sessionStore) GetWithAgent(ctx context.Context, id string) (*model.Session, error) {
	var session model.Session
	if err := s.db.WithContext(ctx).Preload("Agent").Where("id = ?", id).First(&session).Error; err != nil {
		return nil, err
	}
	return &session, nil
}

func (s *sessionStore) Update(ctx context.Context, session *model.Session) error {
	return s.db.WithContext(ctx).Save(session).Error
}

func (s *sessionStore) UpdateTitle(ctx context.Context, id, title string) error {
	return s.db.WithContext(ctx).Model(&model.Session{}).Where("id = ?", id).Update("title", title).Error
}

func (s *sessionStore) UpdateStatus(ctx context.Context, id string, status model.SessionStatus) error {
	return s.db.WithContext(ctx).Model(&model.Session{}).Where("id = ?", id).Update("status", status).Error
}

func (s *sessionStore) Delete(ctx context.Context, id string) error {
	return s.db.WithContext(ctx).Model(&model.Session{}).Where("id = ?", id).Update("status", model.SessionStatusDeleted).Error
}

func (s *sessionStore) List(ctx context.Context, userID string, offset, limit int) ([]*model.Session, int64, error) {
	var sessions []*model.Session
	var total int64

	db := s.db.WithContext(ctx).Model(&model.Session{}).Where("user_id = ? AND status != ?", userID, model.SessionStatusDeleted)
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := db.Offset(offset).Limit(limit).Order("created_at DESC").Find(&sessions).Error; err != nil {
		return nil, 0, err
	}
	return sessions, total, nil
}

func (s *sessionStore) ListByAgent(ctx context.Context, agentID string, offset, limit int) ([]*model.Session, int64, error) {
	var sessions []*model.Session
	var total int64

	db := s.db.WithContext(ctx).Model(&model.Session{}).Where("agent_id = ? AND status != ?", agentID, model.SessionStatusDeleted)
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := db.Offset(offset).Limit(limit).Order("created_at DESC").Find(&sessions).Error; err != nil {
		return nil, 0, err
	}
	return sessions, total, nil
}
