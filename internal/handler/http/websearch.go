// Package http 提供 HTTP Handler 层.
package http

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/ashwinyue/next-show/internal/biz/websearch"
	"github.com/ashwinyue/next-show/internal/model"
)

// ListWebSearchConfigs 列出所有网络搜索配置.
func (h *Handler) ListWebSearchConfigs(c *gin.Context) {
	configs, err := h.biz.WebSearch().List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"configs": configs})
}

// GetWebSearchConfig 获取网络搜索配置详情.
func (h *Handler) GetWebSearchConfig(c *gin.Context) {
	id := c.Param("id")
	config, err := h.biz.WebSearch().Get(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, config)
}

// GetDefaultWebSearchConfig 获取默认网络搜索配置.
func (h *Handler) GetDefaultWebSearchConfig(c *gin.Context) {
	config, err := h.biz.WebSearch().GetDefault(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, config)
}

// CreateWebSearchConfigRequest 创建网络搜索配置请求.
type CreateWebSearchConfigRequest struct {
	Name           string                  `json:"name" binding:"required"`
	DisplayName    string                  `json:"display_name" binding:"required"`
	Provider       model.WebSearchProvider `json:"provider" binding:"required"`
	APIKey         string                  `json:"api_key"`
	BaseURL        string                  `json:"base_url"`
	MaxResults     int                     `json:"max_results"`
	SearchDepth    string                  `json:"search_depth"`
	IncludeDomains model.JSONSlice         `json:"include_domains"`
	ExcludeDomains model.JSONSlice         `json:"exclude_domains"`
	Config         model.JSONMap           `json:"config"`
	IsDefault      bool                    `json:"is_default"`
}

// CreateWebSearchConfig 创建网络搜索配置.
func (h *Handler) CreateWebSearchConfig(c *gin.Context) {
	var req CreateWebSearchConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	config, err := h.biz.WebSearch().Create(c.Request.Context(), &websearch.CreateRequest{
		Name:           req.Name,
		DisplayName:    req.DisplayName,
		Provider:       req.Provider,
		APIKey:         req.APIKey,
		BaseURL:        req.BaseURL,
		MaxResults:     req.MaxResults,
		SearchDepth:    req.SearchDepth,
		IncludeDomains: req.IncludeDomains,
		ExcludeDomains: req.ExcludeDomains,
		Config:         req.Config,
		IsDefault:      req.IsDefault,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, config)
}

// UpdateWebSearchConfigRequest 更新网络搜索配置请求.
type UpdateWebSearchConfigRequest struct {
	Name           *string                  `json:"name"`
	DisplayName    *string                  `json:"display_name"`
	Provider       *model.WebSearchProvider `json:"provider"`
	APIKey         *string                  `json:"api_key"`
	BaseURL        *string                  `json:"base_url"`
	MaxResults     *int                     `json:"max_results"`
	SearchDepth    *string                  `json:"search_depth"`
	IncludeDomains model.JSONSlice          `json:"include_domains"`
	ExcludeDomains model.JSONSlice          `json:"exclude_domains"`
	Config         model.JSONMap            `json:"config"`
	IsEnabled      *bool                    `json:"is_enabled"`
}

// UpdateWebSearchConfig 更新网络搜索配置.
func (h *Handler) UpdateWebSearchConfig(c *gin.Context) {
	id := c.Param("id")
	var req UpdateWebSearchConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	config, err := h.biz.WebSearch().Update(c.Request.Context(), id, &websearch.UpdateRequest{
		Name:           req.Name,
		DisplayName:    req.DisplayName,
		Provider:       req.Provider,
		APIKey:         req.APIKey,
		BaseURL:        req.BaseURL,
		MaxResults:     req.MaxResults,
		SearchDepth:    req.SearchDepth,
		IncludeDomains: req.IncludeDomains,
		ExcludeDomains: req.ExcludeDomains,
		Config:         req.Config,
		IsEnabled:      req.IsEnabled,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, config)
}

// DeleteWebSearchConfig 删除网络搜索配置.
func (h *Handler) DeleteWebSearchConfig(c *gin.Context) {
	id := c.Param("id")
	if err := h.biz.WebSearch().Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// SetDefaultWebSearchConfig 设置默认网络搜索配置.
func (h *Handler) SetDefaultWebSearchConfig(c *gin.Context) {
	id := c.Param("id")
	if err := h.biz.WebSearch().SetDefault(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "updated"})
}
