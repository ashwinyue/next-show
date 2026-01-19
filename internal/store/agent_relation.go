// Package store 提供数据访问层.
package store

import (
	"context"

	"gorm.io/gorm"

	"github.com/ashwinyue/next-show/internal/model"
)

// AgentRelationStore Agent 关系存储接口.
type AgentRelationStore interface {
	Create(ctx context.Context, relation *model.AgentRelation) error
	Delete(ctx context.Context, id string) error
	ListByParent(ctx context.Context, parentAgentID string) ([]*model.AgentRelation, error)
	ListByParentWithChild(ctx context.Context, parentAgentID string) ([]*model.AgentRelation, error)
	DeleteByParent(ctx context.Context, parentAgentID string) error
}

type agentRelationStore struct {
	db *gorm.DB
}

func newAgentRelationStore(db *gorm.DB) AgentRelationStore {
	return &agentRelationStore{db: db}
}

func (s *agentRelationStore) Create(ctx context.Context, relation *model.AgentRelation) error {
	return s.db.WithContext(ctx).Create(relation).Error
}

func (s *agentRelationStore) Delete(ctx context.Context, id string) error {
	return s.db.WithContext(ctx).Where("id = ?", id).Delete(&model.AgentRelation{}).Error
}

func (s *agentRelationStore) ListByParent(ctx context.Context, parentAgentID string) ([]*model.AgentRelation, error) {
	var relations []*model.AgentRelation
	if err := s.db.WithContext(ctx).
		Where("parent_agent_id = ?", parentAgentID).
		Order("sort_order ASC").
		Find(&relations).Error; err != nil {
		return nil, err
	}
	return relations, nil
}

func (s *agentRelationStore) ListByParentWithChild(ctx context.Context, parentAgentID string) ([]*model.AgentRelation, error) {
	var relations []*model.AgentRelation
	if err := s.db.WithContext(ctx).
		Preload("ChildAgent").
		Preload("ChildAgent.Provider").
		Where("parent_agent_id = ?", parentAgentID).
		Order("sort_order ASC").
		Find(&relations).Error; err != nil {
		return nil, err
	}
	return relations, nil
}

func (s *agentRelationStore) DeleteByParent(ctx context.Context, parentAgentID string) error {
	return s.db.WithContext(ctx).Where("parent_agent_id = ?", parentAgentID).Delete(&model.AgentRelation{}).Error
}
