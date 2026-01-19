// Package mcp 提供 MCP 业务逻辑.
package mcp

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/ashwinyue/next-show/internal/model"
	"github.com/ashwinyue/next-show/internal/store"
)

// Biz MCP 业务接口.
type Biz interface {
	// ListServers 列出所有 MCP Server.
	ListServers(ctx context.Context) ([]*model.MCPServer, error)
	// GetServer 获取 MCP Server 详情.
	GetServer(ctx context.Context, id string) (*model.MCPServer, error)
	// CreateServer 创建 MCP Server.
	CreateServer(ctx context.Context, req *CreateServerRequest) (*model.MCPServer, error)
	// UpdateServer 更新 MCP Server.
	UpdateServer(ctx context.Context, id string, req *UpdateServerRequest) (*model.MCPServer, error)
	// DeleteServer 删除 MCP Server.
	DeleteServer(ctx context.Context, id string) error

	// ListTools 列出 MCP Server 的工具.
	ListTools(ctx context.Context, serverID string) ([]*model.MCPTool, error)
	// GetTool 获取 MCP Tool 详情.
	GetTool(ctx context.Context, id string) (*model.MCPTool, error)
	// CreateTool 创建 MCP Tool.
	CreateTool(ctx context.Context, serverID string, req *CreateToolRequest) (*model.MCPTool, error)
	// UpdateTool 更新 MCP Tool.
	UpdateTool(ctx context.Context, id string, req *UpdateToolRequest) (*model.MCPTool, error)
	// DeleteTool 删除 MCP Tool.
	DeleteTool(ctx context.Context, id string) error
}

// CreateServerRequest 创建 MCP Server 请求.
type CreateServerRequest struct {
	Name           string              `json:"name"`
	DisplayName    string              `json:"display_name"`
	Description    string              `json:"description"`
	TransportType  model.TransportType `json:"transport_type"`
	Command        string              `json:"command"`
	Args           model.JSONSlice     `json:"args"`
	Env            model.JSONMap       `json:"env"`
	ServerURL      string              `json:"server_url"`
	CustomHeaders  model.JSONMap       `json:"custom_headers"`
	TimeoutSeconds int                 `json:"timeout_seconds"`
}

// UpdateServerRequest 更新 MCP Server 请求.
type UpdateServerRequest struct {
	Name           *string              `json:"name,omitempty"`
	DisplayName    *string              `json:"display_name,omitempty"`
	Description    *string              `json:"description,omitempty"`
	TransportType  *model.TransportType `json:"transport_type,omitempty"`
	Command        *string              `json:"command,omitempty"`
	Args           model.JSONSlice      `json:"args,omitempty"`
	Env            model.JSONMap        `json:"env,omitempty"`
	ServerURL      *string              `json:"server_url,omitempty"`
	CustomHeaders  model.JSONMap        `json:"custom_headers,omitempty"`
	TimeoutSeconds *int                 `json:"timeout_seconds,omitempty"`
	IsEnabled      *bool                `json:"is_enabled,omitempty"`
}

// CreateToolRequest 创建 MCP Tool 请求.
type CreateToolRequest struct {
	Name           string        `json:"name"`
	DisplayName    string        `json:"display_name"`
	Description    string        `json:"description"`
	InputSchema    model.JSONMap `json:"input_schema"`
	ReturnDirectly bool          `json:"return_directly"`
}

// UpdateToolRequest 更新 MCP Tool 请求.
type UpdateToolRequest struct {
	Name           *string       `json:"name,omitempty"`
	DisplayName    *string       `json:"display_name,omitempty"`
	Description    *string       `json:"description,omitempty"`
	InputSchema    model.JSONMap `json:"input_schema,omitempty"`
	ReturnDirectly *bool         `json:"return_directly,omitempty"`
	IsEnabled      *bool         `json:"is_enabled,omitempty"`
}

type bizImpl struct {
	store store.Store
}

// NewBiz 创建 MCP 业务实例.
func NewBiz(s store.Store) Biz {
	return &bizImpl{store: s}
}

func (b *bizImpl) ListServers(ctx context.Context) ([]*model.MCPServer, error) {
	return b.store.MCPServers().List(ctx)
}

func (b *bizImpl) GetServer(ctx context.Context, id string) (*model.MCPServer, error) {
	return b.store.MCPServers().Get(ctx, id)
}

