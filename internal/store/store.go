// Package store 提供数据访问层.
package store

import "gorm.io/gorm"

// Store 存储层聚合接口.
type Store interface {
	Providers() ProviderStore
	Agents() AgentStore
	Sessions() SessionStore
	Messages() MessageStore
	Checkpoints() CheckpointStore
	MCPServers() MCPServerStore
	MCPTools() MCPToolStore
	AgentTools() AgentToolStore
}

// dataStore 存储层实现.
type dataStore struct {
	db *gorm.DB
}

// NewStore 创建存储层实例.
func NewStore(db *gorm.DB) Store {
	return &dataStore{db: db}
}

func (s *dataStore) Providers() ProviderStore {
	return newProviderStore(s.db)
}

func (s *dataStore) Agents() AgentStore {
	return newAgentStore(s.db)
}

func (s *dataStore) Sessions() SessionStore {
	return newSessionStore(s.db)
}

func (s *dataStore) Messages() MessageStore {
	return newMessageStore(s.db)
}

func (s *dataStore) Checkpoints() CheckpointStore {
	return newCheckpointStore(s.db)
}

func (s *dataStore) MCPServers() MCPServerStore {
	return newMCPServerStore(s.db)
}

func (s *dataStore) MCPTools() MCPToolStore {
	return newMCPToolStore(s.db)
}

func (s *dataStore) AgentTools() AgentToolStore {
	return newAgentToolStore(s.db)
}
