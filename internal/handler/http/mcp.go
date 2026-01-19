// Package http 提供 HTTP Handler 层.
package http

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/ashwinyue/next-show/internal/biz/mcp"
	"github.com/ashwinyue/next-show/internal/model"
)

// ListMCPServers 列出所有 MCP Server.
func (h *Handler) ListMCPServers(c *gin.Context) {
	servers, err := h.biz.MCP().ListServers(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"servers": servers})
}

// GetMCPServer 获取 MCP Server 详情.
func (h *Handler) GetMCPServer(c *gin.Context) {
	id := c.Param("id")
	server, err := h.biz.MCP().GetServer(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, server)
}

// CreateMCPServerRequest 创建 MCP Server 请求.
type CreateMCPServerRequest struct {
	Name           string              `json:"name" binding:"required"`
	DisplayName    string              `json:"display_name" binding:"required"`
	Description    string              `json:"description"`
	TransportType  model.TransportType `json:"transport_type"`
	Command        string              `json:"command"`
	Args           model.JSONSlice     `json:"args"`
	Env            model.JSONMap       `json:"env"`
	ServerURL      string              `json:"server_url"`
	CustomHeaders  model.JSONMap       `json:"custom_headers"`
	TimeoutSeconds int                 `json:"timeout_seconds"`
}

// CreateMCPServer 创建 MCP Server.
func (h *Handler) CreateMCPServer(c *gin.Context) {
	var req CreateMCPServerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	server, err := h.biz.MCP().CreateServer(c.Request.Context(), &mcp.CreateServerRequest{
		Name:           req.Name,
		DisplayName:    req.DisplayName,
		Description:    req.Description,
		TransportType:  req.TransportType,
		Command:        req.Command,
		Args:           req.Args,
		Env:            req.Env,
		ServerURL:      req.ServerURL,
		CustomHeaders:  req.CustomHeaders,
		TimeoutSeconds: req.TimeoutSeconds,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, server)
}

// UpdateMCPServerRequest 更新 MCP Server 请求.
type UpdateMCPServerRequest struct {
	Name           *string              `json:"name"`
	DisplayName    *string              `json:"display_name"`
	Description    *string              `json:"description"`
	TransportType  *model.TransportType `json:"transport_type"`
	Command        *string              `json:"command"`
	Args           model.JSONSlice      `json:"args"`
	Env            model.JSONMap        `json:"env"`
	ServerURL      *string              `json:"server_url"`
	CustomHeaders  model.JSONMap        `json:"custom_headers"`
	TimeoutSeconds *int                 `json:"timeout_seconds"`
	IsEnabled      *bool                `json:"is_enabled"`
}

// UpdateMCPServer 更新 MCP Server.
func (h *Handler) UpdateMCPServer(c *gin.Context) {
	id := c.Param("id")
	var req UpdateMCPServerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	server, err := h.biz.MCP().UpdateServer(c.Request.Context(), id, &mcp.UpdateServerRequest{
		Name:           req.Name,
		DisplayName:    req.DisplayName,
		Description:    req.Description,
		TransportType:  req.TransportType,
		Command:        req.Command,
		Args:           req.Args,
		Env:            req.Env,
		ServerURL:      req.ServerURL,
		CustomHeaders:  req.CustomHeaders,
		TimeoutSeconds: req.TimeoutSeconds,
		IsEnabled:      req.IsEnabled,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, server)
}

// DeleteMCPServer 删除 MCP Server.
func (h *Handler) DeleteMCPServer(c *gin.Context) {
	id := c.Param("id")
	if err := h.biz.MCP().DeleteServer(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// ListMCPTools 列出 MCP Server 的工具.
func (h *Handler) ListMCPTools(c *gin.Context) {
	serverID := c.Param("id")
	tools, err := h.biz.MCP().ListTools(c.Request.Context(), serverID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"tools": tools})
}

// GetMCPTool 获取 MCP Tool 详情.
func (h *Handler) GetMCPTool(c *gin.Context) {
	toolID := c.Param("tool_id")
	tool, err := h.biz.MCP().GetTool(c.Request.Context(), toolID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, tool)
}

// CreateMCPToolRequest 创建 MCP Tool 请求.
type CreateMCPToolRequest struct {
	Name           string        `json:"name" binding:"required"`
	DisplayName    string        `json:"display_name"`
	Description    string        `json:"description"`
	InputSchema    model.JSONMap `json:"input_schema"`
	ReturnDirectly bool          `json:"return_directly"`
}

// CreateMCPTool 创建 MCP Tool.
func (h *Handler) CreateMCPTool(c *gin.Context) {
	serverID := c.Param("id")
	var req CreateMCPToolRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tool, err := h.biz.MCP().CreateTool(c.Request.Context(), serverID, &mcp.CreateToolRequest{
		Name:           req.Name,
		DisplayName:    req.DisplayName,
		Description:    req.Description,
		InputSchema:    req.InputSchema,
		ReturnDirectly: req.ReturnDirectly,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, tool)
}

// UpdateMCPToolRequest 更新 MCP Tool 请求.
type UpdateMCPToolRequest struct {
	Name           *string       `json:"name"`
	DisplayName    *string       `json:"display_name"`
	Description    *string       `json:"description"`
	InputSchema    model.JSONMap `json:"input_schema"`
	ReturnDirectly *bool         `json:"return_directly"`
	IsEnabled      *bool         `json:"is_enabled"`
}

// UpdateMCPTool 更新 MCP Tool.
func (h *Handler) UpdateMCPTool(c *gin.Context) {
	toolID := c.Param("tool_id")
	var req UpdateMCPToolRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tool, err := h.biz.MCP().UpdateTool(c.Request.Context(), toolID, &mcp.UpdateToolRequest{
		Name:           req.Name,
		DisplayName:    req.DisplayName,
		Description:    req.Description,
		InputSchema:    req.InputSchema,
		ReturnDirectly: req.ReturnDirectly,
		IsEnabled:      req.IsEnabled,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, tool)
}

// DeleteMCPTool 删除 MCP Tool.
func (h *Handler) DeleteMCPTool(c *gin.Context) {
	toolID := c.Param("tool_id")
	if err := h.biz.MCP().DeleteTool(c.Request.Context(), toolID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}
