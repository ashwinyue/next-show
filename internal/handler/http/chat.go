// Package http 提供 HTTP Handler 层.
package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/ashwinyue/next-show/internal/pkg/sse"
)

// AgentChatRequest Agent 聊天请求（对齐 WeKnora）.
type AgentChatRequest struct {
	Query               string   `json:"query" binding:"required"`
	KnowledgeBaseIDs    []string `json:"knowledge_base_ids,omitempty"`
	AgentEnabled        bool     `json:"agent_enabled,omitempty"`
	WebSearchEnabled    bool     `json:"web_search_enabled,omitempty"`
	MCPServiceIDs       []string `json:"mcp_service_ids,omitempty"`
	MentionedItems      []MentionedItem `json:"mentioned_items,omitempty"`
}

// MentionedItem 提及的项目（知识库、文档等）.
type MentionedItem struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Type  string `json:"type"`
	KbType string `json:"kb_type,omitempty"`
}

// AgentChat Agent 聊天（对齐 WeKnora: /api/v1/agent-chat/:id）.
func (h *Handler) AgentChat(c *gin.Context) {
	sessionID := c.Param("session_id")

	var req AgentChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 生成消息 ID
	messageID := uuid.New().String()

	// 创建 SSE Writer
	writer := sse.NewGinWriter(c)
	writer.SetHeaders()

	// 发送开始事件
	if err := writer.SendStart(sessionID, messageID); err != nil {
		writer.SendError("failed to send start event")
		return
	}

	// 调用 Agent 业务层（事件已在 SSE adapter 中处理）
	err := h.biz.Agents().Chat(c.Request.Context(), sessionID, req.Query, writer)

	// 发送完成事件
	_ = writer.SendComplete(sessionID, messageID)

	// 持久化消息
	ctx := c.Request.Context()

	// 保存用户消息
	_, _ = h.biz.Sessions().AddMessage(ctx, sessionID, "user", req.Query)

	// 检查是否有错误
	if err != nil {
		// 错误已在 SSE 中发送，这里不需要再处理
		return
	}
}
