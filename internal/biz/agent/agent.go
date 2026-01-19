// Package agent 提供 Agent 业务逻辑.
package agent

import (
	"context"
	"sync"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/schema"

	"github.com/ashwinyue/next-show/internal/pkg/agent/factory"
	"github.com/ashwinyue/next-show/internal/store"
)

// AgentBiz Agent 业务接口.
type AgentBiz interface {
	// Chat 执行 Agent 对话，返回事件迭代器.
	Chat(ctx context.Context, sessionID string, messages []adk.Message) (*adk.AsyncIterator[*adk.AgentEvent], error)
	// Close 关闭业务层，清理资源.
	Close()
}

type agentBiz struct {
	store        store.Store
	agentFactory *factory.AgentFactory
	runners      map[string]*adk.Runner // agentID -> Runner 缓存
	mu           sync.RWMutex
}

// NewAgentBiz 创建 Agent 业务实例.
func NewAgentBiz(s store.Store, af *factory.AgentFactory) AgentBiz {
	return &agentBiz{
		store:        s,
		agentFactory: af,
		runners:      make(map[string]*adk.Runner),
	}
}

// getOrCreateRunner 获取或创建 Agent 的 Runner.
func (b *agentBiz) getOrCreateRunner(ctx context.Context, agentID string) (*adk.Runner, error) {
	b.mu.RLock()
	if runner, ok := b.runners[agentID]; ok {
		b.mu.RUnlock()
		return runner, nil
	}
	b.mu.RUnlock()

	b.mu.Lock()
	defer b.mu.Unlock()

	// 双重检查
	if runner, ok := b.runners[agentID]; ok {
		return runner, nil
	}

	// 创建新 Runner
	runner, err := b.agentFactory.CreateRunner(ctx, agentID)
	if err != nil {
		return nil, err
	}
	b.runners[agentID] = runner
	return runner, nil
}

// Chat 执行 Agent 对话.
func (b *agentBiz) Chat(ctx context.Context, sessionID string, messages []adk.Message) (*adk.AsyncIterator[*adk.AgentEvent], error) {
	// 获取 Session 信息
	session, err := b.store.Sessions().Get(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	// 获取或创建 Runner
	runner, err := b.getOrCreateRunner(ctx, session.AgentID)
	if err != nil {
		return nil, err
	}

	// 获取历史消息
	historyMsgs, err := b.store.Messages().ListBySession(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	// 转换历史消息为 ADK Message
	var allMessages []adk.Message
	for _, m := range historyMsgs {
		var msg *schema.Message
		switch m.Role {
		case "user":
			msg = schema.UserMessage(m.Content)
		case "assistant":
			msg = schema.AssistantMessage(m.Content, nil)
		default:
			continue
		}
		allMessages = append(allMessages, msg)
	}
	allMessages = append(allMessages, messages...)

	// 运行 Agent
	iter := runner.Run(ctx, allMessages)
	return iter, nil
}

// Close 关闭业务层，清理资源.
func (b *agentBiz) Close() {
	if b.agentFactory != nil {
		b.agentFactory.Close()
	}
}

// AgentError Agent 错误.
type AgentError struct {
	Message string
}

func (e *AgentError) Error() string {
	return e.Message
}
