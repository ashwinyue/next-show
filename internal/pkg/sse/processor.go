// Package sse 提供事件处理器.
package sse

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/schema"

	"github.com/ashwinyue/next-show/internal/pkg/agent/event"
)

// EventProcessor 统一事件处理器接口.
type EventProcessor interface {
	Process(evt *adk.AgentEvent, ctx *EventContext) ([]Event, error)
}

// EventContext 事件处理上下文.
type EventContext struct {
	SessionID     string
	MessageID     string
	AgentName     string
	RunPath       string
	ContentBuffer *strings.Builder
}

// ToolResultHandler 工具结果处理器接口.
type ToolResultHandler interface {
	CanHandle(toolName string) bool
	Handle(msg *schema.Message, ctx *EventContext) ([]Event, error)
}

// ToolCallHandler 工具调用处理器接口.
type ToolCallHandler interface {
	CanHandle(toolName string) bool
	Handle(tc schema.ToolCall, ctx *EventContext) ([]Event, error)
}

// DefaultEventProcessor 默认事件处理器.
type DefaultEventProcessor struct {
	toolResultHandlers []ToolResultHandler
	toolCallHandlers   []ToolCallHandler
}

// NewDefaultEventProcessor 创建默认事件处理器.
func NewDefaultEventProcessor() *DefaultEventProcessor {
	p := &DefaultEventProcessor{
		toolResultHandlers: make([]ToolResultHandler, 0),
		toolCallHandlers:   make([]ToolCallHandler, 0),
	}
	// 注册默认处理器
	p.RegisterToolResultHandler(&ThinkingToolResultHandler{})
	p.RegisterToolResultHandler(&TodoWriteToolResultHandler{})
	p.RegisterToolCallHandler(&ThinkToolCallHandler{})
	return p
}

// RegisterToolResultHandler 注册工具结果处理器.
func (p *DefaultEventProcessor) RegisterToolResultHandler(h ToolResultHandler) {
	p.toolResultHandlers = append(p.toolResultHandlers, h)
}

// RegisterToolCallHandler 注册工具调用处理器.
func (p *DefaultEventProcessor) RegisterToolCallHandler(h ToolCallHandler) {
	p.toolCallHandlers = append(p.toolCallHandlers, h)
}

// Process 处理 AgentEvent.
func (p *DefaultEventProcessor) Process(evt *adk.AgentEvent, ctx *EventContext) ([]Event, error) {
	if evt == nil {
		return nil, nil
	}

	if evt.AgentName != "" {
		ctx.AgentName = evt.AgentName
	}
	ctx.RunPath = formatRunPath(evt.RunPath)

	var events []Event

	// 1. 处理错误
	if evt.Err != nil {
		events = append(events, Event{
			Type:      EventTypeError,
			ID:        ctx.MessageID,
			AgentName: ctx.AgentName,
			RunPath:   ctx.RunPath,
			Error:     evt.Err.Error(),
			SessionID: ctx.SessionID,
		})
		return events, nil
	}

	// 2. 处理输出
	if evt.Output != nil {
		outputEvents, err := p.processOutput(evt.Output, ctx)
		if err != nil {
			return nil, err
		}
		events = append(events, outputEvents...)
	}

	// 3. 处理 Action
	if evt.Action != nil {
		actionEvents := p.processAction(evt.Action, ctx)
		events = append(events, actionEvents...)
	}

	return events, nil
}

func (p *DefaultEventProcessor) processOutput(output *adk.AgentOutput, ctx *EventContext) ([]Event, error) {
	var events []Event

	// 处理 CustomizedOutput
	if output.CustomizedOutput != nil {
		customEvents, err := p.processCustomizedOutput(output.CustomizedOutput, ctx)
		if err != nil {
			return nil, err
		}
		events = append(events, customEvents...)
	}

	// 处理 MessageOutput
	if output.MessageOutput != nil && output.MessageOutput.Message != nil {
		msgEvents, err := p.processMessage(output.MessageOutput.Message, ctx)
		if err != nil {
			return nil, err
		}
		events = append(events, msgEvents...)
	}

	return events, nil
}

