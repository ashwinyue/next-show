// Package event 提供 Agent 事件发送辅助工具.
package event

import (
	"encoding/json"

	"github.com/cloudwego/eino/adk"
)

// Sender 事件发送器.
type Sender struct {
	gen *adk.AsyncGenerator[*adk.AgentEvent]
}

// NewSender 创建事件发送器.
func NewSender(gen *adk.AsyncGenerator[*adk.AgentEvent]) *Sender {
	return &Sender{gen: gen}
}

func mustJSON(v any) json.RawMessage {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return b
}

// SendThinking 发送思考事件.
func (s *Sender) SendThinking(thought string, step int, total int) {
	s.gen.Send(&adk.AgentEvent{
		Output: &adk.AgentOutput{
			CustomizedOutput: Envelope{
				Type: EventTypeThinking,
				Payload: mustJSON(ThinkingEvent{
					Thought:           thought,
					ThoughtNumber:     step,
					TotalThoughts:     total,
					NextThoughtNeeded: step < total,
				}),
			},
		},
	})
}

// SendThinkingSimple 发送简单思考事件.
func (s *Sender) SendThinkingSimple(thought string) {
	s.gen.Send(&adk.AgentEvent{
		Output: &adk.AgentOutput{
			CustomizedOutput: Envelope{
				Type:    EventTypeThinking,
				Content: thought,
			},
		},
	})
}

// SendReferences 发送知识引用事件.
func (s *Sender) SendReferences(chunks []ReferenceChunk) {
	s.gen.Send(&adk.AgentEvent{
		Output: &adk.AgentOutput{
			CustomizedOutput: Envelope{
				Type:    EventTypeReferences,
				Payload: mustJSON(ReferencesEvent{Chunks: chunks}),
			},
		},
	})
}

// SendReflection 发送反思事件.
func (s *Sender) SendReflection(reflection string, score int) {
	s.gen.Send(&adk.AgentEvent{
		Output: &adk.AgentOutput{
			CustomizedOutput: Envelope{
				Type:    EventTypeReflection,
				Payload: mustJSON(ReflectionEvent{Reflection: reflection, Score: score}),
			},
		},
	})
}

// SendCustom 发送自定义事件.
func (s *Sender) SendCustom(eventType string, content string, data map[string]interface{}) {
	s.gen.Send(&adk.AgentEvent{
		Output: &adk.AgentOutput{
			CustomizedOutput: Envelope{
				Type:    EventType(eventType),
				Content: content,
				Data:    data,
			},
		},
	})
}
