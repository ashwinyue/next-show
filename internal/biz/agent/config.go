// Package agent 提供 Agent 业务逻辑.
package agent

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/ashwinyue/next-show/internal/model"
	"github.com/ashwinyue/next-show/internal/pkg/agent/builtin"
	"github.com/ashwinyue/next-show/internal/store"
)

// ConfigBiz Agent 配置业务接口.
type ConfigBiz interface {
	// ListAgents 列出所有 Agent.
	ListAgents(ctx context.Context) ([]*model.Agent, error)
	// GetAgent 获取 Agent 详情.
	GetAgent(ctx context.Context, id string) (*model.Agent, error)
	// CreateAgent 创建 Agent.
	CreateAgent(ctx context.Context, req *CreateAgentRequest) (*model.Agent, error)
	// UpdateAgent 更新 Agent.
	UpdateAgent(ctx context.Context, id string, req *UpdateAgentRequest) (*model.Agent, error)
	// DeleteAgent 删除 Agent.
	DeleteAgent(ctx context.Context, id string) error

	// ListBuiltinAgents 列出内置 Agent.
	ListBuiltinAgents(ctx context.Context) ([]*model.Agent, error)
	// ListOrchestratorAgents 列出主控 Agent.
	ListOrchestratorAgents(ctx context.Context) ([]*model.Agent, error)
	// ListSpecialistAgents 列出子 Agent.
	ListSpecialistAgents(ctx context.Context) ([]*model.Agent, error)

	// GetAgentRelations 获取 Agent 的子 Agent 关系.
	GetAgentRelations(ctx context.Context, agentID string) ([]*model.AgentRelation, error)
	// SetAgentRelations 设置 Agent 的子 Agent 关系.
	SetAgentRelations(ctx context.Context, agentID string, subAgentIDs []string) error

	// ListAgentTools 列出 Agent 的工具.
	ListAgentTools(ctx context.Context, agentID string) ([]*model.AgentTool, error)
	// AddAgentTool 为 Agent 添加工具.
	AddAgentTool(ctx context.Context, agentID string, req *AddAgentToolRequest) (*model.AgentTool, error)
	// UpdateAgentTool 更新 Agent 工具.
	UpdateAgentTool(ctx context.Context, toolID string, req *UpdateAgentToolRequest) (*model.AgentTool, error)
	// RemoveAgentTool 移除 Agent 工具.
	RemoveAgentTool(ctx context.Context, toolID string) error
	// ListBuiltinTools 列出可用的内置工具.
	ListBuiltinTools(ctx context.Context) []string
}

// CreateAgentRequest 创建 Agent 请求.
type CreateAgentRequest struct {
	Name          string          `json:"name"`
	DisplayName   string          `json:"display_name"`
	Description   string          `json:"description"`
	ProviderID    string          `json:"provider_id"`
	ModelName     string          `json:"model_name"`
	SystemPrompt  string          `json:"system_prompt"`
	AgentType     model.AgentType `json:"agent_type"`
	AgentRole     model.AgentRole `json:"agent_role"`
	MaxIterations int             `json:"max_iterations"`
	Temperature   *float64        `json:"temperature"`
	MaxTokens     *int            `json:"max_tokens"`
	Config        model.JSONMap   `json:"config"`
	SubAgentIDs   []string        `json:"sub_agent_ids,omitempty"`
}

// UpdateAgentRequest 更新 Agent 请求.
type UpdateAgentRequest struct {
	Name          *string          `json:"name,omitempty"`
	DisplayName   *string          `json:"display_name,omitempty"`
	Description   *string          `json:"description,omitempty"`
	ProviderID    *string          `json:"provider_id,omitempty"`
	ModelName     *string          `json:"model_name,omitempty"`
	SystemPrompt  *string          `json:"system_prompt,omitempty"`
	AgentType     *model.AgentType `json:"agent_type,omitempty"`
	AgentRole     *model.AgentRole `json:"agent_role,omitempty"`
	MaxIterations *int             `json:"max_iterations,omitempty"`
	Temperature   *float64         `json:"temperature,omitempty"`
	MaxTokens     *int             `json:"max_tokens,omitempty"`
	Config        model.JSONMap    `json:"config,omitempty"`
	IsEnabled     *bool            `json:"is_enabled,omitempty"`
	SubAgentIDs   []string         `json:"sub_agent_ids,omitempty"`
}

type configBiz struct {
	store store.Store
}

// NewConfigBiz 创建 Agent 配置业务实例.
func NewConfigBiz(s store.Store) ConfigBiz {
	return &configBiz{store: s}
}

func (b *configBiz) ListAgents(ctx context.Context) ([]*model.Agent, error) {
	return b.store.Agents().ListAll(ctx)
}

