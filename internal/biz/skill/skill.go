// Package skill 提供 Skill 业务逻辑层.
package skill

import (
	"context"

	"github.com/ashwinyue/next-show/internal/model"
	"github.com/ashwinyue/next-show/internal/store"
)

// Biz Skill 业务逻辑接口.
type Biz interface {
	// CRUD 操作
	Create(ctx context.Context, skill *model.Skill) (*model.Skill, error)
	Get(ctx context.Context, id string) (*model.Skill, error)
	Update(ctx context.Context, skill *model.Skill) (*model.Skill, error)
	Delete(ctx context.Context, id string) error

	// 列表和搜索
	List(ctx context.Context, page, pageSize int) ([]*model.Skill, int64, error)
	ListAll(ctx context.Context) ([]*model.Skill, error)
	ListByCategory(ctx context.Context, category string) ([]*model.Skill, error)
	ListEnabled(ctx context.Context) ([]*model.Skill, error)
	Search(ctx context.Context, keyword string, page, pageSize int) ([]*model.Skill, int64, error)

	// Agent 集成
	ApplyToAgent(ctx context.Context, agentID, skillID string, temporary bool) error
}

type biz struct {
	store store.Store
}

// NewBiz 创建 Skill 业务层实例.
func NewBiz(store store.Store) Biz {
	return &biz{store: store}
}

func (b *biz) Create(ctx context.Context, skill *model.Skill) (*model.Skill, error) {
	if err := b.store.Skills().Create(ctx, skill); err != nil {
		return nil, err
	}
	return skill, nil
}

func (b *biz) Get(ctx context.Context, id string) (*model.Skill, error) {
	return b.store.Skills().Get(ctx, id)
}

func (b *biz) Update(ctx context.Context, skill *model.Skill) (*model.Skill, error) {
	if err := b.store.Skills().Update(ctx, skill); err != nil {
		return nil, err
	}
	return skill, nil
}

func (b *biz) Delete(ctx context.Context, id string) error {
	return b.store.Skills().Delete(ctx, id)
}

func (b *biz) List(ctx context.Context, page, pageSize int) ([]*model.Skill, int64, error) {
	offset := (page - 1) * pageSize
	return b.store.Skills().List(ctx, offset, pageSize)
}

func (b *biz) ListAll(ctx context.Context) ([]*model.Skill, error) {
	return b.store.Skills().ListAll(ctx)
}

func (b *biz) ListByCategory(ctx context.Context, category string) ([]*model.Skill, error) {
	return b.store.Skills().ListByCategory(ctx, category)
}

func (b *biz) ListEnabled(ctx context.Context) ([]*model.Skill, error) {
	return b.store.Skills().ListEnabled(ctx)
}

func (b *biz) Search(ctx context.Context, keyword string, page, pageSize int) ([]*model.Skill, int64, error) {
	offset := (page - 1) * pageSize
	return b.store.Skills().Search(ctx, keyword, offset, pageSize)
}

// ApplyToAgent 将 Skill 应用到 Agent.
func (b *biz) ApplyToAgent(ctx context.Context, agentID, skillID string, temporary bool) error {
	// 1. 加载 Skill
	skill, err := b.store.Skills().Get(ctx, skillID)
	if err != nil {
		return err
	}

	// 2. 获取 Agent
	agent, err := b.store.Agents().Get(ctx, agentID)
	if err != nil {
		return err
	}

	// 初始化 Config（如果为 nil）
	if agent.Config == nil {
		agent.Config = make(model.JSONMap)
	}

	// 3. 应用配置
	if skill.SystemPrompt != "" {
		agent.SystemPrompt = skill.SystemPrompt
	}
	if skill.Instructions != "" {
		agent.Config["instructions"] = skill.Instructions
	}
	if len(skill.ToolIDs) > 0 {
		agent.Config["tool_ids"] = skill.ToolIDs
	}
	if len(skill.KnowledgeBaseIDs) > 0 {
		agent.Config["knowledge_base_ids"] = skill.KnowledgeBaseIDs
	}
	if skill.ModelProvider != "" {
		agent.ProviderID = skill.ModelProvider
	}
	if skill.ModelName != "" {
		agent.ModelName = skill.ModelName
	}
	if skill.Temperature > 0 {
		agent.Temperature = &skill.Temperature
	}
	if skill.MaxIterations > 0 {
		agent.MaxIterations = skill.MaxIterations
	}

	// 4. 更新 Agent
	return b.store.Agents().Update(ctx, agent)
}
