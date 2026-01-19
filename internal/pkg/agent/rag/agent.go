// Package rag 提供 RAG 图编排能力.
package rag

import (
	"context"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/schema"
)

// RAGAgentConfig RAG Agent 配置.
type RAGAgentConfig struct {
	Name             string
	Description      string
	Graph            *Graph
	KnowledgeBaseIDs []string
	TopK             int
}

// RAGAgent 将 RAG Graph 包装为 adk.Agent.
type RAGAgent struct {
	name             string
	description      string
	graph            *Graph
	knowledgeBaseIDs []string
	topK             int
}

// NewRAGAgent 创建 RAG Agent.
func NewRAGAgent(cfg *RAGAgentConfig) *RAGAgent {
	name := cfg.Name
	if name == "" {
		name = "rag_agent"
	}
	desc := cfg.Description
	if desc == "" {
		desc = "知识库检索与问答 Agent，能够从知识库中检索相关信息并生成准确的回答"
	}
	topK := cfg.TopK
	if topK <= 0 {
		topK = 5
	}
	return &RAGAgent{
		name:             name,
		description:      desc,
		graph:            cfg.Graph,
		knowledgeBaseIDs: cfg.KnowledgeBaseIDs,
		topK:             topK,
	}
}

// Name 返回 Agent 名称.
func (a *RAGAgent) Name(_ context.Context) string {
	return a.name
}

// Description 返回 Agent 描述.
func (a *RAGAgent) Description(_ context.Context) string {
	return a.description
}

// Run 执行 RAG Agent.
func (a *RAGAgent) Run(ctx context.Context, input *adk.AgentInput, options ...adk.AgentRunOption) *adk.AsyncIterator[*adk.AgentEvent] {
	iter, generator := adk.NewAsyncIteratorPair[*adk.AgentEvent]()

	go func() {
		defer generator.Close()

		// 提取用户查询
		query := extractQuery(input.Messages)
		if query == "" {
			generator.Send(&adk.AgentEvent{
				AgentName: a.name,
				Output: &adk.AgentOutput{
					MessageOutput: &adk.MessageVariant{
						Message: schema.AssistantMessage("无法识别您的问题，请重新提问。", nil),
						Role:    schema.Assistant,
					},
				},
			})
			return
		}

		// 转换历史消息
		var history []*schema.Message
		if len(input.Messages) > 1 {
			history = make([]*schema.Message, 0, len(input.Messages)-1)
			for _, msg := range input.Messages[:len(input.Messages)-1] {
				history = append(history, msg)
			}
		}

		// 执行 RAG Graph
		ragInput := &RAGInput{
			Query:            query,
			KnowledgeBaseIDs: a.knowledgeBaseIDs,
			TopK:             a.topK,
			History:          history,
		}

		output, err := a.graph.Run(ctx, ragInput)
		if err != nil {
			generator.Send(&adk.AgentEvent{
				AgentName: a.name,
				Err:       err,
			})
			return
		}

		// 构建回答消息
		answer := output.Answer
		if len(output.Sources) > 0 {
			answer += "\n\n---\n**参考来源：**\n"
			for i, src := range output.Sources {
				if src.DocumentTitle != "" {
					answer += "- [" + string(rune('1'+i)) + "] " + src.DocumentTitle + "\n"
				}
			}
		}

		// 发送结果
		generator.Send(&adk.AgentEvent{
			AgentName: a.name,
			Output: &adk.AgentOutput{
				MessageOutput: &adk.MessageVariant{
					Message: schema.AssistantMessage(answer, nil),
					Role:    schema.Assistant,
				},
			},
			Action: adk.NewExitAction(),
		})
	}()

	return iter
}

// extractQuery 从消息列表中提取用户查询.
func extractQuery(messages []adk.Message) string {
	// 从最后一条用户消息中提取查询
	for i := len(messages) - 1; i >= 0; i-- {
		if messages[i] != nil && messages[i].Role == schema.User {
			return messages[i].Content
		}
	}
	return ""
}

// Ensure RAGAgent implements adk.Agent
var _ adk.Agent = (*RAGAgent)(nil)
