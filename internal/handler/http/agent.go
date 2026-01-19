// Package http 提供 HTTP Handler 层.
package http

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/ashwinyue/next-show/internal/biz/agent"
	"github.com/ashwinyue/next-show/internal/model"
)

// ListAgents 列出所有 Agent.
func (h *Handler) ListAgents(c *gin.Context) {
	agents, err := h.biz.AgentConfig().ListAgents(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"agents": agents})
}

// GetAgent 获取 Agent 详情.
func (h *Handler) GetAgent(c *gin.Context) {
	id := c.Param("id")
	agentModel, err := h.biz.AgentConfig().GetAgent(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, agentModel)
}

// CreateAgentRequest 创建 Agent 请求.
type CreateAgentRequest struct {
	Name          string          `json:"name" binding:"required"`
	DisplayName   string          `json:"display_name" binding:"required"`
	Description   string          `json:"description"`
	ProviderID    string          `json:"provider_id" binding:"required"`
	ModelName     string          `json:"model_name" binding:"required"`
	SystemPrompt  string          `json:"system_prompt"`
	AgentType     model.AgentType `json:"agent_type" binding:"required"`
	AgentRole     model.AgentRole `json:"agent_role"`
	MaxIterations int             `json:"max_iterations"`
	Temperature   *float64        `json:"temperature"`
	MaxTokens     *int            `json:"max_tokens"`
	Config        model.JSONMap   `json:"config"`
	SubAgentIDs   []string        `json:"sub_agent_ids"`
}

// CreateAgent 创建 Agent.
func (h *Handler) CreateAgent(c *gin.Context) {
	var req CreateAgentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	agentModel, err := h.biz.AgentConfig().CreateAgent(c.Request.Context(), &agent.CreateAgentRequest{
		Name:          req.Name,
		DisplayName:   req.DisplayName,
		Description:   req.Description,
		ProviderID:    req.ProviderID,
		ModelName:     req.ModelName,
		SystemPrompt:  req.SystemPrompt,
		AgentType:     req.AgentType,
		AgentRole:     req.AgentRole,
		MaxIterations: req.MaxIterations,
		Temperature:   req.Temperature,
		MaxTokens:     req.MaxTokens,
		Config:        req.Config,
		SubAgentIDs:   req.SubAgentIDs,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, agentModel)
}

// UpdateAgentRequest 更新 Agent 请求.
type UpdateAgentRequest struct {
	Name          *string          `json:"name"`
	DisplayName   *string          `json:"display_name"`
	Description   *string          `json:"description"`
	ProviderID    *string          `json:"provider_id"`
	ModelName     *string          `json:"model_name"`
	SystemPrompt  *string          `json:"system_prompt"`
	AgentType     *model.AgentType `json:"agent_type"`
	AgentRole     *model.AgentRole `json:"agent_role"`
	MaxIterations *int             `json:"max_iterations"`
	Temperature   *float64         `json:"temperature"`
	MaxTokens     *int             `json:"max_tokens"`
	Config        model.JSONMap    `json:"config"`
	IsEnabled     *bool            `json:"is_enabled"`
	SubAgentIDs   []string         `json:"sub_agent_ids"`
}

// UpdateAgent 更新 Agent.
func (h *Handler) UpdateAgent(c *gin.Context) {
	id := c.Param("id")
	var req UpdateAgentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	agentModel, err := h.biz.AgentConfig().UpdateAgent(c.Request.Context(), id, &agent.UpdateAgentRequest{
		Name:          req.Name,
		DisplayName:   req.DisplayName,
		Description:   req.Description,
		ProviderID:    req.ProviderID,
		ModelName:     req.ModelName,
		SystemPrompt:  req.SystemPrompt,
		AgentType:     req.AgentType,
		AgentRole:     req.AgentRole,
		MaxIterations: req.MaxIterations,
		Temperature:   req.Temperature,
		MaxTokens:     req.MaxTokens,
		Config:        req.Config,
		IsEnabled:     req.IsEnabled,
		SubAgentIDs:   req.SubAgentIDs,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, agentModel)
}

// DeleteAgent 删除 Agent.
func (h *Handler) DeleteAgent(c *gin.Context) {
	id := c.Param("id")
	if err := h.biz.AgentConfig().DeleteAgent(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// ListBuiltinAgents 列出内置 Agent.
func (h *Handler) ListBuiltinAgents(c *gin.Context) {
	agents, err := h.biz.AgentConfig().ListBuiltinAgents(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"agents": agents})
}

// ListOrchestratorAgents 列出主控 Agent.
func (h *Handler) ListOrchestratorAgents(c *gin.Context) {
	agents, err := h.biz.AgentConfig().ListOrchestratorAgents(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"agents": agents})
}

// ListSpecialistAgents 列出子 Agent.
func (h *Handler) ListSpecialistAgents(c *gin.Context) {
	agents, err := h.biz.AgentConfig().ListSpecialistAgents(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"agents": agents})
}

// GetAgentRelations 获取 Agent 的子 Agent 关系.
func (h *Handler) GetAgentRelations(c *gin.Context) {
	id := c.Param("id")
	relations, err := h.biz.AgentConfig().GetAgentRelations(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"relations": relations})
}

