-- 删除 agent_tools 表
DROP TABLE IF EXISTS agent_tools;

-- 删除 mcp_tools 表
DROP TABLE IF EXISTS mcp_tools;

-- 删除 mcp_servers 表
DROP TABLE IF EXISTS mcp_servers;

-- 删除 embeddings 表
DROP TABLE IF EXISTS embeddings;

-- 删除 knowledge_chunks 表
DROP TABLE IF EXISTS knowledge_chunks;

-- 删除 knowledge_documents 表
DROP TABLE IF EXISTS knowledge_documents;

-- 删除 knowledge_bases 表
DROP TABLE IF EXISTS knowledge_bases;

-- 删除 checkpoints 表
DROP TABLE IF EXISTS checkpoints;

-- 删除 messages 表
DROP TABLE IF EXISTS messages;

-- 删除 sessions 表
DROP TABLE IF EXISTS sessions;

-- 删除 agent_relations 表
DROP TABLE IF EXISTS agent_relations;

-- 删除 agents 表
DROP TABLE IF EXISTS agents;

-- 删除 providers 表
DROP TABLE IF EXISTS providers;

-- 删除扩展
DROP EXTENSION IF EXISTS pg_trgm;
DROP EXTENSION IF EXISTS vector;
