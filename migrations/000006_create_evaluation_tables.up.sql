-- 创建评估数据集表
CREATE TABLE IF NOT EXISTS evaluation_datasets (
    id VARCHAR(36) PRIMARY KEY,
    tenant_id BIGINT NOT NULL,
    name VARCHAR(200) NOT NULL,
    description TEXT,
    source_type VARCHAR(50) DEFAULT 'manual',

    -- Coze Loop 关联（如果使用云服务）
    coze_loop_workspace_id BIGINT,
    coze_loop_evaluation_set_id BIGINT,

    -- 统计
    item_count INT DEFAULT 0,
    version INT DEFAULT 1,

    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_evaluation_datasets_tenant_id ON evaluation_datasets(tenant_id);
CREATE INDEX IF NOT EXISTS idx_evaluation_datasets_name ON evaluation_datasets(name);
CREATE INDEX IF NOT EXISTS idx_evaluation_datasets_deleted_at ON evaluation_datasets(deleted_at);

-- 创建数据集条目表
CREATE TABLE IF NOT EXISTS dataset_items (
    id VARCHAR(36) PRIMARY KEY,
    dataset_id VARCHAR(36) NOT NULL,

    -- Query 输入
    query TEXT NOT NULL,
    query_id VARCHAR(100),

    -- Ground Truth: 检索部分
    relevant_doc_ids TEXT[],
    expected_doc_count INT DEFAULT 1,

    -- Ground Truth: 生成部分
    expected_answer TEXT,
    expected_answer_id VARCHAR(100),

    -- 元数据
    metadata JSON,

    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_dataset_items_dataset_id ON dataset_items(dataset_id);

-- 创建评估任务表
CREATE TABLE IF NOT EXISTS evaluation_tasks (
    id VARCHAR(36) PRIMARY KEY,
    tenant_id BIGINT NOT NULL,
    dataset_id VARCHAR(36) NOT NULL,

    -- 配置
    agent_id VARCHAR(36) NOT NULL,
    knowledge_base_id VARCHAR(36),

    -- Coze Loop 关联
    coze_loop_experiment_id BIGINT,

    -- 任务状态
    status VARCHAR(50) DEFAULT 'pending',
    progress INT DEFAULT 0,
    total_items INT DEFAULT 0,

    -- 错误信息
    error_message TEXT,

    -- 结果汇总
    avg_recall FLOAT,
    avg_precision FLOAT,
    avg_mrr FLOAT,
    avg_bleu FLOAT,

    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    started_at TIMESTAMP,
    completed_at TIMESTAMP
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_evaluation_tasks_tenant_id ON evaluation_tasks(tenant_id);
CREATE INDEX IF NOT EXISTS idx_evaluation_tasks_dataset_id ON evaluation_tasks(dataset_id);
CREATE INDEX IF NOT EXISTS idx_evaluation_tasks_agent_id ON evaluation_tasks(agent_id);
CREATE INDEX IF NOT EXISTS idx_evaluation_tasks_status ON evaluation_tasks(status);

-- 创建评估结果表
CREATE TABLE IF NOT EXISTS evaluation_results (
    id VARCHAR(36) PRIMARY KEY,
    task_id VARCHAR(36) NOT NULL,
    item_id VARCHAR(36) NOT NULL,

    -- 检索结果
    retrieved_doc_ids TEXT[],
    retrieval_latency_ms BIGINT,
    retrieval_ok BOOLEAN DEFAULT true,

    -- 生成结果
    generated_answer TEXT,
    generation_latency_ms BIGINT,
    generation_ok BOOLEAN DEFAULT true,
    total_tokens_used INT,

    -- 评估指标
    metric_recall FLOAT,
    metric_precision FLOAT,
    metric_mrr FLOAT,
    metric_bleu FLOAT,
    metric_rouge_rouge1 FLOAT,
    metric_rouge_rouge2 FLOAT,
    metric_rouge_rougel FLOAT,
    metric_answer_relevance FLOAT,
    metric_context_coverage FLOAT,

    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_evaluation_results_task_id ON evaluation_results(task_id);
CREATE INDEX IF NOT EXISTS idx_evaluation_results_item_id ON evaluation_results(item_id);

-- 添加注释
COMMENT ON TABLE evaluation_datasets IS '评估数据集：包含多个测试用例';
COMMENT ON TABLE dataset_items IS '数据集条目：单个测试用例';
COMMENT ON TABLE evaluation_tasks IS '评估任务：对数据集执行评估';
COMMENT ON TABLE evaluation_results IS '评估结果：单个测试用例的评估结果';

COMMENT ON COLUMN evaluation_datasets.source_type IS '数据集来源：manual=手动创建, file=文件导入, trace=从Trace导出';
COMMENT ON COLUMN evaluation_tasks.status IS '任务状态：pending=待执行, running=执行中, completed=已完成, failed=失败';

COMMENT ON COLUMN evaluation_results.metric_recall IS '召回率：检索到的相关文档数 / 总相关文档数';
COMMENT ON COLUMN evaluation_results.metric_precision IS '精确率：检索到的相关文档数 / 总检索文档数';
COMMENT ON COLUMN evaluation_results.metric_mrr IS '平均倒数排名：1 / 第一个相关文档的排名';
COMMENT ON COLUMN evaluation_results.metric_bleu IS 'BLEU分数：机器翻译质量指标';
COMMENT ON COLUMN evaluation_results.metric_rouge_rouge1 IS 'ROUGE-1分数：基于unigram的摘要质量';
COMMENT ON COLUMN evaluation_results.metric_rouge_rouge2 IS 'ROUGE-2分数：基于bigram的摘要质量';
COMMENT ON COLUMN evaluation_results.metric_rouge_rougel IS 'ROUGE-L分数：基于最长公共子序列的摘要质量';
