// Package http 提供 HTTP Handler 层.
package http

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/ashwinyue/next-show/internal/biz/settings"
)

// GetSystemInfo 获取系统信息.
func (h *Handler) GetSystemInfo(c *gin.Context) {
	info := h.biz.Settings().GetSystemInfo(c.Request.Context())
	c.JSON(http.StatusOK, info)
}

// ListSettings 列出所有设置.
func (h *Handler) ListSettings(c *gin.Context) {
	category := c.Query("category")

	var (
		settingsList []*struct{}
		err          error
	)

	if category != "" {
		result, e := h.biz.Settings().ListByCategory(c.Request.Context(), category)
		if e != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": e.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"settings": result})
		return
	}

	result, err := h.biz.Settings().List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	_ = settingsList
	c.JSON(http.StatusOK, gin.H{"settings": result})
}

// GetSetting 获取单个设置.
func (h *Handler) GetSetting(c *gin.Context) {
	key := c.Param("key")
	setting, err := h.biz.Settings().Get(c.Request.Context(), key)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, setting)
}

// SetSettingRequest 设置请求.
type SetSettingRequest struct {
	Key         string `json:"key" binding:"required"`
	Value       string `json:"value"`
	ValueType   string `json:"value_type"`
	Category    string `json:"category"`
	Label       string `json:"label"`
	Description string `json:"description"`
}

// SetSetting 设置配置项.
func (h *Handler) SetSetting(c *gin.Context) {
	var req SetSettingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	setting, err := h.biz.Settings().Set(c.Request.Context(), &settings.SetRequest{
		Key:         req.Key,
		Value:       req.Value,
		ValueType:   req.ValueType,
		Category:    req.Category,
		Label:       req.Label,
		Description: req.Description,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, setting)
}

// DeleteSetting 删除设置.
func (h *Handler) DeleteSetting(c *gin.Context) {
	key := c.Param("key")
	if err := h.biz.Settings().Delete(c.Request.Context(), key); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// GetMultipleSettingsRequest 批量获取设置请求.
type GetMultipleSettingsRequest struct {
	Keys []string `json:"keys" binding:"required"`
}

// GetMultipleSettings 批量获取设置.
func (h *Handler) GetMultipleSettings(c *gin.Context) {
	var req GetMultipleSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.biz.Settings().GetMultiple(c.Request.Context(), req.Keys)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"settings": result})
}

// SetMultipleSettingsRequest 批量设置请求.
type SetMultipleSettingsRequest struct {
	Settings map[string]string `json:"settings" binding:"required"`
}

// SetMultipleSettings 批量设置.
func (h *Handler) SetMultipleSettings(c *gin.Context) {
	var req SetMultipleSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.biz.Settings().SetMultiple(c.Request.Context(), req.Settings); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "updated"})
}
