// Package biz 提供业务逻辑层.
package biz

import (
	"github.com/ashwinyue/next-show/internal/biz/agent"
	"github.com/ashwinyue/next-show/internal/biz/knowledge"
	"github.com/ashwinyue/next-show/internal/biz/session"
	"github.com/ashwinyue/next-show/internal/pkg/agent/factory"
	"github.com/ashwinyue/next-show/internal/store"
)

// Biz 业务层聚合接口.
type Biz interface {
	Agents() agent.AgentBiz
	Sessions() session.SessionBiz
	Knowledge() knowledge.Biz
}

type biz struct {
	agentBiz     agent.AgentBiz
	sessionBiz   session.SessionBiz
	knowledgeBiz knowledge.Biz
}

// NewBiz 创建业务层实例.
func NewBiz(store store.Store, agentFactory *factory.AgentFactory) Biz {
	return &biz{
		agentBiz:     agent.NewAgentBiz(store, agentFactory),
		sessionBiz:   session.NewSessionBiz(store),
		knowledgeBiz: knowledge.NewBiz(store),
	}
}

func (b *biz) Agents() agent.AgentBiz {
	return b.agentBiz
}

func (b *biz) Sessions() session.SessionBiz {
	return b.sessionBiz
}

func (b *biz) Knowledge() knowledge.Biz {
	return b.knowledgeBiz
}
