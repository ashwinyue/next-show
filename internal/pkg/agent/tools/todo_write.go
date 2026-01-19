// Package tools 提供内置工具和中间件.
package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

const todoWriteToolDesc = `创建和管理结构化任务列表的工具。

用于跟踪进度、组织复杂操作，并向用户展示任务完成情况。

## 使用时机

1. **复杂多步骤任务** - 任务需要3个或更多不同步骤
2. **非平凡复杂任务** - 需要仔细规划或多个操作
3. **用户明确请求** - 用户直接要求使用任务列表
4. **用户提供多个任务** - 用户提供编号或逗号分隔的任务列表
5. **接收新指令后** - 立即将用户需求捕获为待办事项
6. **开始任务时** - 开始工作前标记为 in_progress
7. **完成任务后** - 标记为 completed 并添加后续任务

## 不使用时机

1. 只有单个简单任务
2. 任务过于简单，跟踪没有价值
3. 纯粹的对话或信息性问题

## 任务状态

- **pending**: 待处理，尚未开始
- **in_progress**: 进行中（同时只能有一个）
- **completed**: 已完成

## 任务管理规则

1. 实时更新任务状态
2. 完成后立即标记为 completed
3. 同时只能有一个任务处于 in_progress
4. 完成当前任务后再开始新任务
5. 移除不再相关的任务

## 参数说明

- **task**: 任务或问题的描述
- **steps**: 计划步骤数组，每个步骤包含 id、description、status`

// TodoWriteInput 计划工具输入.
type TodoWriteInput struct {
	Task  string         `json:"task" jsonschema:"description=任务或问题的描述"`
	Steps []TodoPlanStep `json:"steps" jsonschema:"description=计划步骤数组"`
}

// TodoPlanStep 计划步骤.
type TodoPlanStep struct {
	ID          string `json:"id" jsonschema:"description=步骤唯一标识符"`
	Description string `json:"description" jsonschema:"description=步骤描述"`
	Status      string `json:"status" jsonschema:"enum=pending,in_progress,completed,description=步骤状态"`
}

// TodoWriteTool 任务计划工具.
type TodoWriteTool struct {
	currentTask  string
	currentSteps []TodoPlanStep
}

// NewTodoWriteTool 创建任务计划工具.
func NewTodoWriteTool() *TodoWriteTool {
	return &TodoWriteTool{
		currentSteps: make([]TodoPlanStep, 0),
	}
}

// Info 返回工具信息.
func (t *TodoWriteTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: ToolTodoWrite,
		Desc: todoWriteToolDesc,
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"task": {
				Type:     schema.String,
				Desc:     "任务或问题的描述",
				Required: true,
			},
			"steps": {
				Type: schema.Array,
				Desc: "计划步骤数组",
				ElemInfo: &schema.ParameterInfo{
					Type: schema.Object,
					SubParams: map[string]*schema.ParameterInfo{
						"id": {
							Type:     schema.String,
							Desc:     "步骤唯一标识符",
							Required: true,
						},
						"description": {
							Type:     schema.String,
							Desc:     "步骤描述",
							Required: true,
						},
						"status": {
							Type:     schema.String,
							Desc:     "步骤状态: pending, in_progress, completed",
							Enum:     []string{"pending", "in_progress", "completed"},
							Required: true,
						},
					},
				},
				Required: true,
			},
		}),
	}, nil
}

// InvokableRun 执行任务计划工具.
func (t *TodoWriteTool) InvokableRun(ctx context.Context, arguments string, opts ...tool.Option) (string, error) {
	var input TodoWriteInput
	if err := json.Unmarshal([]byte(arguments), &input); err != nil {
		return t.formatError(fmt.Sprintf("参数解析失败: %v", err)), nil
	}

	if strings.TrimSpace(input.Task) == "" {
		input.Task = "未提供任务描述"
	}

	t.currentTask = input.Task
	t.currentSteps = input.Steps

	return t.formatOutput(input.Task, input.Steps), nil
}

func (t *TodoWriteTool) formatOutput(task string, steps []TodoPlanStep) string {
	var sb strings.Builder

	sb.WriteString("=== 任务计划 ===\n")
	sb.WriteString(fmt.Sprintf("任务: %s\n\n", task))

	if len(steps) == 0 {
		sb.WriteString("注意：未提供具体步骤。\n\n")
		sb.WriteString("建议创建 3-7 个任务以系统化研究：\n")
		sb.WriteString("1. 搜索知识库获取相关信息\n")
		sb.WriteString("2. 检索关键文档内容\n")
		sb.WriteString("3. 使用网络搜索补充信息（如需要）\n")
		return sb.String()
	}

	pendingCount := 0
	inProgressCount := 0
	completedCount := 0

	for _, step := range steps {
		switch step.Status {
		case "pending":
			pendingCount++
		case "in_progress":
			inProgressCount++
		case "completed":
			completedCount++
		}
	}

	totalCount := len(steps)
	remainingCount := pendingCount + inProgressCount

	sb.WriteString("计划步骤:\n\n")

	for i, step := range steps {
		sb.WriteString(t.formatStep(i+1, step))
	}

	sb.WriteString("\n=== 任务进度 ===\n")
	sb.WriteString(fmt.Sprintf("总计: %d 个任务\n", totalCount))
	sb.WriteString(fmt.Sprintf("✅ 已完成: %d 个\n", completedCount))
	sb.WriteString(fmt.Sprintf("🔄 进行中: %d 个\n", inProgressCount))
	sb.WriteString(fmt.Sprintf("⏳ 待处理: %d 个\n", pendingCount))

	sb.WriteString("\n=== 下一步 ===\n")
	if remainingCount > 0 {
		sb.WriteString(fmt.Sprintf("还有 %d 个任务未完成\n", remainingCount))
		if inProgressCount > 0 {
			sb.WriteString("- 继续完成当前进行中的任务\n")
		}
		if pendingCount > 0 {
			sb.WriteString(fmt.Sprintf("- 开始处理 %d 个待处理任务\n", pendingCount))
		}
		sb.WriteString("- 完成每个任务后更新状态为 completed\n")
	} else {
		sb.WriteString("✅ 所有任务已完成！\n")
		sb.WriteString("- 综合所有任务的发现\n")
		sb.WriteString("- 生成完整的最终答案\n")
	}

	return sb.String()
}

func (t *TodoWriteTool) formatStep(index int, step TodoPlanStep) string {
	statusEmoji := map[string]string{
		"pending":     "⏳",
		"in_progress": "🔄",
		"completed":   "✅",
	}

	emoji, ok := statusEmoji[step.Status]
	if !ok {
		emoji = "⏳"
	}

	return fmt.Sprintf("  %d. %s [%s] %s\n", index, emoji, step.Status, step.Description)
}

func (t *TodoWriteTool) formatError(errMsg string) string {
	var sb strings.Builder
	sb.WriteString("=== Todo Write Error ===\n")
	sb.WriteString(fmt.Sprintf("Error: %s\n", errMsg))
	return sb.String()
}

// GetCurrentTask 获取当前任务.
func (t *TodoWriteTool) GetCurrentTask() string {
	return t.currentTask
}

// GetCurrentSteps 获取当前步骤.
func (t *TodoWriteTool) GetCurrentSteps() []TodoPlanStep {
	return t.currentSteps
}

// Reset 重置任务状态.
func (t *TodoWriteTool) Reset() {
	t.currentTask = ""
	t.currentSteps = make([]TodoPlanStep, 0)
}