// SetAgentRelationsRequest 设置子 Agent 关系请求.
type SetAgentRelationsRequest struct {
	SubAgentIDs []string `json:"sub_agent_ids" binding:"required"`
}

// SetAgentRelations 设置 Agent 的子 Agent 关系.
func (h *Handler) SetAgentRelations(c *gin.Context) {
	id := c.Param("id")
	var req SetAgentRelationsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.biz.AgentConfig().SetAgentRelations(c.Request.Context(), id, req.SubAgentIDs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "updated"})
}

// ListAgentTools 列出 Agent 的工具.
func (h *Handler) ListAgentTools(c *gin.Context) {
	id := c.Param("id")
	tools, err := h.biz.AgentConfig().ListAgentTools(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"tools": tools})
}

// AddAgentToolRequest 添加 Agent 工具请求.
type AddAgentToolRequest struct {
	ToolType         model.ToolType `json:"tool_type" binding:"required"`
	MCPToolID        *string        `json:"mcp_tool_id"`
	BuiltinToolName  string         `json:"builtin_tool_name"`
	CustomToolConfig model.JSONMap  `json:"custom_tool_config"`
	ReturnDirectly   bool           `json:"return_directly"`
	Priority         int            `json:"priority"`
}

// AddAgentTool 为 Agent 添加工具.
func (h *Handler) AddAgentTool(c *gin.Context) {
	id := c.Param("id")
	var req AddAgentToolRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tool, err := h.biz.AgentConfig().AddAgentTool(c.Request.Context(), id, &agent.AddAgentToolRequest{
		ToolType:         req.ToolType,
		MCPToolID:        req.MCPToolID,
		BuiltinToolName:  req.BuiltinToolName,
		CustomToolConfig: req.CustomToolConfig,
		ReturnDirectly:   req.ReturnDirectly,
		Priority:         req.Priority,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, tool)
}

// UpdateAgentToolRequest 更新 Agent 工具请求.
type UpdateAgentToolRequestHTTP struct {
	ReturnDirectly *bool `json:"return_directly"`
	IsEnabled      *bool `json:"is_enabled"`
	Priority       *int  `json:"priority"`
}

// UpdateAgentTool 更新 Agent 工具.
func (h *Handler) UpdateAgentTool(c *gin.Context) {
	toolID := c.Param("tool_id")
	var req UpdateAgentToolRequestHTTP
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tool, err := h.biz.AgentConfig().UpdateAgentTool(c.Request.Context(), toolID, &agent.UpdateAgentToolRequest{
		ReturnDirectly: req.ReturnDirectly,
		IsEnabled:      req.IsEnabled,
		Priority:       req.Priority,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, tool)
}

// RemoveAgentTool 移除 Agent 工具.
func (h *Handler) RemoveAgentTool(c *gin.Context) {
	toolID := c.Param("tool_id")
	if err := h.biz.AgentConfig().RemoveAgentTool(c.Request.Context(), toolID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// ListBuiltinTools 列出可用的内置工具.
func (h *Handler) ListBuiltinTools(c *gin.Context) {
	tools := h.biz.AgentConfig().ListBuiltinTools(c.Request.Context())
	c.JSON(http.StatusOK, gin.H{"tools": tools})
}
