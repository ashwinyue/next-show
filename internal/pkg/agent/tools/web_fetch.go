// Package tools 提供内置工具和中间件.
package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	browseruse "github.com/cloudwego/eino-ext/components/tool/browseruse"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

const webFetchTimeout = 120 * time.Second

const webFetchToolDesc = `抓取网页的完整内容（支持动态渲染）。

## 使用场景
- web_search 返回的 snippet 不够，需要抓取完整页面
- 需要对某个 URL 做深入阅读
- 需要抓取动态渲染的页面（JavaScript 生成内容）

## 参数
- items: 批量抓取任务，每项包含 url 与 prompt（prompt 可用于描述你希望从页面中提取的内容）

## 返回
- 每个 URL 的抓取结果（可能截断），并给出下一步建议

## 注意
- 返回结果可能仍会因长度被截断
- 本工具使用浏览器渲染，可以获取 JavaScript 生成的内容
- 批量处理时会依次处理每个 URL`

// WebFetchConfig 网页抓取配置.
type WebFetchConfig struct {
	Timeout          time.Duration `json:"timeout"`
	Headless         bool          `json:"headless"`
	ChromePath       string        `json:"chrome_path"`
	ExtractChatModel tool.BaseTool `json:"-"` // 可选：用于智能提取内容的模型
}

// DefaultWebFetchConfig 默认配置.
func DefaultWebFetchConfig() *WebFetchConfig {
	return &WebFetchConfig{
		Timeout:  webFetchTimeout,
		Headless: true,
	}
}

// WebFetchInput web_fetch 输入.
type WebFetchInput struct {
	Items []WebFetchItem `json:"items" jsonschema:"description=批量抓取任务，每项包含 url 与 prompt"`
}

// WebFetchItem 单个抓取任务.
type WebFetchItem struct {
	URL    string `json:"url" jsonschema:"description=待抓取的网页 URL"`
	Prompt string `json:"prompt" jsonschema:"description=希望从页面中提取/关注的内容"`
}

type webFetchItemResult struct {
	output string
	err    error
}

// WebFetchTool 基于 browseruse 的网页抓取工具.
type WebFetchTool struct {
	config *WebFetchConfig
}

// NewWebFetchTool 创建 web_fetch 工具.
func NewWebFetchTool(config *WebFetchConfig) *WebFetchTool {
	if config == nil {
		config = DefaultWebFetchConfig()
	}
	if config.Timeout == 0 {
		config.Timeout = webFetchTimeout
	}
	return &WebFetchTool{config: config}
}

// Info 返回工具信息.
func (t *WebFetchTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: ToolWebFetch,
		Desc: webFetchToolDesc,
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"items": {
				Type: schema.Array,
				Desc: "批量抓取任务列表",
				ElemInfo: &schema.ParameterInfo{
					Type: schema.Object,
					SubParams: map[string]*schema.ParameterInfo{
						"url": {
							Type:     schema.String,
							Desc:     "待抓取的网页 URL",
							Required: true,
						},
						"prompt": {
							Type: schema.String,
							Desc: "希望从页面中提取/关注的内容",
						},
					},
				},
			},
		}),
	}, nil
}

// InvokableRun 执行网页抓取.
func (t *WebFetchTool) InvokableRun(ctx context.Context, arguments string, opts ...tool.Option) (string, error) {
	var input WebFetchInput
	if err := json.Unmarshal([]byte(arguments), &input); err != nil {
		return t.formatError(fmt.Sprintf("参数解析失败: %v", err)), nil
	}
	if len(input.Items) == 0 {
		return t.formatError("missing required parameter: items"), nil
	}

	results := make([]*webFetchItemResult, len(input.Items))
	var wg sync.WaitGroup

	for idx := range input.Items {
		i := idx
		item := input.Items[i]
		wg.Add(1)
		go func(index int, it WebFetchItem) {
			defer wg.Done()
			result := t.fetchSingleURL(ctx, it)
			results[index] = result
		}(i, item)
	}

	wg.Wait()

	return t.buildOutput(results), nil
}

