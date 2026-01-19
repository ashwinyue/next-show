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

const thinkingToolDesc = `动态和反思性的问题解决思考工具。

通过灵活的思考过程帮助分析问题，可以适应和演进。

每个思考可以在之前的见解基础上构建、质疑或修订。

## 使用时机

- 将复杂问题分解为步骤
- 规划和设计，允许修订
- 可能需要纠正方向的分析
- 问题的全部范围最初可能不清楚
- 需要多步骤解决方案的问题
- 需要在多个步骤中保持上下文的任务
- 需要过滤无关信息的情况

## 关键特性

- 可以随着进展调整 total_thoughts
- 可以质疑或修订之前的思考
- 即使看起来已经结束，也可以添加更多思考
- 可以表达不确定性并探索替代方法
- 不是每个思考都需要线性构建 - 可以分支或回溯
- 生成解决方案假设
- 基于思考链步骤验证假设
- 重复过程直到满意
- 提供正确答案

## 参数说明

- **thought**: 当前思考步骤，用自然语言描述
- **next_thought_needed**: 是否需要更多思考
- **thought_number**: 当前思考编号
- **total_thoughts**: 预计需要的思考总数
- **is_revision**: 是否修订之前的思考
- **revises_thought**: 正在修订哪个思考编号
- **branch_from_thought**: 分支点思考编号
- **branch_id**: 分支标识符
- **needs_more_thoughts**: 是否需要更多思考`

// SequentialThinkingInput 顺序思考工具输入.
type SequentialThinkingInput struct {
	Thought           string `json:"thought" jsonschema:"description=当前思考步骤，用自然语言描述"`
	NextThoughtNeeded bool   `json:"next_thought_needed" jsonschema:"description=是否需要更多思考"`
	ThoughtNumber     int    `json:"thought_number" jsonschema:"description=当前思考编号,minimum=1"`
	TotalThoughts     int    `json:"total_thoughts" jsonschema:"description=预计需要的思考总数,minimum=1"`
	IsRevision        bool   `json:"is_revision,omitempty" jsonschema:"description=是否修订之前的思考"`
	RevisesThought    *int   `json:"revises_thought,omitempty" jsonschema:"description=正在修订哪个思考编号"`
	BranchFromThought *int   `json:"branch_from_thought,omitempty" jsonschema:"description=分支点思考编号"`
	BranchID          string `json:"branch_id,omitempty" jsonschema:"description=分支标识符"`
	NeedsMoreThoughts bool   `json:"needs_more_thoughts,omitempty" jsonschema:"description=是否需要更多思考"`
}

// SequentialThinkingTool 顺序思考工具.
type SequentialThinkingTool struct {
	thoughtHistory []SequentialThinkingInput
	branches       map[string][]SequentialThinkingInput
}

// NewSequentialThinkingTool 创建顺序思考工具.
func NewSequentialThinkingTool() *SequentialThinkingTool {
	return &SequentialThinkingTool{
		thoughtHistory: make([]SequentialThinkingInput, 0),
		branches:       make(map[string][]SequentialThinkingInput),
	}
}

// Info 返回工具信息.
func (t *SequentialThinkingTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: ToolThinking,
		Desc: thinkingToolDesc,
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"thought": {
				Type:     schema.String,
				Desc:     "当前思考步骤，用自然语言描述",
				Required: true,
			},
			"next_thought_needed": {
				Type:     schema.Boolean,
				Desc:     "是否需要更多思考",
				Required: true,
			},
			"thought_number": {
				Type:     schema.Integer,
				Desc:     "当前思考编号",
				Required: true,
			},
			"total_thoughts": {
				Type:     schema.Integer,
				Desc:     "预计需要的思考总数",
				Required: true,
			},
			"is_revision": {
				Type: schema.Boolean,
				Desc: "是否修订之前的思考",
			},
			"revises_thought": {
				Type: schema.Integer,
				Desc: "正在修订哪个思考编号",
			},
			"branch_from_thought": {
				Type: schema.Integer,
				Desc: "分支点思考编号",
			},
			"branch_id": {
				Type: schema.String,
				Desc: "分支标识符",
			},
			"needs_more_thoughts": {
				Type: schema.Boolean,
				Desc: "是否需要更多思考",
			},
		}),
	}, nil
}

