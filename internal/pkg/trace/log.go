// Package trace 提供可观测性集成.
package trace

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/cloudwego/eino/callbacks"
	"github.com/cloudwego/eino/components"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

const (
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorPurple = "\033[35m"
	colorCyan   = "\033[36m"
	colorReset  = "\033[0m"
)

// LogTracer 本地日志追踪器.
type LogTracer struct {
	handler callbacks.Handler
	verbose bool
}

// NewLogTracer 创建本地日志追踪器.
func NewLogTracer(verbose bool) *LogTracer {
	t := &LogTracer{verbose: verbose}
	t.handler = t.buildHandler()
	return t
}

// Register 注册全局回调处理器.
func (t *LogTracer) Register() {
	callbacks.AppendGlobalHandlers(t.handler)
}

// Handler 获取回调处理器.
func (t *LogTracer) Handler() callbacks.Handler {
	return t.handler
}

func (t *LogTracer) buildHandler() callbacks.Handler {
	builder := callbacks.NewHandlerBuilder()
	builder.
		OnStartFn(func(ctx context.Context, info *callbacks.RunInfo, input callbacks.CallbackInput) context.Context {
			timestamp := time.Now().Format("15:04:05.000")
			switch info.Component {
			case components.ComponentOfTool:
				tci := tool.ConvCallbackInput(input)
				args := tci.ArgumentsInJSON
				if len(args) > 200 && !t.verbose {
					args = args[:200] + "..."
				}
				fmt.Printf("%s[%s]%s %s▶ TOOL%s [%s] args: %s\n",
					colorCyan, timestamp, colorReset,
					colorYellow, colorReset,
					info.Name, args)
			case components.ComponentOfChatModel, components.ComponentOfAgenticModel:
				fmt.Printf("%s[%s]%s %s▶ MODEL%s [%s] generating...\n",
					colorCyan, timestamp, colorReset,
					colorBlue, colorReset,
					info.Name)
			default:
				fmt.Printf("%s[%s]%s %s▶ %s%s [%s] started\n",
					colorCyan, timestamp, colorReset,
					colorPurple, info.Component, colorReset,
					info.Name)
			}
			return ctx
		}).
		OnEndFn(func(ctx context.Context, info *callbacks.RunInfo, output callbacks.CallbackOutput) context.Context {
			timestamp := time.Now().Format("15:04:05.000")
			switch info.Component {
			case components.ComponentOfTool:
				tco := tool.ConvCallbackOutput(output)
				result := tco.Response
				if len(result) > 200 && !t.verbose {
					result = result[:200] + "..."
				}
				fmt.Printf("%s[%s]%s %s◀ TOOL%s [%s] result: %s\n",
					colorCyan, timestamp, colorReset,
					colorYellow, colorReset,
					info.Name, result)
			case components.ComponentOfChatModel:
				cco := model.ConvCallbackOutput(output)
				msg := formatMessage(cco.Message, t.verbose)
				usage := ""
				if cco.TokenUsage != nil {
					usage = fmt.Sprintf(" (tokens: %d→%d)", cco.TokenUsage.PromptTokens, cco.TokenUsage.CompletionTokens)
				}
				fmt.Printf("%s[%s]%s %s◀ MODEL%s [%s] %s%s\n",
					colorCyan, timestamp, colorReset,
					colorBlue, colorReset,
					info.Name, msg, usage)
			case components.ComponentOfAgenticModel:
				aco := model.ConvAgenticCallbackOutput(output)
				if aco.Message != nil {
					msg := formatAgenticMessage(aco.Message, t.verbose)
					usage := ""
					if aco.TokenUsage != nil {
						usage = fmt.Sprintf(" (tokens: %d→%d)", aco.TokenUsage.PromptTokens, aco.TokenUsage.CompletionTokens)
					}
					fmt.Printf("%s[%s]%s %s◀ MODEL%s [%s] %s%s\n",
						colorCyan, timestamp, colorReset,
						colorBlue, colorReset,
						info.Name, msg, usage)
				}
			default:
				fmt.Printf("%s[%s]%s %s◀ %s%s [%s] completed\n",
					colorCyan, timestamp, colorReset,
					colorPurple, info.Component, colorReset,
					info.Name)
			}
			return ctx
		}).
		OnStartWithStreamInputFn(func(ctx context.Context, info *callbacks.RunInfo, input *schema.StreamReader[callbacks.CallbackInput]) context.Context {
			input.Close()
			timestamp := time.Now().Format("15:04:05.000")
			fmt.Printf("%s[%s]%s %s▶ %s%s [%s] stream started\n",
				colorCyan, timestamp, colorReset,
				colorGreen, info.Component, colorReset,
				info.Name)
			return ctx
		}).
		OnEndWithStreamOutputFn(func(ctx context.Context, info *callbacks.RunInfo, output *schema.StreamReader[callbacks.CallbackOutput]) context.Context {
			output.Close()
			timestamp := time.Now().Format("15:04:05.000")
			fmt.Printf("%s[%s]%s %s◀ %s%s [%s] stream completed\n",
				colorCyan, timestamp, colorReset,
				colorGreen, info.Component, colorReset,
				info.Name)
			return ctx
		}).
		OnErrorFn(func(ctx context.Context, info *callbacks.RunInfo, err error) context.Context {
			timestamp := time.Now().Format("15:04:05.000")
			fmt.Printf("%s[%s]%s %s✖ ERROR%s [%s:%s] %v\n",
				colorCyan, timestamp, colorReset,
				colorRed, colorReset,
				info.Component, info.Name, err)
			return ctx
		})
	return builder.Build()
}

