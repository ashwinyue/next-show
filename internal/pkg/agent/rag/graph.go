// Package rag 提供 RAG 图编排能力.
package rag

import (
	"context"
	"fmt"
	"strings"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

// RAGInput RAG 图输入.
type RAGInput struct {
	Query            string            // 用户查询
	SessionID        string            // 会话 ID
	KnowledgeBaseIDs []string          // 知识库 ID 列表
	TopK             int               // 检索数量
	History          []*schema.Message // 历史消息
	Metadata         map[string]any    // 额外元数据
}

// RAGOutput RAG 图输出.
type RAGOutput struct {
	Answer     string         // 生成的回答
	Sources    []*SourceChunk // 引用的来源
	RetrieveOK bool           // 检索是否成功
	Metadata   map[string]any // 额外元数据
}

// SourceChunk 来源分块.
type SourceChunk struct {
	ChunkID       string  // 分块 ID
	DocumentID    string  // 文档 ID
	Content       string  // 内容
	Score         float64 // 相关性得分
	DocumentTitle string  // 文档标题
}

// KnowledgeSearcher 知识检索接口.
type KnowledgeSearcher interface {
	SemanticSearch(ctx context.Context, query string, kbIDs []string, topK int) ([]*SourceChunk, error)
}

// GraphConfig RAG 图配置.
type GraphConfig struct {
	// ChatModel 用于生成回答的模型
	ChatModel model.BaseChatModel

	// Searcher 知识检索器
	Searcher KnowledgeSearcher

	// SystemPrompt 系统提示词模板
	SystemPrompt string

	// DefaultTopK 默认检索数量
	DefaultTopK int

	// MinConfidenceScore 最小置信度分数
	MinConfidenceScore float64
}

// Graph RAG 图.
type Graph struct {
	cfg      *GraphConfig
	compiled compose.Runnable[*RAGInput, *RAGOutput]
}

// NewGraph 创建 RAG 图.
func NewGraph(ctx context.Context, cfg *GraphConfig) (*Graph, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config is required")
	}
	if cfg.ChatModel == nil {
		return nil, fmt.Errorf("chat model is required")
	}
	if cfg.Searcher == nil {
		return nil, fmt.Errorf("searcher is required")
	}
	if cfg.DefaultTopK <= 0 {
		cfg.DefaultTopK = 5
	}
	if cfg.SystemPrompt == "" {
		cfg.SystemPrompt = defaultRAGSystemPrompt
	}

	g := &Graph{cfg: cfg}

	compiled, err := g.buildGraph(ctx)
	if err != nil {
		return nil, fmt.Errorf("build graph failed: %w", err)
	}

	g.compiled = compiled
	return g, nil
}

// Run 执行 RAG 图.
func (g *Graph) Run(ctx context.Context, input *RAGInput) (*RAGOutput, error) {
	return g.compiled.Invoke(ctx, input)
}

// buildGraph 构建 RAG 图.
func (g *Graph) buildGraph(ctx context.Context) (compose.Runnable[*RAGInput, *RAGOutput], error) {
	graph := compose.NewGraph[*RAGInput, *RAGOutput]()

	// 节点 1: 检索
	if err := graph.AddLambdaNode("retrieve", compose.InvokableLambda(g.retrieve)); err != nil {
		return nil, err
	}

	// 节点 2: 格式化上下文
	if err := graph.AddLambdaNode("format", compose.InvokableLambda(g.formatContext)); err != nil {
		return nil, err
	}

	// 节点 3: 生成回答
	if err := graph.AddLambdaNode("generate", compose.InvokableLambda(g.generate)); err != nil {
		return nil, err
	}

	// 节点 4: 聚合结果
	if err := graph.AddLambdaNode("aggregate", compose.InvokableLambda(g.aggregate)); err != nil {
		return nil, err
	}

	// 边: START -> retrieve
	if err := graph.AddEdge(compose.START, "retrieve"); err != nil {
		return nil, err
	}

	// 分支: retrieve -> (format | aggregate)
	if err := graph.AddBranch("retrieve", compose.NewGraphBranch(
		func(_ context.Context, result *retrieveResult) (string, error) {
			if len(result.Chunks) == 0 {
				return "aggregate", nil // 无检索结果，直接聚合
			}
			return "format", nil
		},
		map[string]bool{"format": true, "aggregate": true},
	)); err != nil {
		return nil, err
	}

	// 边: format -> generate
	if err := graph.AddEdge("format", "generate"); err != nil {
		return nil, err
	}

	// 边: generate -> aggregate
	if err := graph.AddEdge("generate", "aggregate"); err != nil {
		return nil, err
	}

	// 边: aggregate -> END
	if err := graph.AddEdge("aggregate", compose.END); err != nil {
		return nil, err
	}

	return graph.Compile(ctx)
}