// InvokableRun 执行思考工具.
func (t *SequentialThinkingTool) InvokableRun(ctx context.Context, arguments string, opts ...tool.Option) (string, error) {
	var input SequentialThinkingInput
	if err := json.Unmarshal([]byte(arguments), &input); err != nil {
		return t.formatError(fmt.Sprintf("参数解析失败: %v", err)), nil
	}

	if err := t.validate(input); err != nil {
		return t.formatError(fmt.Sprintf("参数验证失败: %v", err)), nil
	}

	if input.ThoughtNumber > input.TotalThoughts {
		input.TotalThoughts = input.ThoughtNumber
	}

	t.thoughtHistory = append(t.thoughtHistory, input)

	if input.BranchFromThought != nil && input.BranchID != "" {
		if t.branches[input.BranchID] == nil {
			t.branches[input.BranchID] = make([]SequentialThinkingInput, 0)
		}
		t.branches[input.BranchID] = append(t.branches[input.BranchID], input)
	}

	incomplete := input.NextThoughtNeeded || input.NeedsMoreThoughts ||
		input.ThoughtNumber < input.TotalThoughts

	return t.formatOutput(input, incomplete), nil
}

func (t *SequentialThinkingTool) validate(input SequentialThinkingInput) error {
	if strings.TrimSpace(input.Thought) == "" {
		return fmt.Errorf("thought 不能为空")
	}
	if input.ThoughtNumber < 1 {
		return fmt.Errorf("thought_number 必须 >= 1")
	}
	if input.TotalThoughts < 1 {
		return fmt.Errorf("total_thoughts 必须 >= 1")
	}
	return nil
}

func (t *SequentialThinkingTool) formatOutput(input SequentialThinkingInput, incomplete bool) string {
	var sb strings.Builder

	sb.WriteString("=== Thinking Process ===\n")
	sb.WriteString(fmt.Sprintf("Thought %d/%d\n", input.ThoughtNumber, input.TotalThoughts))

	if input.IsRevision && input.RevisesThought != nil {
		sb.WriteString(fmt.Sprintf("(Revising thought #%d)\n", *input.RevisesThought))
	}
	if input.BranchFromThought != nil && input.BranchID != "" {
		sb.WriteString(fmt.Sprintf("(Branch '%s' from thought #%d)\n", input.BranchID, *input.BranchFromThought))
	}

	sb.WriteString("\n")
	sb.WriteString(input.Thought)
	sb.WriteString("\n\n")

	sb.WriteString("=== Status ===\n")
	if incomplete {
		sb.WriteString("思考进行中，请继续探索和调用工具\n")
	} else {
		sb.WriteString("思考过程已完成\n")
	}

	sb.WriteString(fmt.Sprintf("History: %d thoughts recorded\n", len(t.thoughtHistory)))
	if len(t.branches) > 0 {
		branchKeys := make([]string, 0, len(t.branches))
		for k := range t.branches {
			branchKeys = append(branchKeys, k)
		}
		sb.WriteString(fmt.Sprintf("Branches: %v\n", branchKeys))
	}

	return sb.String()
}

func (t *SequentialThinkingTool) formatError(errMsg string) string {
	var sb strings.Builder
	sb.WriteString("=== Thinking Error ===\n")
	sb.WriteString(fmt.Sprintf("Error: %s\n", errMsg))
	return sb.String()
}

// GetThoughtHistory 获取思考历史.
func (t *SequentialThinkingTool) GetThoughtHistory() []SequentialThinkingInput {
	return t.thoughtHistory
}

// GetBranches 获取分支.
func (t *SequentialThinkingTool) GetBranches() map[string][]SequentialThinkingInput {
	return t.branches
}

// Reset 重置思考状态.
func (t *SequentialThinkingTool) Reset() {
	t.thoughtHistory = make([]SequentialThinkingInput, 0)
	t.branches = make(map[string][]SequentialThinkingInput)
}
