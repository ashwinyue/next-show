// Package http 提供 HTTP Handler 层.
package http

import (
	"github.com/gin-gonic/gin"

	"github.com/ashwinyue/next-show/internal/biz"
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

	// Session 路由
	h.registerSessionRoutes(v1)

	// Chat 路由
	h.registerChatRoutes(v1)

	// Knowledge 路由
	h.registerKnowledgeRoutes(v1)

	// 健康检查
	r.GET("/health", h.Health)
}

// registerSessionRoutes 注册 Session 路由.
func (h *Handler) registerSessionRoutes(r *gin.RouterGroup) {
	sessions := r.Group("/sessions")
	{
		sessions.POST("", h.CreateSession)
		sessions.GET("", h.ListSessions)
		sessions.GET("/:id", h.GetSession)
		sessions.DELETE("/:id", h.DeleteSession)
		sessions.GET("/:id/messages", h.GetMessages)
	}
}

// registerChatRoutes 注册 Chat 路由.
func (h *Handler) registerChatRoutes(r *gin.RouterGroup) {
	r.POST("/chat", h.Chat)
	r.POST("/chat/stream", h.ChatStream)
}

// registerKnowledgeRoutes 注册 Knowledge 路由.
func (h *Handler) registerKnowledgeRoutes(r *gin.RouterGroup) {
	kb := r.Group("/knowledge-bases")
	{
		kb.POST("", h.CreateKnowledgeBase)
		kb.GET("", h.ListKnowledgeBases)
		kb.GET("/:id", h.GetKnowledgeBase)
		kb.PUT("/:id", h.UpdateKnowledgeBase)
		kb.DELETE("/:id", h.DeleteKnowledgeBase)

		// Documents
		kb.POST("/:id/documents", h.CreateDocument)
		kb.GET("/:id/documents", h.ListDocuments)
	}

	doc := r.Group("/documents")
	{
		doc.GET("/:id", h.GetDocument)
		doc.DELETE("/:id", h.DeleteDocument)
		doc.GET("/:id/chunks", h.ListChunks)
	}
}

// Health 健康检查.
func (h *Handler) Health(c *gin.Context) {
	c.JSON(200, gin.H{"status": "ok"})
}
