// Package http 提供 HTTP Handler 层.
package http

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/ashwinyue/next-show/internal/biz/tenant"
)

// === 租户管理 ===

// ListTenants 列出租户.
func (h *Handler) ListTenants(c *gin.Context) {
	tenants, err := h.biz.Tenants().List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"tenants": tenants})
}

// GetTenant 获取租户详情.
func (h *Handler) GetTenant(c *gin.Context) {
	id := c.Param("id")
	t, err := h.biz.Tenants().Get(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, t)
}

// CreateTenant 创建租户.
func (h *Handler) CreateTenant(c *gin.Context) {
	var req tenant.CreateTenantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	t, err := h.biz.Tenants().Create(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, t)
}

// UpdateTenant 更新租户.
func (h *Handler) UpdateTenant(c *gin.Context) {
	id := c.Param("id")
	var req tenant.UpdateTenantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	t, err := h.biz.Tenants().Update(c.Request.Context(), id, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, t)
}

// DeleteTenant 删除租户.
func (h *Handler) DeleteTenant(c *gin.Context) {
	id := c.Param("id")
	if err := h.biz.Tenants().Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// === API Key 管理 ===

// ListAPIKeys 列出租户的 API Keys.
func (h *Handler) ListAPIKeys(c *gin.Context) {
	tenantID := c.Param("tenant_id")
	keys, err := h.biz.Tenants().ListAPIKeys(c.Request.Context(), tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"api_keys": keys})
}

// GetAPIKey 获取 API Key 详情.
func (h *Handler) GetAPIKey(c *gin.Context) {
	id := c.Param("key_id")
	key, err := h.biz.Tenants().GetAPIKey(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, key)
}

// CreateAPIKey 创建 API Key.
func (h *Handler) CreateAPIKey(c *gin.Context) {
	tenantID := c.Param("tenant_id")
	var req tenant.CreateAPIKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.TenantID = tenantID

	keyWithSecret, err := h.biz.Tenants().CreateAPIKey(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, keyWithSecret)
}

// RevokeAPIKey 撤销 API Key.
func (h *Handler) RevokeAPIKey(c *gin.Context) {
	id := c.Param("key_id")
	if err := h.biz.Tenants().RevokeAPIKey(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "revoked"})
}

// DeleteAPIKey 删除 API Key.
func (h *Handler) DeleteAPIKey(c *gin.Context) {
	id := c.Param("key_id")
	if err := h.biz.Tenants().DeleteAPIKey(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}
