// Package store 提供数据访问层.
package store

import (
	"context"

	"gorm.io/gorm"

	"github.com/ashwinyue/next-show/internal/model"
)

// MCPServerStore MCP Server 存储接口.
type MCPServerStore interface {
	Create(ctx context.Context, server *model.MCPServer) error
	Get(ctx context.Context, id string) (*model.MCPServer, error)
	GetByName(ctx context.Context, name string) (*model.MCPServer, error)
	Update(ctx context.Context, server *model.MCPServer) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context) ([]*model.MCPServer, error)
	ListEnabled(ctx context.Context) ([]*model.MCPServer, error)
}

type mcpServerStore struct {
	db *gorm.DB
}

func newMCPServerStore(db *gorm.DB) MCPServerStore {
	return &mcpServerStore{db: db}
}

func (s *mcpServerStore) Create(ctx context.Context, server *model.MCPServer) error {
	return s.db.WithContext(ctx).Create(server).Error
}

func (s *mcpServerStore) Get(ctx context.Context, id string) (*model.MCPServer, error) {
	var server model.MCPServer
	if err := s.db.WithContext(ctx).Where("id = ?", id).First(&server).Error; err != nil {
		return nil, err
	}
	return &server, nil
}

func (s *mcpServerStore) GetByName(ctx context.Context, name string) (*model.MCPServer, error) {
	var server model.MCPServer
	if err := s.db.WithContext(ctx).Where("name = ?", name).First(&server).Error; err != nil {
		return nil, err
	}
	return &server, nil
}

func (s *mcpServerStore) Update(ctx context.Context, server *model.MCPServer) error {
	return s.db.WithContext(ctx).Save(server).Error
}

func (s *mcpServerStore) Delete(ctx context.Context, id string) error {
	return s.db.WithContext(ctx).Where("id = ?", id).Delete(&model.MCPServer{}).Error
}

func (s *mcpServerStore) List(ctx context.Context) ([]*model.MCPServer, error) {
	var servers []*model.MCPServer
	if err := s.db.WithContext(ctx).Find(&servers).Error; err != nil {
		return nil, err
	}
	return servers, nil
}

func (s *mcpServerStore) ListEnabled(ctx context.Context) ([]*model.MCPServer, error) {
	var servers []*model.MCPServer
	if err := s.db.WithContext(ctx).Where("is_enabled = ?", true).Find(&servers).Error; err != nil {
		return nil, err
	}
	return servers, nil
}

// MCPToolStore MCP Tool 存储接口.
type MCPToolStore interface {
	Create(ctx context.Context, tool *model.MCPTool) error
	Get(ctx context.Context, id string) (*model.MCPTool, error)
	Update(ctx context.Context, tool *model.MCPTool) error
	Delete(ctx context.Context, id string) error
	ListByServer(ctx context.Context, serverID string) ([]*model.MCPTool, error)
	ListEnabledByServer(ctx context.Context, serverID string) ([]*model.MCPTool, error)
}

type mcpToolStore struct {
	db *gorm.DB
}

func newMCPToolStore(db *gorm.DB) MCPToolStore {
	return &mcpToolStore{db: db}
}

func (s *mcpToolStore) Create(ctx context.Context, tool *model.MCPTool) error {
	return s.db.WithContext(ctx).Create(tool).Error
}

func (s *mcpToolStore) Get(ctx context.Context, id string) (*model.MCPTool, error) {
	var tool model.MCPTool
	if err := s.db.WithContext(ctx).Where("id = ?", id).First(&tool).Error; err != nil {
		return nil, err
	}
	return &tool, nil
}

func (s *mcpToolStore) Update(ctx context.Context, tool *model.MCPTool) error {
	return s.db.WithContext(ctx).Save(tool).Error
}

func (s *mcpToolStore) Delete(ctx context.Context, id string) error {
	return s.db.WithContext(ctx).Where("id = ?", id).Delete(&model.MCPTool{}).Error
}

func (s *mcpToolStore) ListByServer(ctx context.Context, serverID string) ([]*model.MCPTool, error) {
	var tools []*model.MCPTool
	if err := s.db.WithContext(ctx).Where("mcp_server_id = ?", serverID).Find(&tools).Error; err != nil {
		return nil, err
	}
	return tools, nil
}

func (s *mcpToolStore) ListEnabledByServer(ctx context.Context, serverID string) ([]*model.MCPTool, error) {
	var tools []*model.MCPTool
	if err := s.db.WithContext(ctx).Where("mcp_server_id = ? AND is_enabled = ?", serverID, true).Find(&tools).Error; err != nil {
		return nil, err
	}
	return tools, nil
}

// AgentToolStore Agent Tool 存储接口.
type AgentToolStore interface {
	Create(ctx context.Context, agentTool *model.AgentTool) error
	Get(ctx context.Context, id string) (*model.AgentTool, error)
	Update(ctx context.Context, agentTool *model.AgentTool) error
	Delete(ctx context.Context, id string) error
	ListByAgent(ctx context.Context, agentID string) ([]*model.AgentTool, error)
	ListEnabledByAgent(ctx context.Context, agentID string) ([]*model.AgentTool, error)
}

type agentToolStore struct {
	db *gorm.DB
}

func newAgentToolStore(db *gorm.DB) AgentToolStore {
	return &agentToolStore{db: db}
}

func (s *agentToolStore) Create(ctx context.Context, agentTool *model.AgentTool) error {
	return s.db.WithContext(ctx).Create(agentTool).Error
}

func (s *agentToolStore) Get(ctx context.Context, id string) (*model.AgentTool, error) {
	var agentTool model.AgentTool
	if err := s.db.WithContext(ctx).Where("id = ?", id).First(&agentTool).Error; err != nil {
		return nil, err
	}
	return &agentTool, nil
}

func (s *agentToolStore) Update(ctx context.Context, agentTool *model.AgentTool) error {
	return s.db.WithContext(ctx).Save(agentTool).Error
}

func (s *agentToolStore) Delete(ctx context.Context, id string) error {
	return s.db.WithContext(ctx).Where("id = ?", id).Delete(&model.AgentTool{}).Error
}

func (s *agentToolStore) ListByAgent(ctx context.Context, agentID string) ([]*model.AgentTool, error) {
	var agentTools []*model.AgentTool
	if err := s.db.WithContext(ctx).Where("agent_id = ?", agentID).Order("priority DESC").Find(&agentTools).Error; err != nil {
		return nil, err
	}
	return agentTools, nil
}

func (s *agentToolStore) ListEnabledByAgent(ctx context.Context, agentID string) ([]*model.AgentTool, error) {
	var agentTools []*model.AgentTool
	if err := s.db.WithContext(ctx).Where("agent_id = ? AND is_enabled = ?", agentID, true).Order("priority DESC").Find(&agentTools).Error; err != nil {
		return nil, err
	}
	return agentTools, nil
}