func (p *DefaultEventProcessor) processCustomizedOutput(customOutput any, ctx *EventContext) ([]Event, error) {
	switch v := customOutput.(type) {
	case event.Envelope:
		return p.processEnvelope(&v, ctx)
	case *event.Envelope:
		return p.processEnvelope(v, ctx)
	case string:
		return []Event{{
			Type:      EventTypeThinking,
			ID:        ctx.MessageID,
			Content:   v,
			AgentName: ctx.AgentName,
			RunPath:   ctx.RunPath,
			SessionID: ctx.SessionID,
		}}, nil
	}
	return nil, nil
}

func (p *DefaultEventProcessor) processEnvelope(env *event.Envelope, ctx *EventContext) ([]Event, error) {
	if env == nil {
		return nil, nil
	}

	switch env.Type {
	case event.EventTypeThinking:
		return p.processThinkingEnvelope(env, ctx)
	case event.EventTypeReferences:
		return p.processReferencesEnvelope(env, ctx)
	case event.EventTypeReflection:
		return p.processReflectionEnvelope(env, ctx)
	default:
		data := make(map[string]interface{})
		for k, v := range env.Data {
			data[k] = v
		}
		return []Event{{
			Type:      EventType(env.Type),
			ID:        ctx.MessageID,
			Content:   env.Content,
			AgentName: ctx.AgentName,
			RunPath:   ctx.RunPath,
			SessionID: ctx.SessionID,
			Data:      data,
		}}, nil
	}
}

func (p *DefaultEventProcessor) processThinkingEnvelope(env *event.Envelope, ctx *EventContext) ([]Event, error) {
	content := env.Content
	data := make(map[string]interface{})
	var done bool

	if len(env.Payload) > 0 {
		var payload event.ThinkingEvent
		if err := json.Unmarshal(env.Payload, &payload); err == nil {
			if content == "" {
				content = payload.Thought
			}
			done = !payload.NextThoughtNeeded && payload.ThoughtNumber >= payload.TotalThoughts
			data["iteration"] = payload.ThoughtNumber
			data["done"] = done
			data["thought_number"] = payload.ThoughtNumber
			data["total_thoughts"] = payload.TotalThoughts
		}
	}

	return []Event{{
		Type:      EventTypeThinking,
		ID:        ctx.MessageID,
		Content:   content,
		Done:      done,
		AgentName: ctx.AgentName,
		RunPath:   ctx.RunPath,
		SessionID: ctx.SessionID,
		Data:      data,
	}}, nil
}

func (p *DefaultEventProcessor) processReferencesEnvelope(env *event.Envelope, ctx *EventContext) ([]Event, error) {
	data := make(map[string]interface{})
	if len(env.Payload) > 0 {
		var payload event.ReferencesEvent
		if err := json.Unmarshal(env.Payload, &payload); err == nil {
			data["chunks"] = payload.Chunks
		}
	}

	return []Event{{
		Type:      EventTypeReferences,
		ID:        ctx.MessageID,
		AgentName: ctx.AgentName,
		RunPath:   ctx.RunPath,
		SessionID: ctx.SessionID,
		Data:      data,
	}}, nil
}

func (p *DefaultEventProcessor) processReflectionEnvelope(env *event.Envelope, ctx *EventContext) ([]Event, error) {
	data := make(map[string]interface{})
	if len(env.Payload) > 0 {
		var payload event.ReflectionEvent
		if err := json.Unmarshal(env.Payload, &payload); err == nil {
			data["reflection"] = payload.Reflection
			data["score"] = payload.Score
		}
	}

	return []Event{{
		Type:      EventTypeReflection,
		ID:        ctx.MessageID,
		AgentName: ctx.AgentName,
		RunPath:   ctx.RunPath,
		SessionID: ctx.SessionID,
		Data:      data,
	}}, nil
}

