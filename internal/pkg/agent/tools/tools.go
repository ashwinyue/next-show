// Package tools 提供内置工具和中间件.
package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cloudwego/eino/components/tool"
	sequentialthinking "github.com/cloudwego/eino-ext/components/tool/sequentialthinking"
)

// ToolRegistry 工具注册表.
type ToolRegistry struct {
	tools map[string]tool.BaseTool
}

// NewToolRegistry 创建工具注册表.
func NewToolRegistry() *ToolRegistry {
	return &ToolRegistry{
		tools: make(map[string]tool.BaseTool),
	}
}

// Register 注册工具.
func (r *ToolRegistry) Register(t tool.BaseTool) error {
	info, err := t.Info(context.Background())
	if err != nil {
		return err
	}
	r.tools[info.Name] = t
	return nil
}

// Get 获取工具.
func (r *ToolRegistry) Get(name string) (tool.BaseTool, error) {
	t, ok := r.tools[name]
	if !ok {
		return nil, fmt.Errorf("tool not found: %s", name)
	}
	return t, nil
}

// List 列出所有工具.
func (r *ToolRegistry) List() []tool.BaseTool {
	tools := make([]tool.BaseTool, 0, len(r.tools))
	for _, t := range r.tools {
		tools = append(tools, t)
	}
	return tools
}

// GetByNames 根据名称列表获取工具.
func (r *ToolRegistry) GetByNames(names []string) []tool.BaseTool {
	tools := make([]tool.BaseTool, 0, len(names))
	for _, name := range names {
		if t, ok := r.tools[name]; ok {
			tools = append(tools, t)
		}
	}
	return tools
}

// RegisterBuiltinTools 注册所有内置工具.
func (r *ToolRegistry) RegisterBuiltinTools() error {
	// 使用 eino-ext 官方实现
	thinkingTool, err := sequentialthinking.NewTool()
	if err != nil {
		return fmt.Errorf("failed to create sequential thinking tool: %w", err)
	}

	builtins := []tool.BaseTool{
		thinkingTool,
		NewTodoWriteTool(),
	}

	for _, t := range builtins {
		if err := r.Register(t); err != nil {
			return err
		}
	}
	return nil
}

// RegisterWebSearchTool 注册网络搜索工具.
func (r *ToolRegistry) RegisterWebSearchTool(config *WebSearchConfig) error {
	t, err := NewWebSearchTool(config)
	if err != nil {
		return err
	}
	return r.Register(t)
}

// RegisterKnowledgeSearchTool 注册语义搜索工具.
func (r *ToolRegistry) RegisterKnowledgeSearchTool(config *KnowledgeSearchConfig) error {
	t := NewKnowledgeSearchTool(config)
	return r.Register(t)
}

// RegisterGrepChunksTool 注册关键词搜索工具.
func (r *ToolRegistry) RegisterGrepChunksTool(config *GrepChunksConfig) error {
	t := NewGrepChunksTool(config)
	return r.Register(t)
}

// RegisterListKnowledgeChunksTool 注册列出分块工具.
func (r *ToolRegistry) RegisterListKnowledgeChunksTool(config *ListKnowledgeChunksConfig) error {
	t := NewListKnowledgeChunksTool(config)
	return r.Register(t)
}

// DefaultRegistry 创建并初始化默认工具注册表.
func DefaultRegistry() (*ToolRegistry, error) {
	r := NewToolRegistry()
	if err := r.RegisterBuiltinTools(); err != nil {
		return nil, err
	}
	return r, nil
}

// ParseToolResult 解析工具结果为 JSON.
func ParseToolResult(content string, target interface{}) error {
	return json.Unmarshal([]byte(content), target)
}
