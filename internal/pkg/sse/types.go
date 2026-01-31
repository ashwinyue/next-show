// Package sse 提供 SSE（Server-Sent Events）协议封装.
package sse

// EventType SSE 事件类型（对齐 WeKnora）.
type EventType string

const (
	// EventTypeQuery 查询开始事件
	EventTypeQuery EventType = "agent_query"
	// EventTypeAnswer Agent 最终答案（流式）
	EventTypeAnswer EventType = "answer"
	// EventTypeThinking Agent 思考过程
	EventTypeThinking EventType = "thinking"
	// EventTypeToolCall 工具调用
	EventTypeToolCall EventType = "tool_call"
	// EventTypeToolResult 工具执行结果
	EventTypeToolResult EventType = "tool_result"
	// EventTypeComplete 完成事件
	EventTypeComplete EventType = "stop"
	// EventTypeError 错误事件
	EventTypeError EventType = "error"
)

// Event SSE 事件结构（对齐 WeKnora）.
type Event struct {
	Type               EventType              `json:"response_type"`
	ID                 string                 `json:"id"`
	Content            string                 `json:"content,omitempty"`
	Done               bool                   `json:"done,omitempty"`
	AgentName          string                 `json:"agent_name,omitempty"`
	RunPath            string                 `json:"run_path,omitempty"`
	ToolCalls          interface{}            `json:"tool_calls,omitempty"`
	ActionType         string                 `json:"action_type,omitempty"`
	Error              string                 `json:"error,omitempty"`
	SessionID          string                 `json:"session_id,omitempty"`
	Data               map[string]interface{} `json:"data,omitempty"`
	AssistantMessageID string                 `json:"assistant_message_id,omitempty"`
}