// fetchSingleURL 抓取单个 URL.
func (t *WebFetchTool) fetchSingleURL(ctx context.Context, item WebFetchItem) *webFetchItemResult {
	url := strings.TrimSpace(item.URL)
	prompt := strings.TrimSpace(item.Prompt)

	if url == "" {
		err := fmt.Errorf("url is required")
		return &webFetchItemResult{
			output: fmt.Sprintf("URL: %s\n错误: %v\n", item.URL, err),
			err:    err,
		}
	}

	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		err := fmt.Errorf("invalid URL format")
		return &webFetchItemResult{
			output: fmt.Sprintf("URL: %s\n错误: %v\n", item.URL, err),
			err:    err,
		}
	}

	// 创建 browseruse 工具配置
	browserConfig := &browseruse.Config{
		Headless: t.config.Headless,
	}
	if t.config.ChromePath != "" {
		browserConfig.ChromeInstancePath = t.config.ChromePath
	}

	// 创建 browseruse 工具
	browserTool, err := browseruse.NewBrowserUseTool(ctx, browserConfig)
	if err != nil {
		return &webFetchItemResult{
			output: fmt.Sprintf("URL: %s\n错误: 创建浏览器失败: %v\n", url, err),
			err:    fmt.Errorf("failed to create browser: %w", err),
		}
	}
	defer browserTool.Cleanup()

	// 导航到 URL
	navResult, err := browserTool.Execute(&browseruse.Param{
		Action: browseruse.ActionGoToURL,
		URL:    &url,
	})
	if err != nil {
		return &webFetchItemResult{
			output: fmt.Sprintf("URL: %s\n错误: 导航失败: %v\n", url, err),
			err:    fmt.Errorf("failed to navigate: %w", err),
		}
	}
	if navResult.Error != "" {
		return &webFetchItemResult{
			output: fmt.Sprintf("URL: %s\n错误: %s\n", url, navResult.Error),
			err:    fmt.Errorf("navigation error: %s", navResult.Error),
		}
	}

	// 提取内容
	content, err := t.extractContent(ctx, browserTool, prompt)
	if err != nil {
		return &webFetchItemResult{
			output: fmt.Sprintf("URL: %s\n错误: 提取内容失败: %v\n", url, err),
			err:    fmt.Errorf("failed to extract content: %w", err),
		}
	}

	output := buildWebFetchOutput(url, prompt, content, false)
	return &webFetchItemResult{
		output: output,
		err:    nil,
	}
}

// extractContent 提取页面内容.
func (t *WebFetchTool) extractContent(ctx context.Context, browserTool *browseruse.Tool, prompt string) (string, error) {
	// 如果有 prompt，使用智能提取
	if prompt != "" {
		goal := prompt
		result, err := browserTool.Execute(&browseruse.Param{
			Action: browseruse.ActionExtractContent,
			Goal:   &goal,
		})
		if err != nil {
			return "", err
		}
		if result.Error != "" {
			return "", fmt.Errorf("extract error: %s", result.Error)
		}
		return result.Output, nil
	}

	// 否则直接获取页面内容
	goal := "summarize the page content"
	result, err := browserTool.Execute(&browseruse.Param{
		Action: browseruse.ActionExtractContent,
		Goal:   &goal,
	})
	if err != nil {
		return "", err
	}
	if result.Error != "" {
		return "", fmt.Errorf("extract error: %s", result.Error)
	}
	return result.Output, nil
}

// buildOutput 构建输出.
func (t *WebFetchTool) buildOutput(results []*webFetchItemResult) string {
	var sb strings.Builder
	sb.WriteString("=== Browser Fetch Results ===\n\n")

	successCount := 0
	for i, r := range results {
		if r == nil {
			sb.WriteString(fmt.Sprintf("#%d: 无结果（内部错误）\n\n", i+1))
			continue
		}
		sb.WriteString(fmt.Sprintf("#%d:\n%s\n\n", i+1, r.output))
		if r.err == nil {
			successCount++
		}
	}

	sb.WriteString("=== Next Steps ===\n")
	if successCount > 0 {
		sb.WriteString(fmt.Sprintf("- ✅ 成功抓取 %d/%d 个页面。\n", successCount, len(results)))
		sb.WriteString("- 如需进一步阅读，请针对抓取结果进行总结/引用。\n")
		if successCount < len(results) {
			sb.WriteString("- ⚠️ 部分 URL 抓取失败，可尝试其它来源或稍后重试。\n")
		}
	} else {
		sb.WriteString("- ❌ 没有成功抓取到内容，请检查 URL 是否可访问。\n")
		sb.WriteString("- 可回退使用知识库检索或换其它搜索结果 URL。\n")
	}

	return sb.String()
}

func (t *WebFetchTool) formatError(errMsg string) string {
	var sb strings.Builder
	sb.WriteString("=== Browser Fetch Error ===\n")
	sb.WriteString(fmt.Sprintf("Error: %s\n\n", errMsg))
	sb.WriteString("=== Suggestions ===\n")
	sb.WriteString("- Check your URLs\n")
	sb.WriteString("- Try again later\n")
	sb.WriteString("- Use web_search to find alternative sources\n")
	return sb.String()
}

func buildWebFetchOutput(url, prompt, content string, truncated bool) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("URL: %s\n", url))
	if prompt != "" {
		sb.WriteString(fmt.Sprintf("Prompt: %s\n", prompt))
	}
	if truncated {
		sb.WriteString("Content Preview (truncated):\n")
	} else {
		sb.WriteString("Content Preview:\n")
	}
	sb.WriteString(content)
	sb.WriteString("\n")
	return sb.String()
}
