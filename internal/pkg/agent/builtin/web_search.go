// Package builtin 提供内置工具和中间件.
package builtin

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/cloudwego/eino-ext/components/tool/duckduckgo"
	"github.com/cloudwego/eino-ext/components/tool/duckduckgo/ddgsearch"
	"github.com/cloudwego/eino/components/tool"
	toolutils "github.com/cloudwego/eino/components/tool/utils"
	"github.com/cloudwego/eino/schema"
)

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
	Title       string `json:"title"`
	Description string `json:"description"`
	URL         string `json:"url"`
}

// WebSearchOutput 网络搜索输出.
type WebSearchOutput struct {
	Query   string            `json:"query"`
	Results []WebSearchResult `json:"results"`
	Count   int               `json:"count"`
	Summary string            `json:"summary"`
}

// NewWebSearchTool 创建网络搜索工具（使用 DuckDuckGo）.
func NewWebSearchTool(config *WebSearchConfig) (tool.InvokableTool, error) {
	if config == nil {
		config = DefaultWebSearchConfig()
	}

	// 创建 DuckDuckGo 搜索工具
	ddgTool, err := duckduckgo.NewTool(context.Background(), &duckduckgo.Config{
		ToolName:   ToolWebSearch,
		ToolDesc:   "搜索互联网获取实时信息",
		MaxResults: config.MaxResults,
		Region:     ddgsearch.Region(config.Region),
		DDGConfig: &ddgsearch.Config{
			Timeout:    config.Timeout,
			MaxRetries: 3,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create duckduckgo tool: %w", err)
	}

	// 包装为自定义格式
	return &webSearchToolWrapper{
		ddgTool: ddgTool,
		config:  config,
	}, nil
}

type webSearchToolWrapper struct {
	ddgTool tool.InvokableTool
	config  *WebSearchConfig
}

func (w *webSearchToolWrapper) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: ToolWebSearch,
		Desc: `搜索互联网获取实时信息。

用于：
- 查询最新新闻和事件
- 获取实时信息
- 搜索知识库中没有的内容
- 验证或补充现有信息

使用时机：
- 知识库搜索无结果时
- 需要最新信息时
- 需要外部数据源时`,
	}, nil
}

func (w *webSearchToolWrapper) InvokableRun(ctx context.Context, arguments string, opts ...tool.Option) (string, error) {
	// 直接调用 DuckDuckGo 工具
	return w.ddgTool.InvokableRun(ctx, arguments, opts...)
}

// NewWebSearchToolSimple 创建简化版网络搜索工具.
func NewWebSearchToolSimple(config *WebSearchConfig) tool.InvokableTool {
	if config == nil {
		config = DefaultWebSearchConfig()
	}

	t, _ := toolutils.InferTool(
		ToolWebSearch,
		`搜索互联网获取实时信息。

用于：
- 查询最新新闻和事件
- 获取实时信息
- 搜索知识库中没有的内容

参数：
- query: 搜索关键词`,
		func(ctx context.Context, input *WebSearchInput) (*WebSearchOutput, error) {
			// 创建 DuckDuckGo 客户端
			ddg, err := ddgsearch.New(&ddgsearch.Config{
				Timeout:    config.Timeout,
				MaxRetries: 3,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to create ddg client: %w", err)
			}

			// 执行搜索
			results, err := ddg.Search(ctx, &ddgsearch.SearchParams{
				Query:      input.Query,
				Region:     ddgsearch.Region(config.Region),
				MaxResults: config.MaxResults,
			})
			if err != nil {
				return nil, fmt.Errorf("search failed: %w", err)
			}

			// 转换结果
			output := &WebSearchOutput{
				Query:   input.Query,
				Results: make([]WebSearchResult, 0, len(results.Results)),
				Count:   len(results.Results),
			}

			var summaryParts []string
			for i, r := range results.Results {
				output.Results = append(output.Results, WebSearchResult{
					Title:       r.Title,
					Description: r.Description,
					URL:         r.URL,
				})
				summaryParts = append(summaryParts, fmt.Sprintf("%d. %s: %s", i+1, r.Title, r.Description))
			}

			output.Summary = strings.Join(summaryParts, "\n")
			return output, nil
		},
	)
	return t
}
