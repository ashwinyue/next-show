// Package store 提供数据访问层.
package store

import "gorm.io/gorm"

// Store 存储层聚合接口.
type Store interface {
	Providers() ProviderStore
	Agents() AgentStore
	AgentRelations() AgentRelationStore
	Sessions() SessionStore
	Messages() MessageStore
	Checkpoints() CheckpointStore
	MCPServers() MCPServerStore
	MCPTools() MCPToolStore
	AgentTools() AgentToolStore
	Knowledge() KnowledgeStore
	WebSearch() WebSearchStore
	Settings() SettingsStore
	Tenants() TenantStore
	Users() UserStore
	Skills() SkillStore
	// DB 返回底层数据库连接（用于事务等场景）
	DB() *gorm.DB
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

func (s *dataStore) AgentRelations() AgentRelationStore {
	return newAgentRelationStore(s.db)
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

func (s *dataStore) Knowledge() KnowledgeStore {
	return newKnowledgeStore(s.db)
}

func (s *dataStore) WebSearch() WebSearchStore {
	return newWebSearchStore(s.db)
}

func (s *dataStore) Settings() SettingsStore {
	return newSettingsStore(s.db)
}

func (s *dataStore) Tenants() TenantStore {
	return newTenantStore(s.db)
}

func (s *dataStore) Users() UserStore {
	return newUserStore(s.db)
}

func (s *dataStore) Skills() SkillStore {
	return newSkillStore(s.db)
}

func (s *dataStore) DB() *gorm.DB {
	return s.db
}