// retrieveResult 检索结果.
type retrieveResult struct {
	Chunks []*SourceChunk
	Input  *RAGInput
}

// retrieve 检索知识.
func (g *Graph) retrieve(ctx context.Context, input *RAGInput) (*retrieveResult, error) {
	topK := input.TopK
	if topK <= 0 {
		topK = g.cfg.DefaultTopK
	}

	chunks, err := g.cfg.Searcher.SemanticSearch(ctx, input.Query, input.KnowledgeBaseIDs, topK)
	if err != nil {
		// 检索失败，返回空结果
		return &retrieveResult{
			Chunks: nil,
			Input:  input,
		}, nil
	}

	// 过滤低置信度结果
	var filtered []*SourceChunk
	for _, c := range chunks {
		if c.Score >= g.cfg.MinConfidenceScore {
			filtered = append(filtered, c)
		}
	}

	return &retrieveResult{
		Chunks: filtered,
		Input:  input,
	}, nil
}

// formatResult 格式化结果.
type formatResult struct {
	Context string
	Chunks  []*SourceChunk
	Input   *RAGInput
}

// formatContext 格式化上下文.
func (g *Graph) formatContext(_ context.Context, result *retrieveResult) (*formatResult, error) {
	var sb strings.Builder
	sb.WriteString("以下是从知识库中检索到的相关信息：\n\n")

	for i, chunk := range result.Chunks {
		sb.WriteString(fmt.Sprintf("【来源 %d】", i+1))
		if chunk.DocumentTitle != "" {
			sb.WriteString(fmt.Sprintf("（%s）", chunk.DocumentTitle))
		}
		sb.WriteString("\n")
		sb.WriteString(chunk.Content)
		sb.WriteString("\n\n")
	}

	return &formatResult{
		Context: sb.String(),
		Chunks:  result.Chunks,
		Input:   result.Input,
	}, nil
}

// generateResult 生成结果.
type generateResult struct {
	Answer string
	Chunks []*SourceChunk
	Input  *RAGInput
}

// generate 生成回答.
func (g *Graph) generate(ctx context.Context, result *formatResult) (*generateResult, error) {
	// 构建提示词
	systemPrompt := strings.ReplaceAll(g.cfg.SystemPrompt, "{{context}}", result.Context)

	messages := []*schema.Message{
		schema.SystemMessage(systemPrompt),
	}

	// 添加历史消息
	messages = append(messages, result.Input.History...)

	// 添加当前查询
	messages = append(messages, schema.UserMessage(result.Input.Query))

	// 调用模型生成
	resp, err := g.cfg.ChatModel.Generate(ctx, messages)
	if err != nil {
		return &generateResult{
			Answer: "抱歉，生成回答时发生错误，请稍后重试。",
			Chunks: result.Chunks,
			Input:  result.Input,
		}, nil
	}

	return &generateResult{
		Answer: resp.Content,
		Chunks: result.Chunks,
		Input:  result.Input,
	}, nil
}

// aggregate 聚合结果.
func (g *Graph) aggregate(_ context.Context, input any) (*RAGOutput, error) {
	output := &RAGOutput{
		Metadata: make(map[string]any),
	}

	switch v := input.(type) {
	case *retrieveResult:
		// 无检索结果
		output.Answer = "抱歉，在知识库中没有找到与您问题相关的信息。请尝试换一种方式提问，或者确认问题是否在知识库覆盖范围内。"
		output.RetrieveOK = false

	case *generateResult:
		output.Answer = v.Answer
		output.Sources = v.Chunks
		output.RetrieveOK = true
	}

	return output, nil
}

const defaultRAGSystemPrompt = `你是一个专业的问答助手。请根据以下提供的知识库内容回答用户的问题。

{{context}}

回答要求：
1. 仅根据上述知识库内容回答，不要编造信息
2. 如果知识库内容不足以回答问题，请明确告知用户
3. 回答要准确、简洁、有条理
4. 如果引用了特定来源，请在回答中标注来源编号`
