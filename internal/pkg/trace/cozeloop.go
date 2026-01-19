// Package trace 提供可观测性集成.
package trace

import (
	"context"
	"os"

	ccb "github.com/cloudwego/eino-ext/callbacks/cozeloop"
	"github.com/cloudwego/eino/callbacks"
	"github.com/coze-dev/cozeloop-go"
)

// CozeLoopConfig Coze-Loop 配置.
type CozeLoopConfig struct {
	WorkspaceID string
	APIToken    string
	Endpoint    string // 可选，默认使用官方端点
}

// CozeLoopTracer Coze-Loop 追踪器.
type CozeLoopTracer struct {
	client  cozeloop.Client
	handler callbacks.Handler
}

// NewCozeLoopTracer 创建 Coze-Loop 追踪器.
func NewCozeLoopTracer(cfg *CozeLoopConfig) (*CozeLoopTracer, error) {
	// 设置环境变量（如果配置中提供）
	if cfg != nil {
		if cfg.WorkspaceID != "" {
			os.Setenv("COZELOOP_WORKSPACE_ID", cfg.WorkspaceID)
		}
		if cfg.APIToken != "" {
			os.Setenv("COZELOOP_API_TOKEN", cfg.APIToken)
		}
		if cfg.Endpoint != "" {
			os.Setenv("COZELOOP_API_ENDPOINT", cfg.Endpoint)
		}
	}

	client, err := cozeloop.NewClient()
	if err != nil {
		return nil, err
	}

	handler := ccb.NewLoopHandler(client)

	return &CozeLoopTracer{
		client:  client,
		handler: handler,
	}, nil
}

// Register 注册全局回调处理器.
func (t *CozeLoopTracer) Register() {
	callbacks.AppendGlobalHandlers(t.handler)
}

// Close 关闭追踪器.
func (t *CozeLoopTracer) Close(ctx context.Context) {
	if t.client != nil {
		t.client.Close(ctx)
	}
}

// Handler 获取回调处理器.
func (t *CozeLoopTracer) Handler() callbacks.Handler {
	return t.handler
}
