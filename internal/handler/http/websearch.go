// Package http 提供 HTTP Handler 层.
package http

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/ashwinyue/next-show/internal/biz/websearch"
)

// GetWebSearchConfig 获取网络搜索配置（对齐 WeKnora）.
func (h *Handler) GetWebSearchConfig(c *gin.Context) {
	config, err := h.biz.WebSearch().GetConfig(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, config)
}

// UpdateWebSearchConfigRequest 更新网络搜索配置请求.
type UpdateWebSearchConfigRequest struct {
	Provider   string   `json:"provider"`
	APIKey     string   `json:"api_key,omitempty"`
	MaxResults int      `json:"max_results"`
	Blacklist  []string `json:"blacklist,omitempty"`
}

// UpdateWebSearchConfig 更新网络搜索配置（对齐 WeKnora）.
func (h *Handler) UpdateWebSearchConfig(c *gin.Context) {
	var req UpdateWebSearchConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	config, err := h.biz.WebSearch().UpdateConfig(c.Request.Context(), &websearch.UpdateConfigRequest{
		Provider:   req.Provider,
		APIKey:     req.APIKey,
		MaxResults: req.MaxResults,
		Blacklist:  req.Blacklist,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, config)
}
