// Package factory 提供 Agent 和 Provider 工厂.
package factory

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino-ext/components/tool/mcp"
	"github.com/cloudwego/eino/components/tool"
	mcpclient "github.com/mark3labs/mcp-go/client"

	modelDef "github.com/ashwinyue/next-show/internal/model"
	"github.com/ashwinyue/next-show/internal/store"
)

// MCPToolFactory MCP 工具工厂.
type MCPToolFactory struct {
	store   store.Store
	clients map[string]mcpclient.MCPClient // serverID -> client
}

// NewMCPToolFactory 创建 MCP 工具工厂.
func NewMCPToolFactory(s store.Store) *MCPToolFactory {
	return &MCPToolFactory{
		store:   s,
		clients: make(map[string]mcpclient.MCPClient),
	}
}

// GetToolsForAgent 获取 Agent 关联的 MCP 工具.
func (f *MCPToolFactory) GetToolsForAgent(ctx context.Context, agentID string) ([]tool.BaseTool, error) {
	// 获取 Agent 关联的 MCP 工具配置
	agentTools, err := f.store.AgentTools().ListEnabledByAgent(ctx, agentID)
	if err != nil {
		return nil, fmt.Errorf("failed to list agent tools: %w", err)
	}

	var tools []tool.BaseTool
	for _, at := range agentTools {
		if at.ToolType != modelDef.ToolTypeMCP || at.MCPToolID == nil || *at.MCPToolID == "" {
			continue
		}

		// 获取 MCP Tool 配置
		mcpTool, err := f.store.MCPTools().Get(ctx, *at.MCPToolID)
		if err != nil {
			continue // 跳过无法获取的工具
		}

		// 获取 MCP Server
		server, err := f.store.MCPServers().Get(ctx, mcpTool.MCPServerID)
		if err != nil {
			continue
		}

		// 获取或创建 MCP Client
		cli, err := f.getOrCreateClient(ctx, server)
		if err != nil {
			continue
		}

		// 获取指定工具
		mcpTools, err := mcp.GetTools(ctx, &mcp.Config{
			Cli:          cli,
			ToolNameList: []string{mcpTool.Name},
		})
		if err != nil {
			continue
		}

		tools = append(tools, mcpTools...)
	}

	return tools, nil
}

// GetToolsFromServer 从 MCP Server 获取所有工具.
func (f *MCPToolFactory) GetToolsFromServer(ctx context.Context, serverID string) ([]tool.BaseTool, error) {
	server, err := f.store.MCPServers().Get(ctx, serverID)
	if err != nil {
		return nil, fmt.Errorf("failed to get mcp server: %w", err)
	}

	cli, err := f.getOrCreateClient(ctx, server)
	if err != nil {
		return nil, err
	}

	// 获取该 Server 启用的工具列表
	enabledTools, err := f.store.MCPTools().ListEnabledByServer(ctx, serverID)
	if err != nil {
		return nil, err
	}

	var toolNames []string
	for _, t := range enabledTools {
		toolNames = append(toolNames, t.Name)
	}

	return mcp.GetTools(ctx, &mcp.Config{
		Cli:          cli,
		ToolNameList: toolNames,
	})
}

// getOrCreateClient 获取或创建 MCP Client.
func (f *MCPToolFactory) getOrCreateClient(ctx context.Context, server *modelDef.MCPServer) (mcpclient.MCPClient, error) {
	if cli, ok := f.clients[server.ID]; ok {
		return cli, nil
	}

	cli, err := f.createClient(ctx, server)
	if err != nil {
		return nil, err
	}

	f.clients[server.ID] = cli
	return cli, nil
}

// createClient 创建 MCP Client.
func (f *MCPToolFactory) createClient(ctx context.Context, server *modelDef.MCPServer) (mcpclient.MCPClient, error) {
	switch server.TransportType {
	case modelDef.TransportTypeSSE:
		return f.createSSEClient(ctx, server)
	case modelDef.TransportTypeStdio:
		return f.createStdioClient(ctx, server)
	default:
		return nil, fmt.Errorf("unsupported transport type: %s", server.TransportType)
	}
}

// createSSEClient 创建 SSE 传输的 MCP Client.
func (f *MCPToolFactory) createSSEClient(ctx context.Context, server *modelDef.MCPServer) (mcpclient.MCPClient, error) {
	if server.ServerURL == "" {
		return nil, fmt.Errorf("server url is required for SSE transport")
	}

	cli, err := mcpclient.NewSSEMCPClient(server.ServerURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create SSE MCP client: %w", err)
	}

	// 初始化连接
	if err := cli.Start(ctx); err != nil {
		return nil, fmt.Errorf("failed to start SSE MCP client: %w", err)
	}

	return cli, nil
}

// createStdioClient 创建 Stdio 传输的 MCP Client.
func (f *MCPToolFactory) createStdioClient(ctx context.Context, server *modelDef.MCPServer) (mcpclient.MCPClient, error) {
	if server.Command == "" {
		return nil, fmt.Errorf("command is required for Stdio transport")
	}

	cli, err := mcpclient.NewStdioMCPClient(server.Command, nil, server.Args...)
	if err != nil {
		return nil, fmt.Errorf("failed to create Stdio MCP client: %w", err)
	}

	// 初始化连接
	if err := cli.Start(ctx); err != nil {
		return nil, fmt.Errorf("failed to start Stdio MCP client: %w", err)
	}

	return cli, nil
}

// Close 关闭所有 MCP Client.
func (f *MCPToolFactory) Close() error {
	for _, cli := range f.clients {
		if closer, ok := cli.(interface{ Close() error }); ok {
			_ = closer.Close()
		}
	}
	f.clients = make(map[string]mcpclient.MCPClient)
	return nil
}