func (p *DefaultEventProcessor) processMessage(msg *schema.Message, ctx *EventContext) ([]Event, error) {
	var events []Event

	// 检查工具调用拦截
	for _, tc := range msg.ToolCalls {
		for _, h := range p.toolCallHandlers {
			if h.CanHandle(tc.Function.Name) {
				handled, err := h.Handle(tc, ctx)
				if err != nil {
					return nil, err
				}
				events = append(events, handled...)
				return events, nil
			}
		}
	}

	// 处理工具结果
	if msg.Role == schema.Tool {
		for _, h := range p.toolResultHandlers {
			if h.CanHandle(msg.ToolName) {
				return h.Handle(msg, ctx)
			}
		}
		return []Event{{
			Type:      EventTypeToolResult,
			ID:        ctx.MessageID,
			Content:   msg.Content,
			AgentName: ctx.AgentName,
			RunPath:   ctx.RunPath,
			SessionID: ctx.SessionID,
			Data: map[string]interface{}{
				"tool_call_id": msg.ToolCallID,
				"tool_name":    msg.ToolName,
				"success":      true,
				"output":       msg.Content,
			},
		}}, nil
	}

	// 处理 ToolCalls
	if len(msg.ToolCalls) > 0 {
		return []Event{{
			Type:      EventTypeToolCall,
			ID:        ctx.MessageID,
			AgentName: ctx.AgentName,
			RunPath:   ctx.RunPath,
			SessionID: ctx.SessionID,
			ToolCalls: msg.ToolCalls,
		}}, nil
	}

	// 处理 Assistant 消息
	if msg.Role == schema.Assistant && msg.Content != "" {
		if ctx.ContentBuffer != nil {
			ctx.ContentBuffer.WriteString(msg.Content)
		}
		return []Event{{
			Type:      EventTypeAnswer,
			ID:        ctx.MessageID,
			Content:   msg.Content,
			AgentName: ctx.AgentName,
			RunPath:   ctx.RunPath,
			SessionID: ctx.SessionID,
		}}, nil
	}

	return events, nil
}

func (p *DefaultEventProcessor) processAction(action *adk.AgentAction, ctx *EventContext) []Event {
	var events []Event

	if action.TransferToAgent != nil {
		events = append(events, Event{
			Type:       EventTypeAction,
			ID:         ctx.MessageID,
			AgentName:  ctx.AgentName,
			RunPath:    ctx.RunPath,
			SessionID:  ctx.SessionID,
			ActionType: "transfer",
			Data: map[string]interface{}{
				"target_agent": action.TransferToAgent.DestAgentName,
			},
		})
	}

	if action.Interrupted != nil {
		var content string
		if stringer, ok := action.Interrupted.Data.(fmt.Stringer); ok {
			content = stringer.String()
		}
		events = append(events, Event{
			Type:       EventTypeAction,
			ID:         ctx.MessageID,
			Content:    content,
			AgentName:  ctx.AgentName,
			RunPath:    ctx.RunPath,
			SessionID:  ctx.SessionID,
			ActionType: "interrupted",
		})
	}

	if action.Exit {
		events = append(events, Event{
			Type:       EventTypeAction,
			ID:         ctx.MessageID,
			AgentName:  ctx.AgentName,
			RunPath:    ctx.RunPath,
			SessionID:  ctx.SessionID,
			ActionType: "exit",
		})
	}

	return events
}

func formatRunPath(runPath []adk.RunStep) string {
	if len(runPath) == 0 {
		return ""
	}
	var parts []string
	for _, step := range runPath {
		parts = append(parts, step.String())
	}
	return strings.Join(parts, " -> ")
}

// ============================================================================
// 工具处理器实现
// ============================================================================

// ThinkToolCallHandler 处理 think 工具调用.
type ThinkToolCallHandler struct{}

func (h *ThinkToolCallHandler) CanHandle(toolName string) bool {
	return toolName == "think"
}

func (h *ThinkToolCallHandler) Handle(tc schema.ToolCall, ctx *EventContext) ([]Event, error) {
	var args struct {
		Thought           string `json:"thought"`
		ThoughtNumber     int    `json:"thought_number"`
		TotalThoughts     int    `json:"total_thoughts"`
		NextThoughtNeeded bool   `json:"next_thought_needed"`
	}
	if err := json.Unmarshal([]byte(tc.Function.Arguments), &args); err != nil {
		return []Event{{
			Type:      EventTypeToolCall,
			ID:        ctx.MessageID,
			AgentName: ctx.AgentName,
			RunPath:   ctx.RunPath,
			SessionID: ctx.SessionID,
			ToolCalls: []schema.ToolCall{tc},
		}}, nil
	}

	done := !args.NextThoughtNeeded && args.ThoughtNumber >= args.TotalThoughts
	return []Event{{
		Type:      EventTypeThinking,
		ID:        ctx.MessageID,
		Content:   args.Thought,
		Done:      done,
		AgentName: ctx.AgentName,
		RunPath:   ctx.RunPath,
		SessionID: ctx.SessionID,
		Data: map[string]interface{}{
			"iteration":      args.ThoughtNumber,
			"done":           done,
			"thought_number": args.ThoughtNumber,
			"total_thoughts": args.TotalThoughts,
		},
	}}, nil
}

