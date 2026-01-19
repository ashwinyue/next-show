// Package model 定义数据模型.
package model

import "time"

// MessageRole 消息角色.
type MessageRole string

const (
	MessageRoleUser      MessageRole = "user"
	MessageRoleAssistant MessageRole = "assistant"
	MessageRoleSystem    MessageRole = "system"
	MessageRoleTool      MessageRole = "tool"
)

// Message 消息模型（对应 eino schema.Message）.
type Message struct {
	ID               string      `json:"id" gorm:"primaryKey;size:36"`
	SessionID        string      `json:"session_id" gorm:"size:36;not null;index"`
	Role             MessageRole `json:"role" gorm:"size:20;not null;index"`
	Content          string      `json:"content" gorm:"type:text"`
	ReasoningContent string      `json:"reasoning_content,omitempty" gorm:"type:text"`
	Name             string      `json:"name,omitempty" gorm:"size:200"`

	// Tool 相关字段
	ToolCallID string  `json:"tool_call_id,omitempty" gorm:"size:100"`
	ToolName   string  `json:"tool_name,omitempty" gorm:"size:200"`
	ToolCalls  JSONMap `json:"tool_calls,omitempty" gorm:"type:json"`

	// 多模态内容
	MultiContent JSONMap `json:"multi_content,omitempty" gorm:"type:json"`

	// 响应元信息
	FinishReason string  `json:"finish_reason,omitempty" gorm:"size:50"`
	TokenUsage   JSONMap `json:"token_usage,omitempty" gorm:"type:json"`

	// 扩展字段
	Extra           JSONMap `json:"extra,omitempty" gorm:"type:json"`
	Sequence        int     `json:"sequence" gorm:"not null;index:idx_session_sequence"`
	ParentMessageID string  `json:"parent_message_id,omitempty" gorm:"size:36"`

	CreatedAt time.Time `json:"created_at" gorm:"index"`

	// 关联
	Session *Session `json:"session,omitempty" gorm:"foreignKey:SessionID"`
}

// TableName 返回表名.
func (Message) TableName() string {
	return "messages"
}

// AgentStep Agent 执行步骤（用于持久化）.
type AgentStep struct {
	Iteration int             `json:"iteration"`
	Thought   string          `json:"thought,omitempty"`
	ToolCalls []AgentToolCall `json:"tool_calls,omitempty"`
}

// AgentToolCall 工具调用记录.
type AgentToolCall struct {
	ID     string               `json:"id"`
	Name   string               `json:"name"`
	Args   string               `json:"args"`
	Result *AgentToolCallResult `json:"result,omitempty"`
}

// AgentToolCallResult 工具调用结果.
type AgentToolCallResult struct {
	Success bool                   `json:"success"`
	Output  string                 `json:"output"`
	Error   string                 `json:"error,omitempty"`
	Data    map[string]interface{} `json:"data,omitempty"`
}
