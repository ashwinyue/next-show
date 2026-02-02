package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"text/template"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"

	"github.com/ashwinyue/next-show/internal/model"
	"github.com/ashwinyue/next-show/internal/store"
)

// SkillBackend 定义 skill 存储后端接口
type SkillBackend interface {
	List(ctx context.Context) ([]model.Skill, error)
	Get(ctx context.Context, id string) (*model.Skill, error)
}

// StoreSkillBackend 基于数据库 store 的 skill 后端实现
type StoreSkillBackend struct {
	store store.Store
}

// NewStoreSkillBackend 创建基于 store 的 skill 后端
func NewStoreSkillBackend(s store.Store) SkillBackend {
	return &StoreSkillBackend{store: s}
}

// List 列出所有启用的 skills
func (b *StoreSkillBackend) List(ctx context.Context) ([]model.Skill, error) {
	skills, err := b.store.Skills().ListEnabled(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]model.Skill, 0, len(skills))
	for _, skill := range skills {
		result = append(result, *skill)
	}
	return result, nil
}

// Get 根据 ID 或名称获取 skill
func (b *StoreSkillBackend) Get(ctx context.Context, id string) (*model.Skill, error) {
	// 先尝试按 ID 获取
	skill, err := b.store.Skills().Get(ctx, id)
	if err == nil && skill != nil {
		return skill, nil
	}

	// 如果按 ID 获取失败，尝试按名称搜索
	skills, err := b.store.Skills().ListEnabled(ctx)
	if err != nil {
		return nil, err
	}

	for _, s := range skills {
		if s.Name == id {
			return s, nil
		}
	}

	return nil, nil
}

// SkillTool 实现 skill 加载工具
type SkillTool struct {
	backend    SkillBackend
	toolName   string
	useChinese bool
}

// NewSkillTool 创建 skill 工具
func NewSkillTool(backend SkillBackend) tool.InvokableTool {
	return &SkillTool{
		backend:    backend,
		toolName:   "skill",
		useChinese: true,
	}
}

// Info 返回工具信息
func (t *SkillTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	skills, err := t.backend.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list skills: %w", err)
	}

	desc, err := t.renderToolDescription(skills)
	if err != nil {
		return nil, fmt.Errorf("failed to render skill tool description: %w", err)
	}

	descBase := toolDescriptionBaseChinese
	paramDesc := "技能名称（无需其他参数）。例如：\"数据分析\" 或 \"代码审查\""

	return &schema.ToolInfo{
		Name: t.toolName,
		Desc: descBase + desc,
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"skill": {
				Type:     schema.String,
				Desc:     paramDesc,
				Required: true,
			},
		}),
	}, nil
}

// InvokableRun 执行工具调用
func (t *SkillTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	args := &struct {
		Skill string `json:"skill"`
	}{}
	if err := json.Unmarshal([]byte(argumentsInJSON), args); err != nil {
		return "", fmt.Errorf("failed to unmarshal arguments: %w", err)
	}

	skill, err := t.backend.Get(ctx, args.Skill)
	if err != nil {
		return "", fmt.Errorf("failed to get skill '%s': %w", args.Skill, err)
	}

	if skill == nil {
		return "", fmt.Errorf("skill not found: %s", args.Skill)
	}

	// 构建 skill 内容（类似 ADK 格式）
	result := fmt.Sprintf("正在启动技能：%s\n\n", skill.Name)
	result += fmt.Sprintf("描述: %s\n\n", skill.Description)

	if skill.SystemPrompt != "" {
		result += fmt.Sprintf("## 系统提示词\n%s\n\n", skill.SystemPrompt)
	}

	if skill.Instructions != "" {
		result += fmt.Sprintf("## 指令\n%s\n\n", skill.Instructions)
	}

	if len(skill.Examples) > 0 {
		result += "## 示例\n"
		for i, example := range skill.Examples {
			result += fmt.Sprintf("\n### 示例 %d\n", i+1)
			result += fmt.Sprintf("**输入:**\n%s\n\n", example.Input)
			result += fmt.Sprintf("**输出:**\n%s\n", example.Output)
		}
	}

	return result, nil
}

// renderToolDescription 渲染工具描述
func (t *SkillTool) renderToolDescription(skills []model.Skill) (string, error) {
	tplContent := toolDescriptionTemplateChinese
	tpl, err := template.New("skills").Parse(tplContent)
	if err != nil {
		return "", err
	}

	// 转换 skill 列表
	matters := make([]skillInfo, 0, len(skills))
	for _, skill := range skills {
		matters = append(matters, skillInfo{
			Name:        skill.Name,
			Description: skill.Description,
		})
	}

	var buf bytes.Buffer
	err = tpl.Execute(&buf, descriptionHelper{Matters: matters})
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

// GetSystemPrompt 返回技能系统提示词（用于添加到 Agent 的系统提示词中）
func GetSystemPrompt() string {
	return skillSystemPromptChinese
}

type skillInfo struct {
	Name        string
	Description string
}

type descriptionHelper struct {
	Matters []skillInfo
}

const (
	toolDescriptionBaseChinese = `在主对话中执行技能

<技能指令>
当用户要求你执行任务时，检查下方可用技能列表中是否有技能可以更有效地完成任务。技能提供专业能力和领域知识。

如何调用：
- 仅使用技能名称调用此工具（无需其他参数）
- 示例：
  - skill: "数据分析" - 调用数据分析技能
  - skill: "代码审查" - 调用代码审查技能

重要说明：
- 当技能相关时，你必须立即调用此工具作为第一个动作
- 切勿仅在文本回复中提及技能而不实际调用此工具
- 这是阻塞性要求：在生成任何关于任务的其他响应之前，先调用相关的技能工具
- 仅使用 <可用技能> 中列出的技能
- 不要调用已经运行中的技能
</技能指令>

`
	toolDescriptionTemplateChinese = `<可用技能>
{{- range .Matters }}
<技能>
<名称>
{{ .Name }}
</名称>
<描述>
{{ .Description }}
</描述>
</技能>
{{- end }}
</可用技能>
`

	skillSystemPromptChinese = `
# 技能系统

**如何使用技能（渐进式展示）：**

技能遵循**渐进式展示**模式 - 你可以在工具描述中看到技能的名称和描述，但只在需要时才阅读完整说明：

1. **识别技能适用场景**：检查用户的任务是否匹配某个技能的描述
2. **阅读技能的完整说明**：使用 'skill' 工具加载 skill
3. **遵循技能说明操作**：工具结果包含逐步工作流程、最佳实践和示例
4. **访问支持文件**：技能可能包含辅助脚本、配置或参考文档

**何时使用技能：**
- 用户请求匹配某个技能的领域（例如"分析数据" -> 数据分析技能）
- 你需要专业知识或结构化工作流程
- 某个技能为复杂任务提供了经过验证的模式

**示例工作流程：**

用户："帮我分析一下这份销售数据"

1. 检查可用技能 -> 发现 "数据分析" 技能
2. 调用 'skill' 工具读取完整的技能说明
3. 遵循技能的分析工作流程
4. 应用技能提供的最佳实践

记住：技能让你更加强大和稳定。如有疑问，请检查是否存在适用于该任务的技能！
`
)
