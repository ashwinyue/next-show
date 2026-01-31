// Package http 提供 HTTP Handler 层.
package http

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/ashwinyue/next-show/internal/pkg/sse"
)

// ChatRequest Chat 请求.
type ChatRequest struct {
	SessionID string `json:"session_id" binding:"required"`
	Message   string `json:"message" binding:"required"`
	Stream    bool   `json:"stream"`
}

// Chat 非流式对话.
func (h *Handler) Chat(c *gin.Context) {
	var req ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: 实现非流式对话（收集所有事件后返回）
	c.JSON(http.StatusOK, gin.H{
		"message": "chat endpoint - not implemented yet",
		"request": req,
	})
}

// ChatStream 流式对话（SSE）.
func (h *Handler) ChatStream(c *gin.Context) {
	var req ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 生成消息 ID
	messageID := uuid.New().String()

	// 创建 SSE Writer
	writer := sse.NewGinWriter(c)
	writer.SetHeaders()

	// 创建内容缓冲器用于保存助手回复
	contentBuffer := &strings.Builder{}

	// 发送开始事件
	if err := writer.SendStart(req.SessionID, messageID); err != nil {
		writer.SendError("failed to send start event")
		return
	}

	// 调用 Agent 业务层（事件已在 SSE adapter 中处理）
	err := h.biz.Agents().Chat(c.Request.Context(), req.SessionID, req.Message, writer)

	// 发送完成事件
	_ = writer.SendComplete(req.SessionID, messageID)

	// 持久化消息
	ctx := c.Request.Context()

	// 保存用户消息
	_, _ = h.biz.Sessions().AddMessage(ctx, req.SessionID, "user", req.Message)

	// 保存助手回复
	assistantContent := contentBuffer.String()
	if assistantContent != "" {
		_, _ = h.biz.Sessions().AddMessage(ctx, req.SessionID, "assistant", assistantContent)
	}

	// 检查是否有错误
	if err != nil {
		// 错误已在 Chat 方法中通过 SSE 发送，这里不需要再处理
		return
	}
}
