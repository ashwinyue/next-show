// Package http 提供 HTTP Handler 层.
package http

import (
	"github.com/gin-gonic/gin"

	"github.com/mervyn/next-show/internal/biz"
)

// Handler HTTP 处理器聚合.
type Handler struct {
	biz biz.Biz
}

// NewHandler 创建 Handler 实例.
func NewHandler(b biz.Biz) *Handler {
	return &Handler{biz: b}
}

// RegisterRoutes 注册 HTTP 路由.
func (h *Handler) RegisterRoutes(r *gin.Engine) {
	// API v1 路由组
	v1 := r.Group("/api/v1")
	{
		// Session 路由
		sessions := v1.Group("/sessions")
		{
			sessions.POST("", h.CreateSession)
			sessions.GET("", h.ListSessions)
			sessions.GET("/:id", h.GetSession)
			sessions.DELETE("/:id", h.DeleteSession)
			sessions.GET("/:id/messages", h.GetMessages)
		}

		// Chat 路由
		v1.POST("/chat", h.Chat)
		v1.POST("/chat/stream", h.ChatStream)
	}

	// 健康检查
	r.GET("/health", h.Health)
}

// Health 健康检查.
func (h *Handler) Health(c *gin.Context) {
	c.JSON(200, gin.H{"status": "ok"})
}
