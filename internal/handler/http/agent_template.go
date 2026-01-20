// Package http 提供 HTTP Handler 层.
package http

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/ashwinyue/next-show/internal/biz/agent"
	"github.com/ashwinyue/next-show/internal/model"
)

// AgentTemplateResponse Agent 模板响应.
type AgentTemplateResponse struct {
	Code        string                 `json:"code"`
	Name        string                 `json:"name"`
	DisplayName string                 `json:"display_name"`
	Description string                 `json:"description"`
	AgentType   model.AgentType        `json:"agent_type"`
	AgentRole   model.AgentRole        `json:"agent_role"`
	Category    string                 `json:"category"` // orchestrator, specialist
	Config      map[string]interface{} `json:"config"`   // 默认配置
	Icon        string                 `json:"icon"`     // 图标（可选）
}

// ListAgentTemplates 列出所有 Agent 模板.
func (h *Handler) ListAgentTemplates(c *gin.Context) {
	templates := h.getAgentTemplates()
	c.JSON(http.StatusOK, gin.H{
		"templates": templates,
		"total":     len(templates),
	})
}

// CreateAgentFromTemplateRequest 从模板创建 Agent 请求.
type CreateAgentFromTemplateRequest struct {
	TemplateCode  string   `json:"template_code" binding:"required"` // 模板代码
	Name          string   `json:"name" binding:"required"`          // Agent 名称
	DisplayName   string   `json:"display_name" binding:"required"`  // 显示名称
	ProviderID    string   `json:"provider_id" binding:"required"`
	ModelName     string   `json:"model_name" binding:"required"`
	SystemPrompt  string   `json:"system_prompt"` // 覆盖默认提示词
	SubAgentIDs   []string `json:"sub_agent_ids"` // 子 Agent ID 列表
	MaxIterations int      `json:"max_iterations"`
}

// CreateAgentFromTemplate 从模板创建 Agent.
func (h *Handler) CreateAgentFromTemplate(c *gin.Context) {
	var req CreateAgentFromTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 获取模板
	template, err := h.getAgentTemplateByCode(req.TemplateCode)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 构建创建请求
	createReq := &agent.CreateAgentRequest{
		Name:          req.Name,
		DisplayName:   req.DisplayName,
		Description:   template.Description,
		ProviderID:    req.ProviderID,
		ModelName:     req.ModelName,
		SystemPrompt:  req.SystemPrompt,
		AgentType:     template.AgentType,
		AgentRole:     template.AgentRole,
		MaxIterations: req.MaxIterations,
		Config:        model.JSONMap(template.Config),
		SubAgentIDs:   req.SubAgentIDs,
	}

	// 如果用户没有提供自定义提示词，使用模板默认值
	if req.SystemPrompt == "" && template.Config["default_prompt"] != nil {
		if prompt, ok := template.Config["default_prompt"].(string); ok {
			createReq.SystemPrompt = prompt
		}
	}

	// 如果没有指定最大迭代次数，使用模板默认值
	if req.MaxIterations == 0 && template.Config["default_iterations"] != nil {
		if iterations, ok := template.Config["default_iterations"].(int); ok {
			createReq.MaxIterations = iterations
		}
	}

	agentModel, err := h.biz.AgentConfig().CreateAgent(c.Request.Context(), createReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, agentModel)
}

