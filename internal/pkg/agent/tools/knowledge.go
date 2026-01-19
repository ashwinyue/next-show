// Package tools 提供内置工具和中间件.
package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"

	"github.com/ashwinyue/next-show/internal/model"
)

// KnowledgeService 知识库服务接口.
type KnowledgeService interface {
	// SemanticSearch 语义搜索.
	SemanticSearch(ctx context.Context, req *SemanticSearchRequest) (*SemanticSearchResult, error)
	// KeywordSearch 关键词搜索.
	KeywordSearch(ctx context.Context, req *KeywordSearchRequest) (*KeywordSearchResult, error)
	// HybridSearch 混合检索（向量 + BM25）.
	HybridSearch(ctx context.Context, req *HybridSearchRequest) (*HybridSearchResult, error)
	// ListChunks 列出文档分块.
	ListChunks(ctx context.Context, req *ListChunksRequest) (*ListChunksResult, error)
}

// SemanticSearchRequest 语义搜索请求.
type SemanticSearchRequest struct {
	Queries          []string `json:"queries"`
	KnowledgeBaseIDs []string `json:"knowledge_base_ids,omitempty"`
	TopK             int      `json:"top_k,omitempty"`
}

// SemanticSearchResult 语义搜索结果.
type SemanticSearchResult struct {
	Chunks     []*ChunkResult `json:"chunks"`
	TotalCount int            `json:"total_count"`
}

// KeywordSearchRequest 关键词搜索请求.
type KeywordSearchRequest struct {
	Keywords         []string `json:"keywords"`
	KnowledgeBaseIDs []string `json:"knowledge_base_ids,omitempty"`
	TopK             int      `json:"top_k,omitempty"`
}

// KeywordSearchResult 关键词搜索结果.
type KeywordSearchResult struct {
	Chunks     []*ChunkResult `json:"chunks"`
	TotalCount int            `json:"total_count"`
}

// HybridSearchRequest 混合检索请求.
type HybridSearchRequest struct {
	Query            string   `json:"query"`
	KnowledgeBaseIDs []string `json:"knowledge_base_ids,omitempty"`
	TopK             int      `json:"top_k,omitempty"`
	VectorWeight     float64  `json:"vector_weight,omitempty"` // 向量搜索权重，默认 0.7
	BM25Weight       float64  `json:"bm25_weight,omitempty"`   // BM25 搜索权重，默认 0.3
}

// HybridSearchResult 混合检索结果.
type HybridSearchResult struct {
	Chunks     []*ChunkResult `json:"chunks"`
	TotalCount int            `json:"total_count"`
}

// ListChunksRequest 列出分块请求.
type ListChunksRequest struct {
	DocumentID string `json:"document_id"`
	Limit      int    `json:"limit,omitempty"`
	Offset     int    `json:"offset,omitempty"`
}

// ListChunksResult 列出分块结果.
type ListChunksResult struct {
	Chunks     []*ChunkResult `json:"chunks"`
	TotalCount int            `json:"total_count"`
}

// ChunkResult 分块结果.
type ChunkResult struct {
	ID              string  `json:"id"`
	DocumentID      string  `json:"document_id"`
	DocumentTitle   string  `json:"document_title"`
	KnowledgeBaseID string  `json:"knowledge_base_id"`
	ChunkIndex      int     `json:"chunk_index"`
	Content         string  `json:"content"`
	Score           float64 `json:"score,omitempty"`
}

// ============== Knowledge Search Tool ==============

const knowledgeSearchToolDesc = `语义/向量搜索工具，根据含义、意图和概念相关性检索知识。

## 用途
用于高层次理解任务：
- 概念解释
- 主题概述
- 推理型信息需求
- 无法通过关键词匹配的查询

## 输入行为
"queries" 必须包含 1-5 个简短、格式良好的语义问题或概念陈述。

## 示例
- "RAG 的主要思想是什么?"
- "向量数据库如何工作?"
- "解释 Embedding 的用途"

## 参数
- queries (必填): 1-5 个语义问题或概念陈述
- knowledge_base_ids (可选): 限制搜索范围的知识库 ID`

// KnowledgeSearchInput 语义搜索工具输入.
type KnowledgeSearchInput struct {
	Queries          []string `json:"queries" jsonschema:"description=1-5 个语义问题或概念陈述"`
	KnowledgeBaseIDs []string `json:"knowledge_base_ids,omitempty" jsonschema:"description=限制搜索范围的知识库 ID"`
}