func (b *configBiz) GetAgent(ctx context.Context, id string) (*model.Agent, error) {
	// 先检查是否为内置 Agent
	if builtinAgent := builtin.GetBuiltinAgent(id); builtinAgent != nil {
		return builtinAgent, nil
	}
	return b.store.Agents().GetWithProvider(ctx, id)
}

func (b *configBiz) CreateAgent(ctx context.Context, req *CreateAgentRequest) (*model.Agent, error) {
	agent := &model.Agent{
		ID:            uuid.New().String(),
		Name:          req.Name,
		DisplayName:   req.DisplayName,
		Description:   req.Description,
		ProviderID:    req.ProviderID,
		ModelName:     req.ModelName,
		SystemPrompt:  req.SystemPrompt,
		AgentType:     req.AgentType,
		AgentRole:     req.AgentRole,
		MaxIterations: req.MaxIterations,
		Temperature:   req.Temperature,
		MaxTokens:     req.MaxTokens,
		Config:        req.Config,
		IsEnabled:     true,
		IsBuiltin:     false,
	}

	if agent.AgentRole == "" {
		agent.AgentRole = model.AgentRoleSpecialist
	}
	if agent.MaxIterations <= 0 {
		agent.MaxIterations = 10
	}

	if err := b.store.Agents().Create(ctx, agent); err != nil {
		return nil, fmt.Errorf("create agent: %w", err)
	}

	// 如果有子 Agent，创建关系
	if len(req.SubAgentIDs) > 0 {
		if err := b.SetAgentRelations(ctx, agent.ID, req.SubAgentIDs); err != nil {
			return nil, fmt.Errorf("set agent relations: %w", err)
		}
	}

	return agent, nil
}

func (b *configBiz) UpdateAgent(ctx context.Context, id string, req *UpdateAgentRequest) (*model.Agent, error) {
	agent, err := b.store.Agents().Get(ctx, id)
	if err != nil {
		return nil, err
	}

	// 内置 Agent 不能修改
	if agent.IsBuiltin {
		return nil, fmt.Errorf("cannot update builtin agent")
	}

	// 更新字段
	if req.Name != nil {
		agent.Name = *req.Name
	}
	if req.DisplayName != nil {
		agent.DisplayName = *req.DisplayName
	}
	if req.Description != nil {
		agent.Description = *req.Description
	}
	if req.ProviderID != nil {
		agent.ProviderID = *req.ProviderID
	}
	if req.ModelName != nil {
		agent.ModelName = *req.ModelName
	}
	if req.SystemPrompt != nil {
		agent.SystemPrompt = *req.SystemPrompt
	}
	if req.AgentType != nil {
		agent.AgentType = *req.AgentType
	}
	if req.AgentRole != nil {
		agent.AgentRole = *req.AgentRole
	}
	if req.MaxIterations != nil {
		agent.MaxIterations = *req.MaxIterations
	}
	if req.Temperature != nil {
		agent.Temperature = req.Temperature
	}
	if req.MaxTokens != nil {
		agent.MaxTokens = req.MaxTokens
	}
	if req.Config != nil {
		agent.Config = req.Config
	}
	if req.IsEnabled != nil {
		agent.IsEnabled = *req.IsEnabled
	}

	if err := b.store.Agents().Update(ctx, agent); err != nil {
		return nil, fmt.Errorf("update agent: %w", err)
	}

	// 如果有子 Agent，更新关系
	if req.SubAgentIDs != nil {
		if err := b.SetAgentRelations(ctx, id, req.SubAgentIDs); err != nil {
			return nil, fmt.Errorf("set agent relations: %w", err)
		}
	}

	return agent, nil
}

func (b *configBiz) DeleteAgent(ctx context.Context, id string) error {
	agent, err := b.store.Agents().Get(ctx, id)
	if err != nil {
		return err
	}

	// 内置 Agent 不能删除
	if agent.IsBuiltin {
		return fmt.Errorf("cannot delete builtin agent")
	}

	// 删除关联关系
	if err := b.store.AgentRelations().DeleteByParent(ctx, id); err != nil {
		return fmt.Errorf("delete agent relations: %w", err)
	}

	return b.store.Agents().Delete(ctx, id)
}

func (b *configBiz) ListBuiltinAgents(ctx context.Context) ([]*model.Agent, error) {
	return builtin.ListBuiltinAgents(), nil
}

func (b *configBiz) ListOrchestratorAgents(ctx context.Context) ([]*model.Agent, error) {
	// 内置主控 Agent
	builtinOrchestrators := builtin.ListOrchestratorAgents()

	// 数据库中的主控 Agent
	dbAgents, err := b.store.Agents().ListByRole(ctx, model.AgentRoleOrchestrator)
	if err != nil {
		return nil, err
	}

	return append(builtinOrchestrators, dbAgents...), nil
}

