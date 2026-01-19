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

	// Agent 路由
	h.registerAgentRoutes(v1)

	// Session 路由
	h.registerSessionRoutes(v1)

	// Chat 路由
	h.registerChatRoutes(v1)

	// Knowledge 路由
	h.registerKnowledgeRoutes(v1)

	// Tenant 路由
	h.registerTenantRoutes(v1)

	// 健康检查
	r.GET("/health", h.Health)
}

// registerAgentRoutes 注册 Agent 路由.
func (h *Handler) registerAgentRoutes(r *gin.RouterGroup) {
	agents := r.Group("/agents")
	{
		agents.GET("", h.ListAgents)
		agents.POST("", h.CreateAgent)
		agents.GET("/builtin", h.ListBuiltinAgents)
		agents.GET("/orchestrators", h.ListOrchestratorAgents)
		agents.GET("/specialists", h.ListSpecialistAgents)
		agents.GET("/:id", h.GetAgent)
		agents.PUT("/:id", h.UpdateAgent)
		agents.DELETE("/:id", h.DeleteAgent)
		agents.GET("/:id/relations", h.GetAgentRelations)
		agents.PUT("/:id/relations", h.SetAgentRelations)

		// Agent Tools
		agents.GET("/:id/tools", h.ListAgentTools)
		agents.POST("/:id/tools", h.AddAgentTool)
		agents.PUT("/:id/tools/:tool_id", h.UpdateAgentTool)
		agents.DELETE("/:id/tools/:tool_id", h.RemoveAgentTool)
	}

	// 内置工具列表
	r.GET("/tools/builtin", h.ListBuiltinTools)

	// Provider 路由
	h.registerProviderRoutes(r)
}

// registerProviderRoutes 注册 Provider 路由.
func (h *Handler) registerProviderRoutes(r *gin.RouterGroup) {
	providers := r.Group("/providers")
	{
		providers.GET("", h.ListProviders)
		providers.POST("", h.CreateProvider)
		providers.GET("/chat", h.ListChatProviders)
		providers.GET("/embedding", h.ListEmbeddingProviders)
		providers.GET("/rerank", h.ListRerankProviders)
		providers.GET("/:id", h.GetProvider)
		providers.PUT("/:id", h.UpdateProvider)
		providers.DELETE("/:id", h.DeleteProvider)
	}

	// MCP 路由
	h.registerMCPRoutes(r)
}

// registerMCPRoutes 注册 MCP 路由.
func (h *Handler) registerMCPRoutes(r *gin.RouterGroup) {
	mcpServers := r.Group("/mcp-servers")
	{
		mcpServers.GET("", h.ListMCPServers)
		mcpServers.POST("", h.CreateMCPServer)
		mcpServers.GET("/:id", h.GetMCPServer)
		mcpServers.PUT("/:id", h.UpdateMCPServer)
		mcpServers.DELETE("/:id", h.DeleteMCPServer)

		// MCP Tools
		mcpServers.GET("/:id/tools", h.ListMCPTools)
		mcpServers.POST("/:id/tools", h.CreateMCPTool)
		mcpServers.GET("/:id/tools/:tool_id", h.GetMCPTool)
		mcpServers.PUT("/:id/tools/:tool_id", h.UpdateMCPTool)
		mcpServers.DELETE("/:id/tools/:tool_id", h.DeleteMCPTool)
	}

	// WebSearch 路由
	h.registerWebSearchRoutes(r)
}

// registerWebSearchRoutes 注册 WebSearch 路由.
func (h *Handler) registerWebSearchRoutes(r *gin.RouterGroup) {
	webSearch := r.Group("/web-search")
	{
		webSearch.GET("", h.ListWebSearchConfigs)
		webSearch.POST("", h.CreateWebSearchConfig)
		webSearch.GET("/default", h.GetDefaultWebSearchConfig)
		webSearch.GET("/:id", h.GetWebSearchConfig)
		webSearch.PUT("/:id", h.UpdateWebSearchConfig)
		webSearch.DELETE("/:id", h.DeleteWebSearchConfig)
		webSearch.PUT("/:id/default", h.SetDefaultWebSearchConfig)
	}

	// Settings 路由
	h.registerSettingsRoutes(r)
}