// KnowledgeSearchTool 语义搜索工具.
type KnowledgeSearchTool struct {
	service          KnowledgeService
	knowledgeBaseIDs []string // 默认知识库 ID
	topK             int
}

// KnowledgeSearchConfig 语义搜索工具配置.
type KnowledgeSearchConfig struct {
	Service          KnowledgeService
	KnowledgeBaseIDs []string
	TopK             int
}

// NewKnowledgeSearchTool 创建语义搜索工具.
func NewKnowledgeSearchTool(config *KnowledgeSearchConfig) *KnowledgeSearchTool {
	topK := 10
	if config != nil && config.TopK > 0 {
		topK = config.TopK
	}
	var service KnowledgeService
	var kbIDs []string
	if config != nil {
		service = config.Service
		kbIDs = config.KnowledgeBaseIDs
	}
	return &KnowledgeSearchTool{
		service:          service,
		knowledgeBaseIDs: kbIDs,
		topK:             topK,
	}
}

// Info 返回工具信息.
func (t *KnowledgeSearchTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: ToolKnowledgeSearch,
		Desc: knowledgeSearchToolDesc,
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"queries": {
				Type: schema.Array,
				Desc: "1-5 个语义问题或概念陈述",
				ElemInfo: &schema.ParameterInfo{
					Type: schema.String,
				},
				Required: true,
			},
			"knowledge_base_ids": {
				Type: schema.Array,
				Desc: "限制搜索范围的知识库 ID",
				ElemInfo: &schema.ParameterInfo{
					Type: schema.String,
				},
			},
		}),
	}, nil
}

// InvokableRun 执行语义搜索.
func (t *KnowledgeSearchTool) InvokableRun(ctx context.Context, arguments string, opts ...tool.Option) (string, error) {
	var input KnowledgeSearchInput
	if err := json.Unmarshal([]byte(arguments), &input); err != nil {
		return t.formatError(fmt.Sprintf("参数解析失败: %v", err)), nil
	}

	if len(input.Queries) == 0 {
		return t.formatError("queries 参数不能为空"), nil
	}

	if t.service == nil {
		return t.formatError("知识库服务未配置"), nil
	}

	kbIDs := input.KnowledgeBaseIDs
	if len(kbIDs) == 0 {
		kbIDs = t.knowledgeBaseIDs
	}

	result, err := t.service.SemanticSearch(ctx, &SemanticSearchRequest{
		Queries:          input.Queries,
		KnowledgeBaseIDs: kbIDs,
		TopK:             t.topK,
	})
	if err != nil {
		return t.formatError(fmt.Sprintf("搜索失败: %v", err)), nil
	}

	return t.formatOutput(input.Queries, result), nil
}

func (t *KnowledgeSearchTool) formatOutput(queries []string, result *SemanticSearchResult) string {
	var sb strings.Builder

	sb.WriteString("=== 语义搜索结果 ===\n")
	sb.WriteString(fmt.Sprintf("查询: %v\n", queries))
	sb.WriteString(fmt.Sprintf("找到 %d 个相关分块\n\n", result.TotalCount))

	for i, chunk := range result.Chunks {
		sb.WriteString(fmt.Sprintf("--- 结果 %d ---\n", i+1))
		sb.WriteString(fmt.Sprintf("文档: %s\n", chunk.DocumentTitle))
		sb.WriteString(fmt.Sprintf("文档ID: %s\n", chunk.DocumentID))
		sb.WriteString(fmt.Sprintf("分块索引: %d\n", chunk.ChunkIndex))
		if chunk.Score > 0 {
			sb.WriteString(fmt.Sprintf("相关度: %.2f\n", chunk.Score))
		}
		sb.WriteString(fmt.Sprintf("内容:\n%s\n\n", chunk.Content))
	}

	return sb.String()
}

func (t *KnowledgeSearchTool) formatError(errMsg string) string {
	return fmt.Sprintf("=== 语义搜索错误 ===\nError: %s\n", errMsg)
}

// ============== Grep Chunks Tool ==============

const grepChunksToolDesc = `关键词搜索工具，用于快速定位包含特定关键词的文档和分块。

## 用途
用于精确查找：
- 特定术语或实体名称
- 错误代码或日志信息
- 配置参数名称
- 代码片段

## 何时使用
- 知道要搜索的确切关键词
- 需要查找特定名称、代码或术语
- 语义搜索结果不够精确时

## 参数
- keywords (必填): 1-5 个要搜索的关键词
- knowledge_base_ids (可选): 限制搜索范围的知识库 ID`

