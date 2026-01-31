// Package agent 提供 Agent 业务逻辑.
package agent

import (
	"context"
	"fmt"
	"sync"

	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
	"github.com/cloudwego/eino/components/tool"

	"github.com/ashwinyue/next-show/internal/model"
	"github.com/ashwinyue/next-show/internal/pkg/agent/agentic"
	agentmodel "github.com/ashwinyue/next-show/internal/pkg/agent/model"
	"github.com/ashwinyue/next-show/internal/pkg/sse"
	"github.com/ashwinyue/next-show/internal/store"
)

// AgentBiz Agent 业务接口.
type AgentBiz interface {
	// Chat 执行 Agent 对话，通过 SSE writer 发送事件.
	Chat(ctx context.Context, sessionID string, content string, sseWriter sse.Writer) error
	// Close 关闭业务层，清理资源.
	Close()
}

type agentBiz struct {
	store   store.Store
	runners map[string]*agentic.Agent // agentID -> Agent 缓存
	mu      sync.RWMutex
}

// NewAgentBiz 创建 Agent 业务实例.
func NewAgentBiz(s store.Store) AgentBiz {
	return &agentBiz{
		store:   s,
		runners: make(map[string]*agentic.Agent),
	}
}

// getOrCreateAgent 获取或创建 Agent.
func (b *agentBiz) getOrCreateAgent(ctx context.Context, agent *model.Agent) (*agentic.Agent, error) {
	b.mu.RLock()
	if agentInst, ok := b.runners[agent.ID]; ok {
		b.mu.RUnlock()
		return agentInst, nil
	}
	b.mu.RUnlock()

	b.mu.Lock()
	defer b.mu.Unlock()

	// 双重检查
	if agentInst, ok := b.runners[agent.ID]; ok {
		return agentInst, nil
	}

	// 获取 Provider 配置
	provider, err := b.store.Providers().Get(ctx, agent.ProviderID)
	if err != nil {
		return nil, fmt.Errorf("get provider: %w", err)
	}

	// 创建 AgenticModel
	modelCfg := &agentmodel.ModelConfig{
		Provider: provider.Name,
		Model:    agent.ModelName,
		APIKey:   provider.APIKey,
		BaseURL:  provider.BaseURL,
	}

	agenticModel, err := agentmodel.CreateAgenticModel(ctx, modelCfg)
	if err != nil {
		return nil, fmt.Errorf("create agentic model: %w", err)
	}

	// 构建工具配置 - 简化版，暂时为空
	toolsConfig := compose.ToolsNodeConfig{
		Tools: []tool.BaseTool{},
	}

	// 创建 Agentic Agent
	maxStep := agent.MaxIterations
	if maxStep <= 0 {
		maxStep = 10
	}

	agentInst, err := agentic.NewAgent(ctx, &agentic.AgentConfig{
		Model:       agenticModel,
		ToolsConfig: toolsConfig,
		MaxStep:     maxStep,
	})
	if err != nil {
		return nil, fmt.Errorf("create agent: %w", err)
	}

	b.runners[agent.ID] = agentInst
	return agentInst, nil
}

// convertToAgenticMessages 转换消息为 AgenticMessage.
func convertToAgenticMessages(session *model.Session, content string) []*schema.AgenticMessage {
	messages := []*schema.AgenticMessage{}

	// 添加系统提示
	if session.Agent.SystemPrompt != "" {
		messages = append(messages, schema.SystemAgenticMessage(session.Agent.SystemPrompt))
	}

	// 添加用户消息
	messages = append(messages, schema.UserAgenticMessage(content))

	return messages
}

// Chat 执行 Agent 对话.
func (b *agentBiz) Chat(ctx context.Context, sessionID string, content string, sseWriter sse.Writer) error {
	// 获取 Session 信息
	session, err := b.store.Sessions().GetWithAgent(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("get session: %w", err)
	}

	// 发送开始事件
	if err := sseWriter.SendStart(session.ID, session.ID); err != nil {
		return err
	}

	// 获取或创建 Agent
	agentInst, err := b.getOrCreateAgent(ctx, session.Agent)
	if err != nil {
		sseWriter.SendError(err.Error())
		return err
	}

	// 创建 SSE 适配器
	adapter := sse.NewAgenticAdapter(sseWriter)

	// 创建 Callback
	cb := compose.WithCallbacks(adapter.NewCallback())

	// 转换消息为 AgenticMessage
	messages := convertToAgenticMessages(session, content)

	// 流式运行
	stream, err := agentInst.Stream(ctx, messages, cb)
	if err != nil {
		sseWriter.SendError(err.Error())
		return err
	}
	defer stream.Close()

	// 消费流（事件已在 adapter 中发送）
	for {
		_, err := stream.Recv()
		if err != nil {
			break
		}
	}

	return nil
}

// Close 关闭业务层，清理资源.
func (b *agentBiz) Close() {
	b.mu.Lock()
	defer b.mu.Unlock()

	// 清理所有 Agent
	b.runners = nil
}

// AgentError Agent 错误.
type AgentError struct {
	Message string
}

func (e *AgentError) Error() string {
	return e.Message
}
