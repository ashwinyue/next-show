// Package builtin 提供内置工具和中间件.
package builtin

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/cloudwego/eino-ext/components/tool/duckduckgo/ddgsearch"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

// 工具描述模板
const webSearchToolDesc = `搜索互联网获取实时信息和新闻。

## 重要规则
- 优先使用知识库搜索（knowledge_search / grep_chunks）
- 仅当知识库无结果或需要实时信息时使用此工具

## 功能特性
- 实时网络搜索：搜索互联网获取最新信息
- 支持多种搜索引擎：DuckDuckGo（无需API密钥）

## 使用场景
- 知识库搜索无结果时
- 需要最新新闻、事件、更新
- 需要验证或补充知识库信息
- 搜索最新技术发展或趋势

## 参数
- query (必填): 搜索关键词

## 返回
- 搜索结果列表（最多 %d 条），包含标题、URL、摘要

## 示例
{"query": "eino agent framework latest updates"}
{"query": "Go 1.22 new features"}

## 提示
- 返回结果可能被截断，如需完整内容请使用 web_fetch
- 建议综合多个来源的信息`

// WebSearchConfig 网络搜索配置.
type WebSearchConfig struct {
	MaxResults int           `json:"max_results"` // 默认 10
	Region     string        `json:"region"`      // 默认 "wt-wt"
	Timeout    time.Duration `json:"timeout"`     // 默认 30s
}

// DefaultWebSearchConfig 返回默认配置.
func DefaultWebSearchConfig() *WebSearchConfig {
	return &WebSearchConfig{
		MaxResults: 10,
		Region:     "wt-wt",
		Timeout:    30 * time.Second,
	}
}

// WebSearchInput 网络搜索输入.
type WebSearchInput struct {
	Query string `json:"query" jsonschema:"description=搜索关键词"`
}

// WebSearchResult 单个搜索结果.
type WebSearchResult struct {
	Index       int    `json:"index"`
	Title       string `json:"title"`
	URL         string `json:"url"`
	Snippet     string `json:"snippet"`
	Content     string `json:"content,omitempty"`
	Source      string `json:"source,omitempty"`
	PublishedAt string `json:"published_at,omitempty"`
}

// WebSearchTool 网络搜索工具.
type WebSearchTool struct {
	config *WebSearchConfig
	ddg    *ddgsearch.DDGS
}

// NewWebSearchTool 创建网络搜索工具.
func NewWebSearchTool(config *WebSearchConfig) (*WebSearchTool, error) {
	if config == nil {
		config = DefaultWebSearchConfig()
	}

	ddg, err := ddgsearch.New(&ddgsearch.Config{
		Timeout:    config.Timeout,
		MaxRetries: 3,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create ddg client: %w", err)
	}

	return &WebSearchTool{
		config: config,
		ddg:    ddg,
	}, nil
}

// Info 返回工具信息.
func (t *WebSearchTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: ToolWebSearch,
		Desc: fmt.Sprintf(webSearchToolDesc, t.config.MaxResults),
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"query": {
				Type: "string",
				Desc: "搜索关键词",
			},
		}),
	}, nil
}

// InvokableRun 执行搜索.
func (t *WebSearchTool) InvokableRun(ctx context.Context, arguments string, opts ...tool.Option) (string, error) {
	// 解析输入
	var input WebSearchInput
	if err := json.Unmarshal([]byte(arguments), &input); err != nil {
		return t.formatError(fmt.Sprintf("参数解析失败: %v", err)), nil
	}

	if strings.TrimSpace(input.Query) == "" {
		return t.formatError("query 参数不能为空"), nil
	}

	// 执行搜索
	results, err := t.ddg.Search(ctx, &ddgsearch.SearchParams{
		Query:      input.Query,
		Region:     ddgsearch.Region(t.config.Region),
		MaxResults: t.config.MaxResults,
	})
	if err != nil {
		return t.formatError(fmt.Sprintf("搜索失败: %v", err)), nil
	}

	// 格式化输出
	return t.formatOutput(input.Query, results), nil
}

// formatOutput 格式化搜索结果（参考 weknora）.
func (t *WebSearchTool) formatOutput(query string, results *ddgsearch.SearchResponse) string {
	var sb strings.Builder

	sb.WriteString("=== Web Search Results ===\n")
	sb.WriteString(fmt.Sprintf("Query: %s\n", query))
	sb.WriteString(fmt.Sprintf("Found %d result(s)\n\n", len(results.Results)))

	if len(results.Results) == 0 {
		sb.WriteString("No results found.\n\n")
		sb.WriteString("=== Next Steps ===\n")
		sb.WriteString("- Try different search queries or keywords\n")
		sb.WriteString("- Check if question can be answered from knowledge base\n")
		return sb.String()
	}

	for i, r := range results.Results {
		sb.WriteString(fmt.Sprintf("Result #%d:\n", i+1))
		sb.WriteString(fmt.Sprintf("  Title: %s\n", r.Title))
		sb.WriteString(fmt.Sprintf("  URL: %s\n", r.URL))
		if r.Description != "" {
			// 截断过长的描述
			desc := r.Description
			if len(desc) > 500 {
				desc = desc[:500] + "..."
			}
			sb.WriteString(fmt.Sprintf("  Snippet: %s\n", desc))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("=== Next Steps ===\n")
	sb.WriteString("- ⚠️ Content may be truncated. Use web_fetch to get full page content.\n")
	sb.WriteString("- Extract URLs from results and use web_fetch for detailed information.\n")
	sb.WriteString("- Synthesize information from multiple sources for comprehensive answers.\n")

	return sb.String()
}

// formatError 格式化错误信息.
func (t *WebSearchTool) formatError(errMsg string) string {
	var sb strings.Builder
	sb.WriteString("=== Web Search Error ===\n")
	sb.WriteString(fmt.Sprintf("Error: %s\n\n", errMsg))
	sb.WriteString("=== Suggestions ===\n")
	sb.WriteString("- Check your search query\n")
	sb.WriteString("- Try again later\n")
	sb.WriteString("- Use knowledge_search as alternative\n")
	return sb.String()
}