func (b *configBiz) ListSpecialistAgents(ctx context.Context) ([]*model.Agent, error) {
	// 内置子 Agent
	builtinSpecialists := builtin.ListSpecialistAgents()

	// 数据库中的子 Agent
	dbAgents, err := b.store.Agents().ListByRole(ctx, model.AgentRoleSpecialist)
	if err != nil {
		return nil, err
	}

	return append(builtinSpecialists, dbAgents...), nil
}

func (b *configBiz) GetAgentRelations(ctx context.Context, agentID string) ([]*model.AgentRelation, error) {
	return b.store.AgentRelations().ListByParentWithChild(ctx, agentID)
}

func (b *configBiz) SetAgentRelations(ctx context.Context, agentID string, subAgentIDs []string) error {
	// 删除现有关系
	if err := b.store.AgentRelations().DeleteByParent(ctx, agentID); err != nil {
		return fmt.Errorf("delete existing relations: %w", err)
	}

	// 创建新关系
	for i, subAgentID := range subAgentIDs {
		relation := &model.AgentRelation{
			ID:            uuid.New().String(),
			ParentAgentID: agentID,
			ChildAgentID:  subAgentID,
			Role:          "sub_agent",
			SortOrder:     i,
		}
		if err := b.store.AgentRelations().Create(ctx, relation); err != nil {
			return fmt.Errorf("create relation: %w", err)
		}
	}

	return nil
}

// ListAgentTools 列出 Agent 的工具.
func (b *configBiz) ListAgentTools(ctx context.Context, agentID string) ([]*model.AgentTool, error) {
	return b.store.AgentTools().ListByAgent(ctx, agentID)
}

// AddAgentToolRequest 添加 Agent 工具请求.
type AddAgentToolRequest struct {
	ToolType         model.ToolType `json:"tool_type"`
	MCPToolID        *string        `json:"mcp_tool_id,omitempty"`
	BuiltinToolName  string         `json:"builtin_tool_name,omitempty"`
	CustomToolConfig model.JSONMap  `json:"custom_tool_config,omitempty"`
	ReturnDirectly   bool           `json:"return_directly"`
	Priority         int            `json:"priority"`
}

// AddAgentTool 为 Agent 添加工具.
func (b *configBiz) AddAgentTool(ctx context.Context, agentID string, req *AddAgentToolRequest) (*model.AgentTool, error) {
	agentTool := &model.AgentTool{
		ID:               uuid.New().String(),
		AgentID:          agentID,
		ToolType:         req.ToolType,
		MCPToolID:        req.MCPToolID,
		BuiltinToolName:  req.BuiltinToolName,
		CustomToolConfig: req.CustomToolConfig,
		ReturnDirectly:   req.ReturnDirectly,
		IsEnabled:        true,
		Priority:         req.Priority,
	}

	if err := b.store.AgentTools().Create(ctx, agentTool); err != nil {
		return nil, fmt.Errorf("create agent tool: %w", err)
	}

	return agentTool, nil
}

// UpdateAgentToolRequest 更新 Agent 工具请求.
type UpdateAgentToolRequest struct {
	ReturnDirectly *bool `json:"return_directly,omitempty"`
	IsEnabled      *bool `json:"is_enabled,omitempty"`
	Priority       *int  `json:"priority,omitempty"`
}

// UpdateAgentTool 更新 Agent 工具.
func (b *configBiz) UpdateAgentTool(ctx context.Context, toolID string, req *UpdateAgentToolRequest) (*model.AgentTool, error) {
	agentTool, err := b.store.AgentTools().Get(ctx, toolID)
	if err != nil {
		return nil, err
	}

	if req.ReturnDirectly != nil {
		agentTool.ReturnDirectly = *req.ReturnDirectly
	}
	if req.IsEnabled != nil {
		agentTool.IsEnabled = *req.IsEnabled
	}
	if req.Priority != nil {
		agentTool.Priority = *req.Priority
	}

	if err := b.store.AgentTools().Update(ctx, agentTool); err != nil {
		return nil, fmt.Errorf("update agent tool: %w", err)
	}

	return agentTool, nil
}

// RemoveAgentTool 移除 Agent 工具.
func (b *configBiz) RemoveAgentTool(ctx context.Context, toolID string) error {
	return b.store.AgentTools().Delete(ctx, toolID)
}

// ListBuiltinTools 列出可用的内置工具.
func (b *configBiz) ListBuiltinTools(ctx context.Context) []string {
	return []string{
		"web_search",
		"web_fetch",
		"knowledge_search",
		"grep_chunks",
		"list_knowledge_chunks",
		"data_schema",
		"data_analysis",
	}
}
