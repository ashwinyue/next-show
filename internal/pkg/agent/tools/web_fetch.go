// Package tools 提供内置工具和中间件.
package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

const (
	webFetchTimeout  = 60 * time.Second
	webFetchMaxChars = 100000
)

const webFetchToolDesc = `抓取网页的完整内容（比 web_search 更完整）。

## 使用场景
- web_search 返回的 snippet 不够，需要抓取完整页面
- 需要对某个 URL 做深入阅读

## 参数
- items: 批量抓取任务，每项包含 url 与 prompt（prompt 可用于描述你希望从页面中提取的内容）

## 返回
- 每个 URL 的抓取结果（可能截断），并给出下一步建议

## 注意
- 返回结果可能仍会因长度被截断
- 如页面为强动态渲染站点，纯 HTTP 抓取可能获取不到完整内容（后续可再加 headless 渲染）`

// WebFetchConfig 网页抓取配置.
type WebFetchConfig struct {
	Timeout  time.Duration `json:"timeout"`
	MaxChars int           `json:"max_chars"`
}

// DefaultWebFetchConfig 默认配置.
func DefaultWebFetchConfig() *WebFetchConfig {
	return &WebFetchConfig{
		Timeout:  webFetchTimeout,
		MaxChars: webFetchMaxChars,
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
	data   map[string]any
	err    error
}

// WebFetchTool 网页抓取工具.
type WebFetchTool struct {
	config *WebFetchConfig
	client *http.Client
}

// NewWebFetchTool 创建 web_fetch 工具.
func NewWebFetchTool(config *WebFetchConfig) *WebFetchTool {
	if config == nil {
		config = DefaultWebFetchConfig()
	}
	if config.Timeout == 0 {
		config.Timeout = webFetchTimeout
	}
	if config.MaxChars == 0 {
		config.MaxChars = webFetchMaxChars
	}

	return &WebFetchTool{
		config: config,
		client: &http.Client{Timeout: config.Timeout},
	}
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
	wg.Add(len(input.Items))

	for idx := range input.Items {
		i := idx
		item := input.Items[i]
		go func(index int, it WebFetchItem) {
			defer wg.Done()

			url := strings.TrimSpace(it.URL)
			prompt := strings.TrimSpace(it.Prompt)
			if url == "" {
				err := fmt.Errorf("url is required")
				results[index] = &webFetchItemResult{
					output: fmt.Sprintf("URL: %s\n错误: %v\n", it.URL, err),
					data: map[string]any{
						"url":    it.URL,
						"prompt": prompt,
						"error":  err.Error(),
					},
					err: err,
				}
				return
			}
			if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
				err := fmt.Errorf("invalid URL format")
				results[index] = &webFetchItemResult{
					output: fmt.Sprintf("URL: %s\n错误: %v\n", it.URL, err),
					data: map[string]any{
						"url":    it.URL,
						"prompt": prompt,
						"error":  err.Error(),
					},
					err: err,
				}
				return
			}

			finalURL := normalizeGitHubURL(url)
			html, err := t.fetchHTML(ctx, finalURL)
			if err != nil {
				results[index] = &webFetchItemResult{
					output: fmt.Sprintf("URL: %s\n错误: %v\n", it.URL, err),
					data: map[string]any{
						"url":       it.URL,
						"final_url": finalURL,
						"prompt":    prompt,
						"error":     err.Error(),
					},
					err: err,
				}
				return
			}

			content := htmlToText(html)
			truncated := false
			if len(content) > t.config.MaxChars {
				content = content[:t.config.MaxChars] + "..."
				truncated = true
			}

			output := buildWebFetchOutput(it.URL, prompt, content, truncated)
			results[index] = &webFetchItemResult{
				output: output,
				data: map[string]any{
					"url":            it.URL,
					"final_url":      finalURL,
					"prompt":         prompt,
					"content":        content,
					"content_length": len(content),
					"truncated":      truncated,
				},
			}
		}(i, item)
	}

	wg.Wait()

	var sb strings.Builder
	sb.WriteString("=== Web Fetch Results ===\n\n")

	aggregated := make([]map[string]any, 0, len(results))
	success := true
	var firstErr error

	for i, r := range results {
		if r == nil {
			success = false
			if firstErr == nil {
				firstErr = fmt.Errorf("fetch item %d returned nil", i)
			}
			sb.WriteString(fmt.Sprintf("#%d: 无结果（内部错误）\n\n", i+1))
			continue
		}
		sb.WriteString(fmt.Sprintf("#%d:\n%s\n\n", i+1, r.output))
		if r.data != nil {
			aggregated = append(aggregated, r.data)
		}
		if r.err != nil {
			success = false
			if firstErr == nil {
				firstErr = r.err
			}
		}
	}

	sb.WriteString("=== Next Steps ===\n")
	if len(aggregated) > 0 {
		sb.WriteString("- ✅ 页面内容已抓取完成。\n")
		sb.WriteString("- 如需进一步阅读，请针对抓取结果进行总结/引用。\n")
		if !success {
			sb.WriteString("- ⚠️ 部分 URL 抓取失败，可尝试其它来源或稍后重试。\n")
		}
		sb.WriteString("- 若内容被截断或需要更完整细节，可调整抓取策略（后续可加 headless 渲染）。\n")
	} else {
		sb.WriteString("- ❌ 没有成功抓取到内容，请检查 URL 是否可访问。\n")
		sb.WriteString("- 可回退使用知识库检索或换其它搜索结果 URL。\n")
	}

	_ = firstErr
	return sb.String(), nil
}

func (t *WebFetchTool) formatError(errMsg string) string {
	var sb strings.Builder
	sb.WriteString("=== Web Fetch Error ===\n")
	sb.WriteString(fmt.Sprintf("Error: %s\n\n", errMsg))
	sb.WriteString("=== Suggestions ===\n")
	sb.WriteString("- Check your URLs\n")
	sb.WriteString("- Try again later\n")
	sb.WriteString("- Use web_search to find alternative sources\n")
	return sb.String()
}

func (t *WebFetchTool) fetchHTML(ctx context.Context, targetURL string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, targetURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	// 简单 UA，避免部分站点直接拒绝
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; next-show-webfetch/1.0)")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")

	resp, err := t.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("request failed with status %d %s", resp.StatusCode, resp.Status)
	}

	// 粗略限制读取大小，防止过大页面
	limited := io.LimitReader(resp.Body, int64(t.config.MaxChars*2))
	b, err := io.ReadAll(limited)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}
	return string(b), nil
}

func normalizeGitHubURL(source string) string {
	if strings.Contains(source, "github.com") && strings.Contains(source, "/blob/") {
		source = strings.Replace(source, "github.com", "raw.githubusercontent.com", 1)
		source = strings.Replace(source, "/blob/", "/", 1)
	}
	return source
}

func htmlToText(html string) string {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return basicTextExtraction(html)
	}

	// 移除脚本、样式、导航等无关内容
	doc.Find("script, style, nav, footer, header, aside, noscript").Remove()

	// 提取 body 文本
	text := strings.TrimSpace(doc.Find("body").Text())
	text = regexp.MustCompile(`\s+`).ReplaceAllString(text, " ")
	return text
}

func basicTextExtraction(html string) string {
	re := regexp.MustCompile(`<[^>]*>`)
	text := re.ReplaceAllString(html, " ")
	text = regexp.MustCompile(`\s+`).ReplaceAllString(text, " ")
	return strings.TrimSpace(text)
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
