// Package event 提供 Agent 事件类型定义.
package event

import (
	"encoding/gob"
	"encoding/json"
)

// EventType 事件类型.
type EventType string

const (
	EventTypeThinking   EventType = "thinking"
	EventTypeReferences EventType = "references"
	EventTypeReflection EventType = "reflection"
)

// Envelope 统一事件信封.
// 所有自定义事件都通过此结构传递，支持 gob 序列化（兼容 checkpoint）.
type Envelope struct {
	Type    EventType       `json:"type"`
	Content string          `json:"content,omitempty"`
	Payload json.RawMessage `json:"payload,omitempty"`
	Data    map[string]any  `json:"data,omitempty"`
}

func init() {
	gob.Register(Envelope{})
}

// ThinkingEvent 思考事件数据.
type ThinkingEvent struct {
	Thought           string `json:"thought"`
	ThoughtNumber     int    `json:"thought_number,omitempty"`
	TotalThoughts     int    `json:"total_thoughts,omitempty"`
	NextThoughtNeeded bool   `json:"next_thought_needed,omitempty"`
}

// ReferencesEvent 知识引用事件数据.
type ReferencesEvent struct {
	Chunks []ReferenceChunk `json:"chunks"`
}

// ReferenceChunk 知识分块引用.
type ReferenceChunk struct {
	ID          string  `json:"id"`
	Content     string  `json:"content"`
	Score       float64 `json:"score"`
	KnowledgeID string  `json:"knowledge_id,omitempty"`
}

// ReflectionEvent 反思事件数据.
type ReflectionEvent struct {
	Reflection string `json:"reflection"`
	Score      int    `json:"score,omitempty"`
}