// GrepChunksInput 关键词搜索工具输入.
type GrepChunksInput struct {
	Keywords         []string `json:"keywords" jsonschema:"description=1-5 个要搜索的关键词"`
	KnowledgeBaseIDs []string `json:"knowledge_base_ids,omitempty" jsonschema:"description=限制搜索范围的知识库 ID"`
}

// GrepChunksTool 关键词搜索工具.
type GrepChunksTool struct {
	service          KnowledgeService
	knowledgeBaseIDs []string
	topK             int
}

// GrepChunksConfig 关键词搜索工具配置.
type GrepChunksConfig struct {
	Service          KnowledgeService
	KnowledgeBaseIDs []string
	TopK             int
}

// NewGrepChunksTool 创建关键词搜索工具.
func NewGrepChunksTool(config *GrepChunksConfig) *GrepChunksTool {
	topK := 20
	if config != nil && config.TopK > 0 {
		topK = config.TopK
	}
	var service KnowledgeService
	var kbIDs []string
	if config != nil {
		service = config.Service
		kbIDs = config.KnowledgeBaseIDs
	}
	return &GrepChunksTool{
		service:          service,
		knowledgeBaseIDs: kbIDs,
		topK:             topK,
	}
}

// Info 返回工具信息.
func (t *GrepChunksTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: ToolGrepChunks,
		Desc: grepChunksToolDesc,
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"keywords": {
				Type: schema.Array,
				Desc: "1-5 个要搜索的关键词",
				ElemInfo: &schema.ParameterInfo{
					Type: schema.String,
				},
				Required: true,
			},
			"knowledge_base_ids": {
				Type: schema.Array,
				Desc: "限制搜索范围的知识库 ID",
				ElemInfo: &schema.ParameterInfo{
					Type: schema.String,
				},
			},
		}),
	}, nil
}

// InvokableRun 执行关键词搜索.
func (t *GrepChunksTool) InvokableRun(ctx context.Context, arguments string, opts ...tool.Option) (string, error) {
	var input GrepChunksInput
	if err := json.Unmarshal([]byte(arguments), &input); err != nil {
		return t.formatError(fmt.Sprintf("参数解析失败: %v", err)), nil
	}

	if len(input.Keywords) == 0 {
		return t.formatError("keywords 参数不能为空"), nil
	}

	if t.service == nil {
		return t.formatError("知识库服务未配置"), nil
	}

	kbIDs := input.KnowledgeBaseIDs
	if len(kbIDs) == 0 {
		kbIDs = t.knowledgeBaseIDs
	}

	result, err := t.service.KeywordSearch(ctx, &KeywordSearchRequest{
		Keywords:         input.Keywords,
		KnowledgeBaseIDs: kbIDs,
		TopK:             t.topK,
	})
	if err != nil {
		return t.formatError(fmt.Sprintf("搜索失败: %v", err)), nil
	}

	return t.formatOutput(input.Keywords, result), nil
}

func (t *GrepChunksTool) formatOutput(keywords []string, result *KeywordSearchResult) string {
	var sb strings.Builder

	sb.WriteString("=== 关键词搜索结果 ===\n")
	sb.WriteString(fmt.Sprintf("关键词: %v\n", keywords))
	sb.WriteString(fmt.Sprintf("找到 %d 个匹配分块\n\n", result.TotalCount))

	for i, chunk := range result.Chunks {
		sb.WriteString(fmt.Sprintf("--- 结果 %d ---\n", i+1))
		sb.WriteString(fmt.Sprintf("文档: %s\n", chunk.DocumentTitle))
		sb.WriteString(fmt.Sprintf("文档ID: %s\n", chunk.DocumentID))
		sb.WriteString(fmt.Sprintf("分块索引: %d\n", chunk.ChunkIndex))
		sb.WriteString(fmt.Sprintf("内容:\n%s\n\n", chunk.Content))
	}

	return sb.String()
}

func (t *GrepChunksTool) formatError(errMsg string) string {
	return fmt.Sprintf("=== 关键词搜索错误 ===\nError: %s\n", errMsg)
}

// ============== List Knowledge Chunks Tool ==============

const listKnowledgeChunksToolDesc = `获取文档完整分块内容的工具。

## 用途
在使用 grep_chunks 或 knowledge_search 获取文档 ID 后，使用此工具查看完整内容。

## 使用流程
1. grep_chunks(["关键词"]) → 获取 document_id
2. list_knowledge_chunks(document_id) → 查看完整内容

## 何时使用
- 需要查看特定文档的完整分块内容
- 需要了解分块周围的上下文
- 检查文档有多少分块

## 参数
- document_id (必填): 文档 ID
- limit (可选): 每页分块数 (默认 20, 最大 100)
- offset (可选): 起始位置 (默认 0)`

