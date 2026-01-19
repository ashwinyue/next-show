// Package session 提供 Session 业务逻辑.
package session

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/mervyn/next-show/internal/model"
	"github.com/mervyn/next-show/internal/store"
)

// SessionBiz Session 业务接口.
type SessionBiz interface {
	Create(ctx context.Context, userID, agentID string) (*model.Session, error)
	Get(ctx context.Context, id string) (*model.Session, error)
	List(ctx context.Context, userID string, offset, limit int) ([]*model.Session, int64, error)
	UpdateTitle(ctx context.Context, id, title string) error
	Delete(ctx context.Context, id string) error
	AddMessage(ctx context.Context, sessionID, role, content string) (*model.Message, error)
	GetMessages(ctx context.Context, sessionID string) ([]*model.Message, error)
}

type sessionBiz struct {
	store store.Store
}

// NewSessionBiz 创建 Session 业务实例.
func NewSessionBiz(s store.Store) SessionBiz {
	return &sessionBiz{store: s}
}

func (b *sessionBiz) Create(ctx context.Context, userID, agentID string) (*model.Session, error) {
	session := &model.Session{
		ID:        uuid.New().String(),
		UserID:    userID,
		AgentID:   agentID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := b.store.Sessions().Create(ctx, session); err != nil {
		return nil, err
	}
	return session, nil
}

func (b *sessionBiz) Get(ctx context.Context, id string) (*model.Session, error) {
	return b.store.Sessions().Get(ctx, id)
}

func (b *sessionBiz) List(ctx context.Context, userID string, offset, limit int) ([]*model.Session, int64, error) {
	return b.store.Sessions().List(ctx, userID, offset, limit)
}

func (b *sessionBiz) UpdateTitle(ctx context.Context, id, title string) error {
	session, err := b.store.Sessions().Get(ctx, id)
	if err != nil {
		return err
	}
	session.Title = title
	session.UpdatedAt = time.Now()
	return b.store.Sessions().Update(ctx, session)
}

func (b *sessionBiz) Delete(ctx context.Context, id string) error {
	return b.store.Sessions().Delete(ctx, id)
}

func (b *sessionBiz) AddMessage(ctx context.Context, sessionID, role, content string) (*model.Message, error) {
	message := &model.Message{
		ID:        uuid.New().String(),
		SessionID: sessionID,
		Role:      role,
		Content:   content,
		CreatedAt: time.Now(),
	}
	if err := b.store.Messages().Create(ctx, message); err != nil {
		return nil, err
	}
	return message, nil
}

func (b *sessionBiz) GetMessages(ctx context.Context, sessionID string) ([]*model.Message, error) {
	return b.store.Messages().ListBySession(ctx, sessionID)
}
