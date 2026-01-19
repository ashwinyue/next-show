// Package middleware 提供 Agent 中间件配置和构建.
package middleware

import (
	"strings"

	"github.com/cloudwego/eino/schema"
)

type messageGraph struct {
	toolCallToResult   map[string]int
	toolToAssistant    map[int]int
	assistantToolCalls map[int][]string
}

func buildMessageGraph(messages []*schema.Message) messageGraph {
	graph := messageGraph{
		toolCallToResult:   make(map[string]int),
		toolToAssistant:    make(map[int]int),
		assistantToolCalls: make(map[int][]string),
	}

	for i, msg := range messages {
		if msg.Role == schema.Assistant && len(msg.ToolCalls) > 0 {
			toolCallIDs := make([]string, 0, len(msg.ToolCalls))
			for _, tc := range msg.ToolCalls {
				toolCallIDs = append(toolCallIDs, tc.ID)
			}
			graph.assistantToolCalls[i] = toolCallIDs
		}
	}

	for i, msg := range messages {
		if msg.Role == schema.Tool && msg.ToolCallID != "" {
			for assistantIdx, toolCallIDs := range graph.assistantToolCalls {
				for _, tcID := range toolCallIDs {
					if tcID == msg.ToolCallID {
						graph.toolCallToResult[tcID] = i
						graph.toolToAssistant[i] = assistantIdx
						break
					}
				}
			}
		}
	}

	return graph
}

// TrimMessagesWithConsistency 裁剪消息并保持工具调用一致性.
func TrimMessagesWithConsistency(messages []*schema.Message, maxItems int) []*schema.Message {
	if len(messages) <= maxItems {
		return messages
	}

	msgGraph := buildMessageGraph(messages)
	keepCount := 0
	keptIndices := make(map[int]bool)

	for i, msg := range messages {
		if msg.Role == schema.System {
			keptIndices[i] = true
			keepCount++
		}
	}

	for i := len(messages) - 1; i >= 0; i-- {
		if keepCount >= maxItems {
			break
		}
		if messages[i].Role == schema.System || keptIndices[i] {
			continue
		}

		if messages[i].Role == schema.Assistant && len(messages[i].ToolCalls) > 0 {
			keptIndices[i] = true
			keepCount++
			for _, toolCall := range messages[i].ToolCalls {
				if toolResultIdx, ok := msgGraph.toolCallToResult[toolCall.ID]; ok {
					if !keptIndices[toolResultIdx] {
						keptIndices[toolResultIdx] = true
						keepCount++
					}
				}
			}
		} else if messages[i].Role == schema.Tool {
			if assistantIdx, ok := msgGraph.toolToAssistant[i]; ok {
				if !keptIndices[assistantIdx] && !keptIndices[i] {
					keptIndices[assistantIdx] = true
					keepCount++
					keptIndices[i] = true
					keepCount++
				}
			} else {
				if !keptIndices[i] {
					keptIndices[i] = true
					keepCount++
				}
			}
		} else {
			if !keptIndices[i] {
				keptIndices[i] = true
				keepCount++
			}
		}
	}

	result := make([]*schema.Message, 0, keepCount)
	for i := 0; i < len(messages); i++ {
		if keptIndices[i] {
			result = append(result, messages[i])
		}
	}
	return result
}

// RemoveOrphanedToolMessages 移除孤立的工具消息.
func RemoveOrphanedToolMessages(messages []*schema.Message) []*schema.Message {
	graph := buildMessageGraph(messages)
	hasValidAssistant := make(map[int]bool)
	for toolIdx, assistantIdx := range graph.toolToAssistant {
		if assistantIdx >= 0 && assistantIdx < len(messages) {
			hasValidAssistant[toolIdx] = true
		}
	}

	result := make([]*schema.Message, 0, len(messages))
	for i, msg := range messages {
		if msg.Role != schema.Tool || hasValidAssistant[i] {
			result = append(result, msg)
		}
	}
	return result
}

// EstimateMessagesTokens 估算消息列表的 token 数.
func EstimateMessagesTokens(msgs []*schema.Message) int {
	total := 0
	for _, msg := range msgs {
		total += EstimateMessageTokens(msg)
	}
	return total
}

// EstimateMessageTokens 估算单条消息的 token 数.
func EstimateMessageTokens(msg *schema.Message) int {
	if msg == nil {
		return 0
	}
	count := len(msg.Content)
	if len(msg.ToolCalls) > 0 {
		for _, tc := range msg.ToolCalls {
			count += len(tc.Function.Arguments)
		}
	}
	if msg.Role == schema.Tool {
		count += len(msg.ToolCallID)
		count += len(msg.ToolName)
	}
	if count == 0 {
		return 0
	}
	if strings.TrimSpace(msg.Content) == "" && len(msg.ToolCalls) == 0 {
		return 0
	}
	return count / 4
}

// TrimToTokenLimit 将消息裁剪到指定的 token 限制.
func TrimToTokenLimit(messages []*schema.Message, maxTokens int, estimator func([]*schema.Message) int) []*schema.Message {
	if estimator == nil {
		estimator = func(msgs []*schema.Message) int {
			total := 0
			for _, msg := range msgs {
				total += len(msg.Content) / 4
			}
			return total
		}
	}

	currentTokens := estimator(messages)
	if currentTokens <= maxTokens {
		return messages
	}

	for maxItems := len(messages) - 1; maxItems > 1; maxItems-- {
		trimmed := TrimMessagesWithConsistency(messages, maxItems)
		if estimator(trimmed) <= maxTokens {
			return RemoveOrphanedToolMessages(trimmed)
		}
	}

	result := make([]*schema.Message, 0)
	for _, msg := range messages {
		if msg.Role == schema.System {
			result = append(result, msg)
		}
	}
	return result
}
