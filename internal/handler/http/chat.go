// Package http 提供 HTTP Handler 层.
package http

import (
	"net/http"
	"strings"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/schema"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/mervyn/next-show/internal/pkg/sse"
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

	// 创建事件处理器
	processor := sse.NewDefaultEventProcessor()

	// 创建事件上下文
	contentBuffer := &strings.Builder{}
	eventCtx := &sse.EventContext{
		SessionID:     req.SessionID,
		MessageID:     messageID,
		ContentBuffer: contentBuffer,
	}

	// 发送开始事件
	if err := writer.SendStart(req.SessionID, messageID); err != nil {
		writer.SendError("failed to send start event")
		return
	}

	// 构建用户消息
	userMessage := schema.UserMessage(req.Message)

	// 调用 Agent 业务层
	iter, err := h.biz.Agents().Chat(c.Request.Context(), req.SessionID, []adk.Message{userMessage})
	if err != nil {
		_ = writer.Send(sse.Event{
			Type:      sse.EventTypeError,
			ID:        messageID,
			Error:     err.Error(),
			SessionID: req.SessionID,
		})
		return
	}

	// 遍历事件流
	for {
		event, ok := iter.Next()
		if !ok {
			break
		}

		// 使用 EventProcessor 处理事件
		sseEvents, err := processor.Process(event, eventCtx)
		if err != nil {
			_ = writer.Send(sse.Event{
				Type:      sse.EventTypeError,
				ID:        messageID,
				Error:     err.Error(),
				SessionID: req.SessionID,
			})
			return
		}

		// 发送 SSE 事件
		for _, sseEvent := range sseEvents {
			if err := writer.Send(sseEvent); err != nil {
				return
			}
		}
	}

	// 发送完成事件
	_ = writer.SendComplete(req.SessionID, messageID)

	// TODO: 持久化消息（用户消息 + 助手回复）
	// h.biz.Sessions().AddMessage(ctx, req.SessionID, "user", req.Message)
	// h.biz.Sessions().AddMessage(ctx, req.SessionID, "assistant", contentBuffer.String())
}