// ListKnowledgeChunksInput 列出分块工具输入.
type ListKnowledgeChunksInput struct {
	DocumentID string `json:"document_id" jsonschema:"description=文档 ID"`
	Limit      int    `json:"limit,omitempty" jsonschema:"description=每页分块数,default=20"`
	Offset     int    `json:"offset,omitempty" jsonschema:"description=起始位置,default=0"`
}

// ListKnowledgeChunksTool 列出分块工具.
type ListKnowledgeChunksTool struct {
	service KnowledgeService
}

// ListKnowledgeChunksConfig 列出分块工具配置.
type ListKnowledgeChunksConfig struct {
	Service KnowledgeService
}

// NewListKnowledgeChunksTool 创建列出分块工具.
func NewListKnowledgeChunksTool(config *ListKnowledgeChunksConfig) *ListKnowledgeChunksTool {
	var service KnowledgeService
	if config != nil {
		service = config.Service
	}
	return &ListKnowledgeChunksTool{
		service: service,
	}
}

// Info 返回工具信息.
func (t *ListKnowledgeChunksTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: ToolListKnowledgeChunks,
		Desc: listKnowledgeChunksToolDesc,
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"document_id": {
				Type:     schema.String,
				Desc:     "文档 ID",
				Required: true,
			},
			"limit": {
				Type: schema.Integer,
				Desc: "每页分块数 (默认 20, 最大 100)",
			},
			"offset": {
				Type: schema.Integer,
				Desc: "起始位置 (默认 0)",
			},
		}),
	}, nil
}

// InvokableRun 执行列出分块.
func (t *ListKnowledgeChunksTool) InvokableRun(ctx context.Context, arguments string, opts ...tool.Option) (string, error) {
	var input ListKnowledgeChunksInput
	if err := json.Unmarshal([]byte(arguments), &input); err != nil {
		return t.formatError(fmt.Sprintf("参数解析失败: %v", err)), nil
	}

	if strings.TrimSpace(input.DocumentID) == "" {
		return t.formatError("document_id 参数不能为空"), nil
	}

	if t.service == nil {
		return t.formatError("知识库服务未配置"), nil
	}

	limit := input.Limit
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	offset := input.Offset
	if offset < 0 {
		offset = 0
	}

	result, err := t.service.ListChunks(ctx, &ListChunksRequest{
		DocumentID: input.DocumentID,
		Limit:      limit,
		Offset:     offset,
	})
	if err != nil {
		return t.formatError(fmt.Sprintf("获取分块失败: %v", err)), nil
	}

	return t.formatOutput(input.DocumentID, limit, offset, result), nil
}

func (t *ListKnowledgeChunksTool) formatOutput(docID string, limit, offset int, result *ListChunksResult) string {
	var sb strings.Builder

	sb.WriteString("=== 文档分块列表 ===\n")
	sb.WriteString(fmt.Sprintf("文档ID: %s\n", docID))
	sb.WriteString(fmt.Sprintf("总分块数: %d\n", result.TotalCount))
	sb.WriteString(fmt.Sprintf("当前显示: %d-%d\n\n", offset+1, offset+len(result.Chunks)))

	for i, chunk := range result.Chunks {
		sb.WriteString(fmt.Sprintf("--- 分块 %d (索引 %d) ---\n", offset+i+1, chunk.ChunkIndex))
		sb.WriteString(fmt.Sprintf("分块ID: %s\n", chunk.ID))
		sb.WriteString(fmt.Sprintf("内容:\n%s\n\n", chunk.Content))
	}

	if offset+len(result.Chunks) < result.TotalCount {
		sb.WriteString(fmt.Sprintf("提示: 还有更多分块，使用 offset=%d 获取下一页\n", offset+limit))
	}

	return sb.String()
}

func (t *ListKnowledgeChunksTool) formatError(errMsg string) string {
	return fmt.Sprintf("=== 列出分块错误 ===\nError: %s\n", errMsg)
}

// Ensure interfaces are implemented
var (
	_ tool.InvokableTool = (*KnowledgeSearchTool)(nil)
	_ tool.InvokableTool = (*GrepChunksTool)(nil)
	_ tool.InvokableTool = (*ListKnowledgeChunksTool)(nil)
)

// Placeholder to avoid unused import warning
var _ = model.KnowledgeBase{}