func (b *bizImpl) CreateServer(ctx context.Context, req *CreateServerRequest) (*model.MCPServer, error) {
	server := &model.MCPServer{
		ID:             uuid.New().String(),
		Name:           req.Name,
		DisplayName:    req.DisplayName,
		Description:    req.Description,
		TransportType:  req.TransportType,
		Command:        req.Command,
		Args:           req.Args,
		Env:            req.Env,
		ServerURL:      req.ServerURL,
		CustomHeaders:  req.CustomHeaders,
		TimeoutSeconds: req.TimeoutSeconds,
		IsEnabled:      true,
	}

	if server.TransportType == "" {
		server.TransportType = model.TransportTypeStdio
	}
	if server.TimeoutSeconds <= 0 {
		server.TimeoutSeconds = 30
	}

	if err := b.store.MCPServers().Create(ctx, server); err != nil {
		return nil, fmt.Errorf("create mcp server: %w", err)
	}

	return server, nil
}

func (b *bizImpl) UpdateServer(ctx context.Context, id string, req *UpdateServerRequest) (*model.MCPServer, error) {
	server, err := b.store.MCPServers().Get(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Name != nil {
		server.Name = *req.Name
	}
	if req.DisplayName != nil {
		server.DisplayName = *req.DisplayName
	}
	if req.Description != nil {
		server.Description = *req.Description
	}
	if req.TransportType != nil {
		server.TransportType = *req.TransportType
	}
	if req.Command != nil {
		server.Command = *req.Command
	}
	if req.Args != nil {
		server.Args = req.Args
	}
	if req.Env != nil {
		server.Env = req.Env
	}
	if req.ServerURL != nil {
		server.ServerURL = *req.ServerURL
	}
	if req.CustomHeaders != nil {
		server.CustomHeaders = req.CustomHeaders
	}
	if req.TimeoutSeconds != nil {
		server.TimeoutSeconds = *req.TimeoutSeconds
	}
	if req.IsEnabled != nil {
		server.IsEnabled = *req.IsEnabled
	}

	if err := b.store.MCPServers().Update(ctx, server); err != nil {
		return nil, fmt.Errorf("update mcp server: %w", err)
	}

	return server, nil
}

func (b *bizImpl) DeleteServer(ctx context.Context, id string) error {
	return b.store.MCPServers().Delete(ctx, id)
}

func (b *bizImpl) ListTools(ctx context.Context, serverID string) ([]*model.MCPTool, error) {
	return b.store.MCPTools().ListByServer(ctx, serverID)
}

func (b *bizImpl) GetTool(ctx context.Context, id string) (*model.MCPTool, error) {
	return b.store.MCPTools().Get(ctx, id)
}

func (b *bizImpl) CreateTool(ctx context.Context, serverID string, req *CreateToolRequest) (*model.MCPTool, error) {
	tool := &model.MCPTool{
		ID:             uuid.New().String(),
		MCPServerID:    serverID,
		Name:           req.Name,
		DisplayName:    req.DisplayName,
		Description:    req.Description,
		InputSchema:    req.InputSchema,
		ReturnDirectly: req.ReturnDirectly,
		IsEnabled:      true,
	}

	if err := b.store.MCPTools().Create(ctx, tool); err != nil {
		return nil, fmt.Errorf("create mcp tool: %w", err)
	}

	return tool, nil
}

func (b *bizImpl) UpdateTool(ctx context.Context, id string, req *UpdateToolRequest) (*model.MCPTool, error) {
	tool, err := b.store.MCPTools().Get(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Name != nil {
		tool.Name = *req.Name
	}
	if req.DisplayName != nil {
		tool.DisplayName = *req.DisplayName
	}
	if req.Description != nil {
		tool.Description = *req.Description
	}
	if req.InputSchema != nil {
		tool.InputSchema = req.InputSchema
	}
	if req.ReturnDirectly != nil {
		tool.ReturnDirectly = *req.ReturnDirectly
	}
	if req.IsEnabled != nil {
		tool.IsEnabled = *req.IsEnabled
	}

	if err := b.store.MCPTools().Update(ctx, tool); err != nil {
		return nil, fmt.Errorf("update mcp tool: %w", err)
	}

	return tool, nil
}

func (b *bizImpl) DeleteTool(ctx context.Context, id string) error {
	return b.store.MCPTools().Delete(ctx, id)
}
