// Package http 提供 HTTP Handler 层.
package http

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/ashwinyue/next-show/internal/model"
)

// CreateSkillRequest 创建 Skill 请求.
type CreateSkillRequest struct {
	Name             string               `json:"name" binding:"required"`
	Description      string               `json:"description"`
	SystemPrompt     string               `json:"system_prompt"`
	Instructions     string               `json:"instructions"`
	Examples         []model.SkillExample `json:"examples"`
	ToolIDs          []string             `json:"tool_ids"`
	KnowledgeBaseIDs []string             `json:"knowledge_base_ids"`
	ModelProvider    string               `json:"model_provider"`
	ModelName        string               `json:"model_name"`
	Temperature      float64              `json:"temperature"`
	MaxIterations    int                  `json:"max_iterations"`
	Category         string               `json:"category"`
	Tags             []string             `json:"tags"`
}

// UpdateSkillRequest 更新 Skill 请求.
type UpdateSkillRequest struct {
	Name             *string              `json:"name"`
	Description      *string              `json:"description"`
	SystemPrompt     *string              `json:"system_prompt"`
	Instructions     *string              `json:"instructions"`
	Examples         []model.SkillExample `json:"examples"`
	ToolIDs          []string             `json:"tool_ids"`
	KnowledgeBaseIDs []string             `json:"knowledge_base_ids"`
	ModelProvider    *string              `json:"model_provider"`
	ModelName        *string              `json:"model_name"`
	Temperature      *float64             `json:"temperature"`
	MaxIterations    *int                 `json:"max_iterations"`
	Category         *string              `json:"category"`
	Tags             []string             `json:"tags"`
	IsEnabled        *bool                `json:"is_enabled"`
}

// CreateSkill 创建 Skill.
func (h *Handler) CreateSkill(c *gin.Context) {
	var req CreateSkillRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	skill := &model.Skill{
		Name:             req.Name,
		Description:      req.Description,
		SystemPrompt:     req.SystemPrompt,
		Instructions:     req.Instructions,
		Examples:         req.Examples,
		ToolIDs:          req.ToolIDs,
		KnowledgeBaseIDs: req.KnowledgeBaseIDs,
		ModelProvider:    req.ModelProvider,
		ModelName:        req.ModelName,
		Temperature:      req.Temperature,
		MaxIterations:    req.MaxIterations,
		Category:         req.Category,
		Tags:             req.Tags,
	}

	if skill.Category == "" {
		skill.Category = "general"
	}
	if skill.Temperature == 0 {
		skill.Temperature = 0.7
	}
	if skill.MaxIterations == 0 {
		skill.MaxIterations = 10
	}

	created, err := h.biz.Skills().Create(c.Request.Context(), skill)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, created)
}

// GetSkill 获取 Skill 详情.
func (h *Handler) GetSkill(c *gin.Context) {
	id := c.Param("id")

	skill, err := h.biz.Skills().Get(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "skill not found"})
		return
	}

	c.JSON(http.StatusOK, skill)
}

// UpdateSkill 更新 Skill.
func (h *Handler) UpdateSkill(c *gin.Context) {
	id := c.Param("id")

	var req UpdateSkillRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	skill, err := h.biz.Skills().Get(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "skill not found"})
		return
	}

	if req.Name != nil {
		skill.Name = *req.Name
	}
	if req.Description != nil {
		skill.Description = *req.Description
	}
	if req.SystemPrompt != nil {
		skill.SystemPrompt = *req.SystemPrompt
	}
	if req.Instructions != nil {
		skill.Instructions = *req.Instructions
	}
	if req.Examples != nil {
		skill.Examples = req.Examples
	}
	if req.ToolIDs != nil {
		skill.ToolIDs = req.ToolIDs
	}
	if req.KnowledgeBaseIDs != nil {
		skill.KnowledgeBaseIDs = req.KnowledgeBaseIDs
	}
	if req.ModelProvider != nil {
		skill.ModelProvider = *req.ModelProvider
	}
	if req.ModelName != nil {
		skill.ModelName = *req.ModelName
	}
	if req.Temperature != nil {
		skill.Temperature = *req.Temperature
	}
	if req.MaxIterations != nil {
		skill.MaxIterations = *req.MaxIterations
	}
	if req.Category != nil {
		skill.Category = *req.Category
	}
	if req.Tags != nil {
		skill.Tags = req.Tags
	}
	if req.IsEnabled != nil {
		skill.IsEnabled = *req.IsEnabled
	}

	updated, err := h.biz.Skills().Update(c.Request.Context(), skill)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, updated)
}

// DeleteSkill 删除 Skill.
func (h *Handler) DeleteSkill(c *gin.Context) {
	id := c.Param("id")

	if err := h.biz.Skills().Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// ListSkills 列出所有 Skills（分页）.
func (h *Handler) ListSkills(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	skills, total, err := h.biz.Skills().List(c.Request.Context(), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"items":     skills,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// ListSkillsByCategory 按分类列出 Skills.
func (h *Handler) ListSkillsByCategory(c *gin.Context) {
	category := c.Query("category")

	skills, err := h.biz.Skills().ListByCategory(c.Request.Context(), category)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, skills)
}

// SearchSkills 搜索 Skills.
func (h *Handler) SearchSkills(c *gin.Context) {
	keyword := c.Query("q")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	skills, total, err := h.biz.Skills().Search(c.Request.Context(), keyword, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"items":     skills,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}
