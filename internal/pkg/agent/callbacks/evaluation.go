// Package callbacks 提供 Agent 相关的 Callback Handler.
package callbacks

import (
	"context"
	"time"

	"github.com/cloudwego/eino/callbacks"
	"github.com/cloudwego/eino/schema"
)

// EvaluationCallbackHandler 用于收集评估数据的 Callback Handler.
type EvaluationCallbackHandler struct {
	data *EvaluationData
}

// EvaluationData 收集的评估数据.
type EvaluationData struct {
	// 检索相关
	Query            string
	RetrievedDocIDs  []string
	RetrievalStartAt time.Time
	RetrievalLatency time.Duration
	RetrievalError   error

	// 生成相关
	GeneratedAnswer   string
	GenerationStartAt time.Time
	GenerationLatency time.Duration
	GenerationError   error
	TokenUsage        *schema.TokenUsage

	// 原始输出（用于调试）
	RetrievalRawOutput  any
	GenerationRawOutput any
}

// NewEvaluationCallbackHandler 创建评估 Callback Handler.
func NewEvaluationCallbackHandler() *EvaluationCallbackHandler {
	return &EvaluationCallbackHandler{
		data: &EvaluationData{},
	}
}

func (h *EvaluationCallbackHandler) OnStart(ctx context.Context, info *callbacks.RunInfo, input callbacks.CallbackInput) context.Context {
	// 根据节点名称记录不同的信息
	switch info.Name {
	case "retrieve", "retriever", "search":
		// 记录检索开始时间
		h.data.RetrievalStartAt = time.Now()
		h.data.RetrievalError = nil

		// 尝试从输入中提取查询
		if input != nil {
			h.data.Query = extractString(input)
		}

	case "generate", "chat", "llm", "answer":
		// 记录生成开始时间
		h.data.GenerationStartAt = time.Now()
		h.data.GenerationError = nil
	}

	return ctx
}

func (h *EvaluationCallbackHandler) OnEnd(ctx context.Context, info *callbacks.RunInfo, output callbacks.CallbackOutput) context.Context {
	switch info.Name {
	case "retrieve", "retriever", "search":
		// 计算检索延迟
		h.data.RetrievalLatency = time.Since(h.data.RetrievalStartAt)
		h.data.RetrievalRawOutput = output

		// 尝试从输出中提取文档 ID
		if output != nil {
			h.data.RetrievedDocIDs = extractDocIDsFromOutput(output)
		}

	case "generate", "chat", "llm", "answer":
		// 计算生成延迟
		h.data.GenerationLatency = time.Since(h.data.GenerationStartAt)
		h.data.GenerationRawOutput = output

		// 尝试从输出中提取生成的文本和 Token 使用情况
		if output != nil {
			h.data.GeneratedAnswer = extractString(output)
			h.data.TokenUsage = extractTokenUsage(output)
		}
	}

	return ctx
}

func (h *EvaluationCallbackHandler) OnError(ctx context.Context, info *callbacks.RunInfo, err error) context.Context {
	switch info.Name {
	case "retrieve", "retriever", "search":
		h.data.RetrievalError = err

	case "generate", "chat", "llm", "answer":
		h.data.GenerationError = err
	}

	return ctx
}

func (h *EvaluationCallbackHandler) OnStartWithStreamInput(ctx context.Context, info *callbacks.RunInfo,
	input *schema.StreamReader[callbacks.CallbackInput]) context.Context {
	return ctx
}

func (h *EvaluationCallbackHandler) OnEndWithStreamOutput(ctx context.Context, info *callbacks.RunInfo,
	output *schema.StreamReader[callbacks.CallbackOutput]) context.Context {
	return ctx
}

func (h *EvaluationCallbackHandler) Needed(ctx context.Context, info *callbacks.RunInfo, timing callbacks.CallbackTiming) bool {
	// 处理所有节点
	return timing == callbacks.TimingOnStart ||
		timing == callbacks.TimingOnEnd ||
		timing == callbacks.TimingOnError
}

// GetData 获取收集的评估数据.
func (h *EvaluationCallbackHandler) GetData() *EvaluationData {
	return h.data
}

// Reset 重置数据（用于复用 Handler）.
func (h *EvaluationCallbackHandler) Reset() {
	h.data = &EvaluationData{}
}

// 辅助函数：尝试从 any 中提取字符串
func extractString(v any) string {
	if v == nil {
		return ""
	}

	switch val := v.(type) {
	case string:
		return val
	case *string:
		if val != nil {
			return *val
		}
	case []byte:
		return string(val)
	default:
		// 尝试通过反射或其他方式提取
		return ""
	}

	return ""
}

// 辅助函数：尝试从输出中提取 Token 使用情况
func extractTokenUsage(v any) *schema.TokenUsage {
	if v == nil {
		return nil
	}

	// 尝试直接转换为 TokenUsage
	if tu, ok := v.(*schema.TokenUsage); ok {
		return tu
	}

	// 尝试从 map 中提取
	if m, ok := v.(map[string]any); ok {
		if promptTokens, ok1 := m["prompt_tokens"]; ok1 {
			if completionTokens, ok2 := m["completion_tokens"]; ok2 {
				if totalTokens, ok3 := m["total_tokens"]; ok3 {
					prompt := int64(0)
					completion := int64(0)
					total := int64(0)

					switch v := promptTokens.(type) {
					case int:
						prompt = int64(v)
					case int64:
						prompt = v
					case float64:
						prompt = int64(v)
					}

					switch v := completionTokens.(type) {
					case int:
						completion = int64(v)
					case int64:
						completion = v
					case float64:
						completion = int64(v)
					}

					switch v := totalTokens.(type) {
					case int:
						total = int64(v)
					case int64:
						total = v
					case float64:
						total = int64(v)
					}

					return &schema.TokenUsage{
						PromptTokens:     int(prompt),
						CompletionTokens: int(completion),
						TotalTokens:      int(total),
					}
				}
			}
		}
	}

	return nil
}

// 辅助函数：尝试从输出中提取文档 ID
func extractDocIDsFromOutput(output any) []string {
	if output == nil {
		return []string{}
	}

	var docIDs []string

	// 尝试从 []*schema.Document 提取
	if docs, ok := output.([]*schema.Document); ok {
		for _, doc := range docs {
			if doc.ID != "" {
				docIDs = append(docIDs, doc.ID)
			}
		}
		return docIDs
	}

	// 尝试从 []schema.Document 提取（非指针）
	if docs, ok := output.([]schema.Document); ok {
		for _, doc := range docs {
			if doc.ID != "" {
				docIDs = append(docIDs, doc.ID)
			}
		}
		return docIDs
	}

	// 尝试从 map 中提取（某些工具可能返回 map）
	if m, ok := output.(map[string]any); ok {
		// 尝试常见的键名
		for _, key := range []string{"docs", "documents", "chunks", "results"} {
			if val, exists := m[key]; exists {
				if ids := extractDocIDsFromOutput(val); len(ids) > 0 {
					return ids
				}
			}
		}
	}

	// 尝试从字符串切片提取（某些工具可能直接返回 ID 列表）
	if ids, ok := output.([]string); ok {
		return ids
	}

	// 尝试从 interface 切片提取
	if items, ok := output.([]any); ok {
		for _, item := range items {
			if id, ok := item.(string); ok {
				docIDs = append(docIDs, id)
			}
		}
	}

	return docIDs
}

// Ensure EvaluationCallbackHandler implements callbacks.Handler
var _ callbacks.Handler = (*EvaluationCallbackHandler)(nil)
