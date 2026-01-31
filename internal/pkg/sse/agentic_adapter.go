// Package sse 提供从 Agentic 流式事件到 SSE 的适配器。
package sse

import (
	"context"
	"fmt"
	"io"

	"github.com/cloudwego/eino/callbacks"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
)

// AgenticAdapter 将 Agentic 流式事件转换为 SSE 事件。
type AgenticAdapter struct {
	writer Writer
}

// NewAgenticAdapter 创建 Agentic 适配器。
func NewAgenticAdapter(writer Writer) *AgenticAdapter {
	return &AgenticAdapter{writer: writer}
}

// NewCallback 创建 AgenticModel Callback Handler。
func (a *AgenticAdapter) NewCallback() callbacks.Handler {
	builder := callbacks.NewHandlerBuilder()
	builder.OnEndWithStreamOutputFn(a.handleStreamOutput)
	return builder.Build()
}

// handleStreamOutput 处理流式输出。
func (a *AgenticAdapter) handleStreamOutput(
	ctx context.Context,
	runInfo *callbacks.RunInfo,
	output *schema.StreamReader[callbacks.CallbackOutput],
) context.Context {

	go func() {
		defer output.Close()

		for {
			chunk, err := output.Recv()
			if err == io.EOF {
				// 发送完成事件
				if gw, ok := a.writer.(*GinWriter); ok {
					gw.SendComplete("", "")
				}
				break
			}
			if err != nil {
				// 发送错误事件
				if gw, ok := a.writer.(*GinWriter); ok {
					gw.SendError(err.Error())
				}
				break
			}

			// 转换为 AgenticCallbackOutput
			modelOutput := model.ConvAgenticCallbackOutput(chunk)
			if modelOutput.Message == nil {
				continue
			}

			// 转换每个 ContentBlock
			for _, block := range modelOutput.Message.ContentBlocks {
				a.convertBlock(block)
			}
		}
	}()

	return ctx
}

// convertBlock 转换 ContentBlock 为 SSE 事件。
func (a *AgenticAdapter) convertBlock(block *schema.ContentBlock) {
	if block == nil {
		return
	}

	switch block.Type {
	// ========== 推理过程 ==========
	case schema.ContentBlockTypeReasoning:
		if block.Reasoning != nil {
			a.writer.Send(Event{
				Type:    EventTypeThinking,
				Content: block.Reasoning.Text,
			})
		}

	// ========== 自定义工具调用 ==========
	case schema.ContentBlockTypeFunctionToolCall:
		if block.FunctionToolCall != nil {
			a.writer.Send(Event{
				Type: EventTypeToolCall,
				ToolCalls: []map[string]any{
					{
						"name":      block.FunctionToolCall.Name,
						"arguments": block.FunctionToolCall.Arguments,
						"id":        block.FunctionToolCall.CallID,
					},
				},
			})
		}

	// ========== 自定义工具结果 ==========
	case schema.ContentBlockTypeFunctionToolResult:
		if block.FunctionToolResult != nil {
			a.writer.Send(Event{
				Type:    EventTypeToolResult,
				Content: block.FunctionToolResult.Result,
			})
		}

	// ========== Server Tool 调用 ==========
	case schema.ContentBlockTypeServerToolCall:
		if block.ServerToolCall != nil {
			a.writer.Send(Event{
				Type: EventTypeToolCall,
				ToolCalls: []map[string]any{
					{
						"name":        block.ServerToolCall.Name,
						"server_tool": true,
						"id":          block.ServerToolCall.CallID,
					},
				},
			})
		}

	// ========== Server Tool 结果 ==========
	case schema.ContentBlockTypeServerToolResult:
		if block.ServerToolResult != nil {
			result := fmt.Sprintf("%v", block.ServerToolResult.Result)
			a.writer.Send(Event{
				Type:    EventTypeToolResult,
				Content: result,
			})
		}

	// ========== MCP Tool 调用 ==========
	case schema.ContentBlockTypeMCPToolCall:
		if block.MCPToolCall != nil {
			a.writer.Send(Event{
				Type: EventTypeToolCall,
				ToolCalls: []map[string]any{
					{
						"name":         block.MCPToolCall.Name,
						"mcp_tool":     true,
						"server_label": block.MCPToolCall.ServerLabel,
						"arguments":    block.MCPToolCall.Arguments,
						"id":           block.MCPToolCall.CallID,
					},
				},
			})
		}

	// ========== MCP Tool 结果 ==========
	case schema.ContentBlockTypeMCPToolResult:
		if block.MCPToolResult != nil {
			a.writer.Send(Event{
				Type:    EventTypeToolResult,
				Content: block.MCPToolResult.Result,
			})
		}

	// ========== 文本生成 ==========
	case schema.ContentBlockTypeAssistantGenText:
		if block.AssistantGenText != nil {
			a.writer.Send(Event{
				Type:    EventTypeAnswer,
				Content: block.AssistantGenText.Text,
				Done:    false,
			})
		}

	// ========== 图像生成 ==========
	case schema.ContentBlockTypeAssistantGenImage:
		if block.AssistantGenImage != nil {
			a.writer.Send(Event{
				Type: EventTypeAnswer,
				Data: map[string]any{
					"type": "image",
					"url":  block.AssistantGenImage.URL,
				},
			})
		}

	// ========== 音频生成 ==========
	case schema.ContentBlockTypeAssistantGenAudio:
		if block.AssistantGenAudio != nil {
			a.writer.Send(Event{
				Type: EventTypeAnswer,
				Data: map[string]any{
					"type": "audio",
					"url":  block.AssistantGenAudio.URL,
				},
			})
		}

	// ========== 视频生成 ==========
	case schema.ContentBlockTypeAssistantGenVideo:
		if block.AssistantGenVideo != nil {
			a.writer.Send(Event{
				Type: EventTypeAnswer,
				Data: map[string]any{
					"type": "video",
					"url":  block.AssistantGenVideo.URL,
				},
			})
		}
	}
}
