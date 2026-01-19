// Package factory 提供 Agent 和 Provider 工厂.
package factory

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool"

	modelDef "github.com/mervyn/next-show/internal/model"
	"github.com/mervyn/next-show/internal/store"
)

// AgentFactory Agent 工厂.
type AgentFactory struct {
	store            store.Store
	chatModelFactory *ChatModelFactory
}

// NewAgentFactory 创建 Agent 工厂.
func NewAgentFactory(s store.Store) *AgentFactory {
	return &AgentFactory{
		store:            s,
		chatModelFactory: NewChatModelFactory(),
	}
}

// CreateRunner 根据 Agent ID 创建 ADK Runner.
func (f *AgentFactory) CreateRunner(ctx context.Context, agentID string) (*adk.Runner, error) {
	// 获取 Agent 配置（包含 Provider）
	agent, err := f.store.Agents().GetWithProvider(ctx, agentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get agent: %w", err)
	}

	if agent.Provider == nil {
		// 单独获取 Provider
		provider, err := f.store.Providers().Get(ctx, agent.ProviderID)
		if err != nil {
			return nil, fmt.Errorf("failed to get provider: %w", err)
		}
		agent.Provider = provider
	}

	// 创建 ChatModel
	chatModel, err := f.chatModelFactory.CreateChatModel(ctx, agent.Provider, agent.ModelName)
	if err != nil {
		return nil, fmt.Errorf("failed to create chat model: %w", err)
	}

	// 根据 Agent 类型创建不同的 Agent
	var adkAgent adk.Agent
	switch agent.AgentType {
	case modelDef.AgentTypeChatModel:
		adkAgent, err = f.createChatModelAgent(ctx, agent, chatModel)
	case modelDef.AgentTypeReact:
		adkAgent, err = f.createChatModelAgent(ctx, agent, chatModel) // 先用 ChatModel，后续支持 ReAct
	default:
		adkAgent, err = f.createChatModelAgent(ctx, agent, chatModel)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create agent: %w", err)
	}

	// 创建 Runner
	runner := adk.NewRunner(ctx, adk.RunnerConfig{
		Agent:           adkAgent,
		EnableStreaming: true,
	})
	return runner, nil
}

// createChatModelAgent 创建基础 ChatModel Agent.
func (f *AgentFactory) createChatModelAgent(ctx context.Context, agent *modelDef.Agent, chatModel model.ToolCallingChatModel) (adk.Agent, error) {
	// 加载工具
	tools, err := f.loadAgentTools(ctx, agent.ID)
	if err != nil {
		return nil, err
	}

	cfg := &adk.ChatModelAgentConfig{
		Name:          agent.Name,
		Description:   agent.Description,
		Instruction:   agent.SystemPrompt,
		Model:         chatModel,
		MaxIterations: agent.MaxIterations,
	}

	// 如果有工具则配置 ToolsConfig
	if len(tools) > 0 {
		cfg.ToolsConfig.Tools = tools
	}

	return adk.NewChatModelAgent(ctx, cfg)
}

// loadAgentTools 加载 Agent 关联的工具.
func (f *AgentFactory) loadAgentTools(ctx context.Context, agentID string) ([]tool.BaseTool, error) {
	agentTools, err := f.store.AgentTools().ListEnabledByAgent(ctx, agentID)
	if err != nil {
		return nil, err
	}

	var tools []tool.BaseTool
	for _, at := range agentTools {
		switch at.ToolType {
		case modelDef.ToolTypeBuiltin:
			// TODO: 加载内置工具
			t := f.loadBuiltinTool(at.BuiltinToolName)
			if t != nil {
				tools = append(tools, t)
			}
		case modelDef.ToolTypeMCP:
			// TODO: 加载 MCP 工具
		case modelDef.ToolTypeCustom:
			// TODO: 加载自定义工具
		}
	}

	return tools, nil
}

// loadBuiltinTool 加载内置工具.
func (f *AgentFactory) loadBuiltinTool(name string) tool.BaseTool {
	// TODO: 实现内置工具注册表
	return nil
}
