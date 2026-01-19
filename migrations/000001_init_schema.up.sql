-- 启用 pgvector 扩展
CREATE EXTENSION IF NOT EXISTS vector;
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- 创建 providers 表
CREATE TABLE IF NOT EXISTS providers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) UNIQUE NOT NULL,
    display_name VARCHAR(200) NOT NULL,
    provider_type VARCHAR(50) NOT NULL DEFAULT 'openai',
    api_key TEXT,
    base_url TEXT,
    config JSONB DEFAULT '{}',
    is_enabled BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 创建 agents 表
CREATE TABLE IF NOT EXISTS agents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) UNIQUE NOT NULL,
    display_name VARCHAR(200) NOT NULL,
    description TEXT,
    provider_id UUID NOT NULL REFERENCES providers(id),
    model_name VARCHAR(200) NOT NULL,
    system_prompt TEXT,
    agent_type VARCHAR(50) NOT NULL DEFAULT 'chat_model',
    max_iterations INT DEFAULT 10,
    temperature DECIMAL(3,2),
    max_tokens INT,
    config JSONB DEFAULT '{}',
    is_enabled BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_agents_provider_id ON agents(provider_id);
CREATE INDEX idx_agents_agent_type ON agents(agent_type);
CREATE INDEX idx_agents_is_enabled ON agents(is_enabled);

-- 创建 agent_relations 表 (组合型 Agent)
CREATE TABLE IF NOT EXISTS agent_relations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    parent_agent_id UUID NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
    child_agent_id UUID NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
    role VARCHAR(50) NOT NULL,
    sort_order INT DEFAULT 0,
    config JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_agent_relations_parent ON agent_relations(parent_agent_id);
CREATE INDEX idx_agent_relations_child ON agent_relations(child_agent_id);

-- 创建 sessions 表
CREATE TABLE IF NOT EXISTS sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_id UUID NOT NULL REFERENCES agents(id),
    user_id VARCHAR(100),
    title VARCHAR(500),
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    metadata JSONB DEFAULT '{}',
    context JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_sessions_agent_id ON sessions(agent_id);
CREATE INDEX idx_sessions_user_id ON sessions(user_id);
CREATE INDEX idx_sessions_status ON sessions(status);
CREATE INDEX idx_sessions_created_at ON sessions(created_at);

-- 创建 messages 表
CREATE TABLE IF NOT EXISTS messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id UUID NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    role VARCHAR(20) NOT NULL,
    content TEXT NOT NULL,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_messages_session_id ON messages(session_id);
CREATE INDEX idx_messages_created_at ON messages(created_at);

-- 创建 checkpoints 表 (用于 Agent 状态持久化)
CREATE TABLE IF NOT EXISTS checkpoints (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id UUID NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    agent_name VARCHAR(100) NOT NULL,
    state JSONB NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(session_id, agent_name)
);

CREATE INDEX idx_checkpoints_session_id ON checkpoints(session_id);

-- 创建 knowledge_bases 表
CREATE TABLE IF NOT EXISTS knowledge_bases (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    chunking_config JSONB DEFAULT '{}',
    parser_config JSONB DEFAULT '{}',
    indexer_type VARCHAR(50),
    indexer_config JSONB DEFAULT '{}',
    embedding_config JSONB DEFAULT '{}',
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_knowledge_bases_status ON knowledge_bases(status);

-- 创建 knowledge_documents 表
CREATE TABLE IF NOT EXISTS knowledge_documents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    knowledge_base_id UUID NOT NULL REFERENCES knowledge_bases(id) ON DELETE CASCADE,
    source_type VARCHAR(20) NOT NULL,
    title VARCHAR(255),
    source_uri TEXT,
    file_hash VARCHAR(64),
    content_text TEXT,
    metadata JSONB DEFAULT '{}',
    parse_status VARCHAR(20) NOT NULL DEFAULT 'pending',
    error_message TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_knowledge_documents_kb ON knowledge_documents(knowledge_base_id);
CREATE INDEX idx_knowledge_documents_file_hash ON knowledge_documents(file_hash);
CREATE INDEX idx_knowledge_documents_parse_status ON knowledge_documents(parse_status);

-- 创建 knowledge_chunks 表
CREATE TABLE IF NOT EXISTS knowledge_chunks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    knowledge_base_id UUID NOT NULL REFERENCES knowledge_bases(id) ON DELETE CASCADE,
    document_id UUID NOT NULL REFERENCES knowledge_documents(id) ON DELETE CASCADE,
    chunk_index INT NOT NULL,
    content TEXT NOT NULL,
    content_hash VARCHAR(64),
    metadata JSONB DEFAULT '{}',
    is_enabled BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_knowledge_chunks_kb ON knowledge_chunks(knowledge_base_id);
CREATE INDEX idx_knowledge_chunks_doc ON knowledge_chunks(document_id);
CREATE INDEX idx_knowledge_chunks_content_hash ON knowledge_chunks(content_hash);
CREATE INDEX idx_knowledge_chunks_is_enabled ON knowledge_chunks(is_enabled);

-- 创建全文搜索索引
CREATE INDEX idx_knowledge_chunks_content_trgm ON knowledge_chunks USING gin (content gin_trgm_ops);

-- 创建 embeddings 表 (pgvector)
CREATE TABLE IF NOT EXISTS embeddings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    knowledge_base_id UUID NOT NULL REFERENCES knowledge_bases(id) ON DELETE CASCADE,
    chunk_id UUID NOT NULL REFERENCES knowledge_chunks(id) ON DELETE CASCADE UNIQUE,
    embedding vector(1024) NOT NULL,
    embedding_dim INT NOT NULL DEFAULT 1024,
    embedding_model VARCHAR(128) DEFAULT 'dashscope/text-embedding-v3',
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_embeddings_kb ON embeddings(knowledge_base_id);
CREATE INDEX idx_embeddings_chunk ON embeddings(chunk_id);

-- 创建向量索引 (IVFFlat for faster search)
CREATE INDEX idx_embeddings_vector ON embeddings USING ivfflat (embedding vector_cosine_ops) WITH (lists = 100);

-- 创建 mcp_servers 表
CREATE TABLE IF NOT EXISTS mcp_servers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) UNIQUE NOT NULL,
    display_name VARCHAR(200) NOT NULL,
    description TEXT,
    server_type VARCHAR(50) NOT NULL DEFAULT 'stdio',
    command TEXT,
    args TEXT[],
    env JSONB DEFAULT '{}',
    config JSONB DEFAULT '{}',
    is_enabled BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 创建 mcp_tools 表
CREATE TABLE IF NOT EXISTS mcp_tools (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    server_id UUID NOT NULL REFERENCES mcp_servers(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    input_schema JSONB DEFAULT '{}',
    is_enabled BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(server_id, name)
);

CREATE INDEX idx_mcp_tools_server ON mcp_tools(server_id);

-- 创建 agent_tools 表 (Agent 与工具的关联)
CREATE TABLE IF NOT EXISTS agent_tools (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_id UUID NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
    tool_type VARCHAR(50) NOT NULL,
    builtin_tool_name VARCHAR(100),
    mcp_tool_id UUID REFERENCES mcp_tools(id) ON DELETE CASCADE,
    custom_tool_config JSONB DEFAULT '{}',
    is_enabled BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_agent_tools_agent ON agent_tools(agent_id);
CREATE INDEX idx_agent_tools_mcp_tool ON agent_tools(mcp_tool_id);
