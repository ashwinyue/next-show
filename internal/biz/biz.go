// Package biz 提供业务逻辑层.
package biz

import (
	"github.com/mervyn/next-show/internal/biz/agent"
	"github.com/mervyn/next-show/internal/biz/session"
	"github.com/mervyn/next-show/internal/store"
)

// Biz 业务层聚合接口.
type Biz interface {
	Agents() agent.AgentBiz
	Sessions() session.SessionBiz
}

type biz struct {
	agentBiz   agent.AgentBiz
	sessionBiz session.SessionBiz
}

// NewBiz 创建业务层实例.
func NewBiz(store store.Store) Biz {
	return &biz{
		agentBiz:   agent.NewAgentBiz(store),
		sessionBiz: session.NewSessionBiz(store),
	}
}

func (b *biz) Agents() agent.AgentBiz {
	return b.agentBiz
}

func (b *biz) Sessions() session.SessionBiz {
	return b.sessionBiz
}