// getAgentTemplates 获取所有 Agent 模板.
func (h *Handler) getAgentTemplates() []AgentTemplateResponse {
	return []AgentTemplateResponse{
		// ===== 主控 Agent 模板 =====
		{
			Code:        "supervisor",
			Name:        "supervisor",
			DisplayName: "主控协调器",
			Description: "协调多个子 Agent 协作完成任务，根据任务特点分配给最合适的子 Agent",
			AgentType:   model.AgentTypeSupervisor,
			AgentRole:   model.AgentRoleOrchestrator,
			Category:    "orchestrator",
			Config: map[string]interface{}{
				"default_prompt":     "你是一个智能任务协调器。你需要分析用户的需求，并将任务分配给最合适的子 Agent。你可以调用以下子 Agent：\n{sub_agents}\n\n请根据任务的特点，选择最合适的 Agent 来处理。",
				"default_iterations": 20,
			},
			Icon: "🎯",
		},
		{
			Code:        "deep",
			Name:        "deep",
			DisplayName: "深度思考",
			Description: "通过深度思考和多步推理解决复杂问题，可以调用工具和子 Agent",
			AgentType:   model.AgentTypeDeep,
			AgentRole:   model.AgentRoleOrchestrator,
			Category:    "orchestrator",
			Config: map[string]interface{}{
				"default_prompt":     "你是一个深度思考助手。对于复杂问题，你会：\n1. 仔细分析问题\n2. 调用相关工具获取信息\n3. 进行多步推理\n4. 给出详细解答",
				"default_iterations": 50,
			},
			Icon: "🧠",
		},
		{
			Code:        "plan_execute",
			Name:        "plan_execute",
			DisplayName: "计划执行",
			Description: "先制定计划，然后执行计划，执行过程中可以根据情况调整计划",
			AgentType:   model.AgentTypePlanExecute,
			AgentRole:   model.AgentRoleOrchestrator,
			Category:    "orchestrator",
			Config: map[string]interface{}{
				"default_prompt":     "你是一个计划执行专家。你会：\n1. 理解目标\n2. 制定详细计划\n3. 逐步执行\n4. 根据执行情况调整计划",
				"default_iterations": 10,
			},
			Icon: "📋",
		},
		{
			Code:        "sequential",
			Name:        "sequential",
			DisplayName: "顺序执行",
			Description: "按照固定顺序依次执行多个子 Agent，适用于标准化流程",
			AgentType:   model.AgentTypeSequential,
			AgentRole:   model.AgentRoleOrchestrator,
			Category:    "orchestrator",
			Config: map[string]interface{}{
				"default_prompt":     "你是顺序执行流程的协调者。你会按照预定顺序依次调用各个子 Agent。",
				"default_iterations": 1,
			},
			Icon: "➡️",
		},
		{
			Code:        "loop",
			Name:        "loop",
			DisplayName: "循环执行",
			Description: "循环执行子 Agent 列表，直到达到最大迭代次数或任务完成",
			AgentType:   model.AgentTypeLoop,
			AgentRole:   model.AgentRoleOrchestrator,
			Category:    "orchestrator",
			Config: map[string]interface{}{
				"default_prompt":     "你是循环执行的协调者。你会反复执行子 Agent 列表，直到任务完成。",
				"default_iterations": 10,
			},
			Icon: "🔁",
		},

		// ===== 专家 Agent 模板 =====
		{
			Code:        "chat",
			Name:        "chat",
			DisplayName: "对话助手",
			Description: "基础对话 Agent，适合简单的问答和对话场景",
			AgentType:   model.AgentTypeChatModel,
			AgentRole:   model.AgentRoleSpecialist,
			Category:    "specialist",
			Config: map[string]interface{}{
				"default_prompt":     "你是一个友好的 AI 助手。请用简洁、准确的方式回答用户的问题。",
				"default_iterations": 1,
			},
			Icon: "💬",
		},
		{
			Code:        "rag",
			Name:        "rag",
			DisplayName: "知识检索",
			Description: "基于知识库的检索增强生成，适合需要查询文档的场景",
			AgentType:   model.AgentTypeRAG,
			AgentRole:   model.AgentRoleSpecialist,
			Category:    "specialist",
			Config: map[string]interface{}{
				"default_prompt":     "你是一个知识库助手。请根据检索到的知识库内容回答用户问题。如果知识库中没有相关信息，请明确告知。",
				"default_iterations": 1,
			},
			Icon: "📚",
		},
		{
			Code:        "data_analyst",
			Name:        "data_analyst",
			DisplayName: "数据分析",
			Description: "使用 DuckDB 进行数据分析，适合处理结构化数据",
			AgentType:   model.AgentTypeDataAnalyst,
			AgentRole:   model.AgentRoleSpecialist,
			Category:    "specialist",
			Config: map[string]interface{}{
				"default_prompt":     "你是一个数据分析专家。你可以使用 SQL 查询和分析数据。请用清晰的方式展示分析结果。",
				"default_iterations": 10,
			},
			Icon: "📊",
		},
		{
			Code:        "react",
			Name:        "react",
			DisplayName: "反应式",
			Description: "根据当前情况动态决策和行动，适合需要灵活响应的场景",
			AgentType:   model.AgentTypeReact,
			AgentRole:   model.AgentRoleSpecialist,
			Category:    "specialist",
			Config: map[string]interface{}{
				"default_prompt":     "你是一个反应式 Agent。你会观察当前情况，然后灵活地调用工具来完成任务。",
				"default_iterations": 20,
			},
			Icon: "⚡",
		},
	}
}

// getAgentTemplateByCode 根据代码获取模板.
func (h *Handler) getAgentTemplateByCode(code string) (AgentTemplateResponse, error) {
	templates := h.getAgentTemplates()
	for _, t := range templates {
		if t.Code == code {
			return t, nil
		}
	}
	return AgentTemplateResponse{}, fmt.Errorf("template not found: %s", code)
}