// ThinkingToolResultHandler 处理 thinking 工具结果.
type ThinkingToolResultHandler struct{}

func (h *ThinkingToolResultHandler) CanHandle(toolName string) bool {
	return toolName == "thinking" || toolName == "think"
}

func (h *ThinkingToolResultHandler) Handle(msg *schema.Message, ctx *EventContext) ([]Event, error) {
	var result struct {
		Thought           string `json:"thought"`
		ThoughtNumber     int    `json:"thought_number"`
		TotalThoughts     int    `json:"total_thoughts"`
		NextThoughtNeeded bool   `json:"next_thought_needed"`
	}
	if err := json.Unmarshal([]byte(msg.Content), &result); err != nil {
		return []Event{{
			Type:      EventTypeToolResult,
			ID:        ctx.MessageID,
			Content:   msg.Content,
			AgentName: ctx.AgentName,
			RunPath:   ctx.RunPath,
			SessionID: ctx.SessionID,
			Data: map[string]interface{}{
				"tool_call_id": msg.ToolCallID,
				"tool_name":    msg.ToolName,
			},
		}}, nil
	}

	done := !result.NextThoughtNeeded && result.ThoughtNumber >= result.TotalThoughts
	return []Event{{
		Type:      EventTypeThinking,
		ID:        ctx.MessageID,
		Content:   result.Thought,
		Done:      done,
		AgentName: ctx.AgentName,
		RunPath:   ctx.RunPath,
		SessionID: ctx.SessionID,
		Data: map[string]interface{}{
			"iteration":      result.ThoughtNumber,
			"thought_number": result.ThoughtNumber,
			"total_thoughts": result.TotalThoughts,
		},
	}}, nil
}

// TodoWriteToolResultHandler 处理 todo_write 工具结果.
type TodoWriteToolResultHandler struct{}

func (h *TodoWriteToolResultHandler) CanHandle(toolName string) bool {
	return toolName == "todo_write"
}

func (h *TodoWriteToolResultHandler) Handle(msg *schema.Message, ctx *EventContext) ([]Event, error) {
	var result struct {
		PlanID string `json:"plan_id"`
		Task   string `json:"task"`
		Steps  []struct {
			ID          string `json:"id"`
			Description string `json:"description"`
			Status      string `json:"status"`
		} `json:"steps"`
	}

	var steps []PlanStep
	if err := json.Unmarshal([]byte(msg.Content), &result); err == nil && result.Task != "" {
		for _, s := range result.Steps {
			steps = append(steps, PlanStep{
				ID:          s.ID,
				Description: s.Description,
				Status:      s.Status,
			})
		}
		return []Event{{
			Type:      EventTypeToolResult,
			ID:        ctx.MessageID,
			Content:   msg.Content,
			AgentName: ctx.AgentName,
			RunPath:   ctx.RunPath,
			SessionID: ctx.SessionID,
			Data: map[string]interface{}{
				"tool_call_id": msg.ToolCallID,
				"tool_name":    msg.ToolName,
				"display_type": "plan",
				"task":         result.Task,
				"steps":        steps,
				"total_steps":  len(steps),
			},
		}}, nil
	}

	return []Event{{
		Type:      EventTypeToolResult,
		ID:        ctx.MessageID,
		Content:   msg.Content,
		AgentName: ctx.AgentName,
		RunPath:   ctx.RunPath,
		SessionID: ctx.SessionID,
		Data: map[string]interface{}{
			"tool_call_id": msg.ToolCallID,
			"tool_name":    msg.ToolName,
		},
	}}, nil
}
