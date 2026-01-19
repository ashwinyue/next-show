// Package agent 提供 Agent 业务逻辑.
package agent

import (
	"context"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/schema"

	"github.com/mervyn/next-show/internal/store"
)

// AgentBiz Agent 业务接口.
type AgentBiz interface {
	// Chat 执行 Agent 对话，返回事件迭代器.
	Chat(ctx context.Context, sessionID string, messages []adk.Message) (*adk.AsyncIterator[*adk.AgentEvent], error)
}

type agentBiz struct {
	store  store.Store
	runner *adk.Runner
}

// NewAgentBiz 创建 Agent 业务实例.
func NewAgentBiz(s store.Store) AgentBiz {
	return &agentBiz{
		store: s,
	}
}

// SetRunner 设置 ADK Runner（依赖注入）.
func (b *agentBiz) SetRunner(runner *adk.Runner) {
	b.runner = runner
}

// Chat 执行 Agent 对话.
func (b *agentBiz) Chat(ctx context.Context, sessionID string, messages []adk.Message) (*adk.AsyncIterator[*adk.AgentEvent], error) {
	if b.runner == nil {
		return nil, ErrRunnerNotConfigured
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
	iter := b.runner.Run(ctx, allMessages)
	return iter, nil
}

// ErrRunnerNotConfigured Runner 未配置错误.
var ErrRunnerNotConfigured = &AgentError{Message: "agent runner not configured"}

// AgentError Agent 错误.
type AgentError struct {
	Message string
}

func (e *AgentError) Error() string {
	return e.Message
}
