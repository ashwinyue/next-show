// Package http 提供 HTTP Handler 层.
package http

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/ashwinyue/next-show/internal/biz/provider"
	"github.com/ashwinyue/next-show/internal/model"
)

// ListProviders 列出所有 Provider.
func (h *Handler) ListProviders(c *gin.Context) {
	providers, err := h.biz.Providers().ListProviders(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"providers": providers})
}

// GetProvider 获取 Provider 详情.
func (h *Handler) GetProvider(c *gin.Context) {
	id := c.Param("id")
	p, err := h.biz.Providers().GetProvider(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, p)
}

// CreateProviderRequest 创建 Provider 请求.
type CreateProviderRequest struct {
	Name          string              `json:"name" binding:"required"`
	DisplayName   string              `json:"display_name" binding:"required"`
	ProviderType  string              `json:"provider_type" binding:"required"`
	ModelCategory model.ModelCategory `json:"model_category"`
	BaseURL       string              `json:"base_url"`
	APIKey        string              `json:"api_key"`
	APISecret     string              `json:"api_secret"`
	DefaultModel  string              `json:"default_model"`
	Config        model.JSONMap       `json:"config"`
}

// CreateProvider 创建 Provider.
func (h *Handler) CreateProvider(c *gin.Context) {
	var req CreateProviderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	p, err := h.biz.Providers().CreateProvider(c.Request.Context(), &provider.CreateProviderRequest{
		Name:          req.Name,
		DisplayName:   req.DisplayName,
		ProviderType:  req.ProviderType,
		ModelCategory: req.ModelCategory,
		BaseURL:       req.BaseURL,
		APIKey:        req.APIKey,
		APISecret:     req.APISecret,
		DefaultModel:  req.DefaultModel,
		Config:        req.Config,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, p)
}

// UpdateProviderRequest 更新 Provider 请求.
type UpdateProviderRequestHTTP struct {
	Name          *string              `json:"name"`
	DisplayName   *string              `json:"display_name"`
	ProviderType  *string              `json:"provider_type"`
	ModelCategory *model.ModelCategory `json:"model_category"`
	BaseURL       *string              `json:"base_url"`
	APIKey        *string              `json:"api_key"`
	APISecret     *string              `json:"api_secret"`
	DefaultModel  *string              `json:"default_model"`
	Config        model.JSONMap        `json:"config"`
	IsEnabled     *bool                `json:"is_enabled"`
}

// UpdateProvider 更新 Provider.
func (h *Handler) UpdateProvider(c *gin.Context) {
	id := c.Param("id")
	var req UpdateProviderRequestHTTP
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	p, err := h.biz.Providers().UpdateProvider(c.Request.Context(), id, &provider.UpdateProviderRequest{
		Name:          req.Name,
		DisplayName:   req.DisplayName,
		ProviderType:  req.ProviderType,
		ModelCategory: req.ModelCategory,
		BaseURL:       req.BaseURL,
		APIKey:        req.APIKey,
		APISecret:     req.APISecret,
		DefaultModel:  req.DefaultModel,
		Config:        req.Config,
		IsEnabled:     req.IsEnabled,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, p)
}

// DeleteProvider 删除 Provider.
func (h *Handler) DeleteProvider(c *gin.Context) {
	id := c.Param("id")
	if err := h.biz.Providers().DeleteProvider(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// ListChatProviders 列出对话模型 Provider.
func (h *Handler) ListChatProviders(c *gin.Context) {
	providers, err := h.biz.Providers().ListChatProviders(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"providers": providers})
}

// ListEmbeddingProviders 列出 Embedding 模型 Provider.
func (h *Handler) ListEmbeddingProviders(c *gin.Context) {
	providers, err := h.biz.Providers().ListEmbeddingProviders(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"providers": providers})
}

// ListRerankProviders 列出 Rerank 模型 Provider.
func (h *Handler) ListRerankProviders(c *gin.Context) {
	providers, err := h.biz.Providers().ListRerankProviders(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"providers": providers})
}