// registerTenantRoutes 注册租户和 API Key 路由.
func (h *Handler) registerTenantRoutes(r *gin.RouterGroup) {
	// 租户管理
	tenants := r.Group("/tenants")
	{
		tenants.GET("", h.ListTenants)
		tenants.POST("", h.CreateTenant)
		tenants.GET("/:id", h.GetTenant)
		tenants.PUT("/:id", h.UpdateTenant)
		tenants.DELETE("/:id", h.DeleteTenant)

		// API Keys
		tenants.GET("/:tenant_id/api-keys", h.ListAPIKeys)
		tenants.POST("/:tenant_id/api-keys", h.CreateAPIKey)
		tenants.GET("/:tenant_id/api-keys/:key_id", h.GetAPIKey)
		tenants.POST("/:tenant_id/api-keys/:key_id/revoke", h.RevokeAPIKey)
		tenants.DELETE("/:tenant_id/api-keys/:key_id", h.DeleteAPIKey)
	}
}

// registerSettingsRoutes 注册 Settings 路由.
func (h *Handler) registerSettingsRoutes(r *gin.RouterGroup) {
	// 系统信息
	r.GET("/system/info", h.GetSystemInfo)

	// 系统设置
	settings := r.Group("/settings")
	{
		settings.GET("", h.ListSettings)
		settings.POST("", h.SetSetting)
		settings.POST("/batch", h.SetMultipleSettings)
		settings.POST("/batch/get", h.GetMultipleSettings)
		settings.GET("/:key", h.GetSetting)
		settings.DELETE("/:key", h.DeleteSetting)
	}
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
	knowledge := r.Group("/knowledge-bases")
	{
		knowledge.POST("", h.CreateKnowledgeBase)
		knowledge.GET("", h.ListKnowledgeBases)
		knowledge.GET("/:id", h.GetKnowledgeBase)
		knowledge.PUT("/:id", h.UpdateKnowledgeBase)
		knowledge.DELETE("/:id", h.DeleteKnowledgeBase)

		// Documents
		knowledge.GET("/:id/documents", h.ListDocuments)
		knowledge.POST("/:id/documents", h.ImportDocument)
		knowledge.POST("/:id/documents/upload", h.UploadDocument)
		knowledge.DELETE("/:id/documents/:doc_id", h.DeleteDocument)
		knowledge.GET("/:id/documents/:doc_id/chunks", h.ListChunks)

		// Search
		knowledge.POST("/:id/search", h.SearchKnowledgeBase)
	}

	// Chunk & Tag 路由
	h.registerChunkTagRoutes(r)
}

// registerChunkTagRoutes 注册分块和标签路由.
func (h *Handler) registerChunkTagRoutes(r *gin.RouterGroup) {
	// 知识库下的标签
	kbTags := r.Group("/knowledge-bases/:kb_id/tags")
	{
		kbTags.GET("", h.ListTags)
		kbTags.POST("", h.CreateTag)
		kbTags.GET("/:tag_id", h.GetTag)
		kbTags.PUT("/:tag_id", h.UpdateTag)
		kbTags.DELETE("/:tag_id", h.DeleteTag)
		kbTags.GET("/:tag_id/chunks", h.ListChunksByTag)
	}

	// 知识库下的分块
	kbChunks := r.Group("/knowledge-bases/:kb_id/chunks")
	{
		kbChunks.GET("", h.ListChunksByKB)
		kbChunks.GET("/:chunk_id", h.GetChunkDetail)
		kbChunks.PUT("/:chunk_id", h.UpdateChunkHandler)
		kbChunks.DELETE("/:chunk_id", h.DeleteChunkHandler)
		kbChunks.GET("/:chunk_id/tags", h.ListChunkTagsHandler)
		kbChunks.POST("/:chunk_id/tags", h.AddChunkTag)
		kbChunks.DELETE("/:chunk_id/tags/:tag_id", h.RemoveChunkTag)
	}
}

// Health 健康检查.
func (h *Handler) Health(c *gin.Context) {
	c.JSON(200, gin.H{"status": "ok"})
}