func formatMessage(m *schema.Message, verbose bool) string {
	if m == nil {
		return "<nil>"
	}
	sb := strings.Builder{}
	sb.WriteString("[")
	sb.WriteString(string(m.Role))
	sb.WriteString("] ")

	if len(m.Content) > 0 {
		content := m.Content
		if len(content) > 100 && !verbose {
			content = content[:100] + "..."
		}
		sb.WriteString("\"")
		sb.WriteString(content)
		sb.WriteString("\"")
	}

	if len(m.ToolCalls) > 0 {
		sb.WriteString(" ToolCalls:[")
		for i, tc := range m.ToolCalls {
			if i > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(tc.Function.Name)
		}
		sb.WriteString("]")
	}
	return sb.String()
}

func formatAgenticMessage(m *schema.AgenticMessage, verbose bool) string {
	if m == nil {
		return "<nil>"
	}
	sb := strings.Builder{}
	sb.WriteString("[")
	sb.WriteString(string(m.Role))
	sb.WriteString("] ")

	// 遍历 ContentBlocks
	for i, block := range m.ContentBlocks {
		if i > 0 {
			sb.WriteString(", ")
		}
		switch block.Type {
		case schema.ContentBlockTypeReasoning:
			if block.Reasoning != nil {
				content := block.Reasoning.Text
				if len(content) > 100 && !verbose {
					content = content[:100] + "..."
				}
				sb.WriteString("Thinking:\"")
				sb.WriteString(content)
				sb.WriteString("\"")
			}
		case schema.ContentBlockTypeAssistantGenText:
			if block.AssistantGenText != nil {
				content := block.AssistantGenText.Text
				if len(content) > 100 && !verbose {
					content = content[:100] + "..."
				}
				sb.WriteString("\"")
				sb.WriteString(content)
				sb.WriteString("\"")
			}
		case schema.ContentBlockTypeFunctionToolCall:
			if block.FunctionToolCall != nil {
				sb.WriteString("ToolCall:")
				sb.WriteString(block.FunctionToolCall.Name)
			}
		case schema.ContentBlockTypeServerToolCall:
			if block.ServerToolCall != nil {
				sb.WriteString("ServerTool:")
				sb.WriteString(block.ServerToolCall.Name)
			}
		case schema.ContentBlockTypeMCPToolCall:
			if block.MCPToolCall != nil {
				sb.WriteString("MCPTool:")
				sb.WriteString(block.MCPToolCall.Name)
			}
		default:
			sb.WriteString(string(block.Type))
		}
	}
	return sb.String()
}
