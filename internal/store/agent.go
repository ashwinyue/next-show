// Package store 提供数据访问层.
package store

import (
	"context"

	"gorm.io/gorm"

	"github.com/ashwinyue/next-show/internal/model"
)

// AgentStore Agent 存储接口.
type AgentStore interface {
	Create(ctx context.Context, agent *model.Agent) error
	Get(ctx context.Context, id string) (*model.Agent, error)
	GetByName(ctx context.Context, name string) (*model.Agent, error)
	GetWithProvider(ctx context.Context, id string) (*model.Agent, error)
	Update(ctx context.Context, agent *model.Agent) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, offset, limit int) ([]*model.Agent, int64, error)
	ListEnabled(ctx context.Context) ([]*model.Agent, error)
}

type agentStore struct {
	db *gorm.DB
}

func newAgentStore(db *gorm.DB) AgentStore {
	return &agentStore{db: db}
}

func (s *agentStore) Create(ctx context.Context, agent *model.Agent) error {
	return s.db.WithContext(ctx).Create(agent).Error
}

func (s *agentStore) Get(ctx context.Context, id string) (*model.Agent, error) {
	var agent model.Agent
	if err := s.db.WithContext(ctx).Where("id = ?", id).First(&agent).Error; err != nil {
		return nil, err
	}
	return &agent, nil
}

func (s *agentStore) GetByName(ctx context.Context, name string) (*model.Agent, error) {
	var agent model.Agent
	if err := s.db.WithContext(ctx).Where("name = ?", name).First(&agent).Error; err != nil {
		return nil, err
	}
	return &agent, nil
}

func (s *agentStore) GetWithProvider(ctx context.Context, id string) (*model.Agent, error) {
	var agent model.Agent
	if err := s.db.WithContext(ctx).Preload("Provider").Where("id = ?", id).First(&agent).Error; err != nil {
		return nil, err
	}
	return &agent, nil
}

func (s *agentStore) Update(ctx context.Context, agent *model.Agent) error {
	return s.db.WithContext(ctx).Save(agent).Error
}

func (s *agentStore) Delete(ctx context.Context, id string) error {
	return s.db.WithContext(ctx).Where("id = ?", id).Delete(&model.Agent{}).Error
}

func (s *agentStore) List(ctx context.Context, offset, limit int) ([]*model.Agent, int64, error) {
	var agents []*model.Agent
	var total int64

	db := s.db.WithContext(ctx).Model(&model.Agent{})
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := db.Offset(offset).Limit(limit).Order("created_at DESC").Find(&agents).Error; err != nil {
		return nil, 0, err
	}
	return agents, total, nil
}

func (s *agentStore) ListEnabled(ctx context.Context) ([]*model.Agent, error) {
	var agents []*model.Agent
	if err := s.db.WithContext(ctx).Where("is_enabled = ?", true).Find(&agents).Error; err != nil {
		return nil, err
	}
	return agents, nil
}
