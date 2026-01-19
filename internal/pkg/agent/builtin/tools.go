// Package builtin 提供内置工具和中间件.
package builtin

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cloudwego/eino/components/tool"
	toolutils "github.com/cloudwego/eino/components/tool/utils"
)

// ThinkingInput 思考工具输入.
type ThinkingInput struct {
	Thought string `json:"thought" jsonschema:"description=当前的思考内容，包括分析、推理和计划"`
}

// ThinkingOutput 思考工具输出.
type ThinkingOutput struct {
	Acknowledged bool   `json:"acknowledged"`
	Message      string `json:"message"`
}

// NewThinkingTool 创建思考工具.
func NewThinkingTool() tool.InvokableTool {
	t, _ := toolutils.InferTool(
		ToolThinking,
		`动态和反思性的问题解决思考工具。

用于：
- 分解复杂问题
- 规划解决步骤
- 反思和调整策略
- 组织思路

使用时机：
- 遇到复杂问题需要分析
- 需要制定多步骤计划
- 需要反思当前进度
- 需要调整策略方向`,
		func(ctx context.Context, input *ThinkingInput) (*ThinkingOutput, error) {
			return &ThinkingOutput{
				Acknowledged: true,
				Message:      "思考已记录",
			}, nil
		},
	)
	return t
}

// TodoWriteInput 计划工具输入.
type TodoWriteInput struct {
	Plan []PlanStep `json:"plan" jsonschema:"description=计划步骤列表"`
}

// PlanStep 计划步骤.
type PlanStep struct {
	Step   string `json:"step" jsonschema:"description=步骤描述"`
	Status string `json:"status" jsonschema:"enum=pending,in_progress,completed,description=步骤状态"`
}

// TodoWriteOutput 计划工具输出.
type TodoWriteOutput struct {
	Success bool       `json:"success"`
	Plan    []PlanStep `json:"plan"`
}

// NewTodoWriteTool 创建计划工具.
func NewTodoWriteTool() tool.InvokableTool {
	t, _ := toolutils.InferTool(
		ToolTodoWrite,
		`创建和更新结构化的研究计划。

用于：
- 创建任务计划
- 跟踪任务进度
- 更新任务状态

状态说明：
- pending: 待处理
- in_progress: 进行中
- completed: 已完成`,
		func(ctx context.Context, input *TodoWriteInput) (*TodoWriteOutput, error) {
			return &TodoWriteOutput{
				Success: true,
				Plan:    input.Plan,
			}, nil
		},
	)
	return t
}

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
	builtins := []tool.BaseTool{
		NewThinkingTool(),
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
