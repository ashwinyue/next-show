// Package evaluation 提供评估专用的 Eino Callback Handler.
package evaluation

import (
	"context"
	"time"

	"github.com/cloudwego/eino/callbacks"
	"github.com/cloudwego/eino/components/reetrever"
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
	RetrievedDocs    []*RetrievedDoc
	RetrievalStartAt time.Time
	RetrievalLatency time.Duration
	RetrievalError   error

	// 生成相关
	GeneratedAnswer   string
	GenerationStartAt time.Time
	GenerationLatency time.Duration
	GenerationError   error
	TokenUsage        *schema.TokenUsage
}

// RetrievedDoc 检索到的文档.
type RetrievedDoc struct {
	ID       string
	Content  string
	Score    float64
	Metadata map[string]any
}

// NewEvaluationCallbackHandler 创建评估 Callback Handler.
func NewEvaluationCallbackHandler() *EvaluationCallbackHandler {
	return &EvaluationCallbackHandler{
		data: &EvaluationData{},
	}
}

func (h *EvaluationCallbackHandler) OnStart(ctx context.Context, info *callbacks.RunInfo, input callbacks.CallbackInput) context.Context {
	switch info.Component {
	case callbacks.ComponentTypeRetriever:
		// 记录检索开始时间
		h.data.RetrievalStartAt = time.Now()
		h.data.RetrievalError = nil

		// 提取查询
		if retrieverInput, ok := input.(*retriever.RetrieverInput); ok {
			h.data.Query = retrieverInput.Query
		}

	case callbacks.ComponentTypeChatModel:
		// 记录生成开始时间
		h.data.GenerationStartAt = time.Now()
		h.data.GenerationError = nil
	}

	return ctx
}

func (h *EvaluationCallbackHandler) OnEnd(ctx context.Context, info *callbacks.RunInfo, output callbacks.CallbackOutput) context.Context {
	switch info.Component {
	case callbacks.ComponentTypeRetriever:
		// 计算检索延迟
		h.data.RetrievalLatency = time.Since(h.data.RetrievalStartAt)

		// 提取检索结果
		if retrieverOutput, ok := output.(*[]*retriever.Document); ok {
			h.data.RetrievedDocs = make([]*RetrievedDoc, 0, len(*retrieverOutput))
			for _, doc := range *retrieverOutput {
				retrievedDoc := &RetrievedDoc{
					ID:       doc.ID,
					Content:  doc.Content,
					Score:    doc.Score,
					Metadata: make(map[string]any),
				}
				// 复制元数据
				for k, v := range doc.Metadata {
					retrievedDoc.Metadata[k] = v
				}
				h.data.RetrievedDocs = append(h.data.RetrievedDocs, retrievedDoc)
			}
		}

	case callbacks.ComponentTypeChatModel:
		// 计算生成延迟
		h.data.GenerationLatency = time.Since(h.data.GenerationStartAt)

		// 提取生成结果和 Token 使用
		if chatResult, ok := output.(*schema.ChatResult); ok {
			h.data.GeneratedAnswer = chatResult.Content
			h.data.TokenUsage = chatResult.TokenUsage
		}

		// 提取 Token 使用（从 streaming 结果）
		if streamOutput, ok := output.(*schema.StreamReader[*schema.ChatResult]); ok {
			// TODO: 处理流式输出
			_ = streamOutput
		}
	}

	return ctx
}

func (h *EvaluationCallbackHandler) OnError(ctx context.Context, info *callbacks.RunInfo, err error) context.Context {
	switch info.Component {
	case callbacks.ComponentTypeRetriever:
		h.data.RetrievalError = err
	case callbacks.ComponentTypeChatModel:
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
	// 只处理检索和聊天模型组件
	if info.Component != callbacks.ComponentTypeRetriever &&
		info.Component != callbacks.ComponentTypeChatModel {
		return false
	}

	// 处理所有时机
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

// Ensure EvaluationCallbackHandler implements callbacks.Handler
var _ callbacks.Handler = (*EvaluationCallbackHandler)(nil)
