# next-show 功能路线图

## 已实现功能

### Agent 系统
- ✅ Agent 管理（CRUD、多类型：ChatModel/React/Supervisor/Sequential/RAG）
- ✅ Provider 管理（多模型厂商：Ark/OpenAI）
- ✅ Session/Message 管理
- ✅ Supervisor Agent（多子 Agent 调度）
- ✅ RAG Agent（知识库增强问答）
- ✅ MCP 工具集成
- ✅ Web Search 工具（DuckDuckGo）
- ✅ SSE 流式输出
- ✅ Tracing（CozeLoop/本地日志）

### 知识库系统
- ✅ 知识库管理（CRUD）
- ✅ 文档导入（URL/纯文本）
- ✅ 递归分块（Recursive Splitter）
- ✅ Embedding（DashScope/OpenAI）
- ✅ 向量搜索（pgvector）
- ✅ 语义检索 API

---

## 待实现功能（对比 WeKnora）

### 🔴 高优先级

| 功能 | 说明 | eino-ext 组件 |
|------|------|---------------|
| PDF 解析 | 支持 PDF 文档导入 | `document/parser/pdf` |
| Word 解析 | 支持 DOCX 文档导入 | `document/parser/docx` |
| Excel 解析 | 支持 XLSX 文档导入 | `document/parser/xlsx` |
| 文件加载器 | 本地文件批量加载 | `document/loader/file` |

### 🟡 中优先级

| 功能 | 说明 | eino-ext 组件 |
|------|------|---------------|
| FAQ 知识库 | 问答对形式的知识库 | 自行实现 |
| BM25 检索 | 关键词检索，混合检索策略 | 自行实现 |
| Rerank 重排序 | 二次排序提升召回质量 | `transformer/reranker/score` |
| 语义分块 | 基于语义的智能分块 | `transformer/splitter/semantic` |
| GraphRAG | 知识图谱增强检索 | 自行实现 |
| Evaluation | 检索+生成质量评估 | 自行实现 |

### 🟢 低优先级

| 功能 | 说明 | eino-ext 组件 |
|------|------|---------------|
| 多租户/用户管理 | 用户鉴权、租户隔离 | 自行实现 |
| 标签管理 | 文档/知识库标签 | 自行实现 |
| 异步任务队列 | MQ 任务管理 | 自行实现 |
| S3 文件存储 | 云端文件存储 | `document/loader/s3` |
| 前端 UI | Web 管理界面 | 自行实现 |

---

## eino-ext 可用组件

### 文档加载器 (Loader)
```
github.com/cloudwego/eino-ext/components/document/loader/
├── file/    # 本地文件加载
├── s3/      # S3 文件加载
└── url/     # URL 加载 ✅ 已集成
```

### 文档解析器 (Parser)
```
github.com/cloudwego/eino-ext/components/document/parser/
├── pdf/     # PDF 解析
├── docx/    # Word 解析
├── xlsx/    # Excel 解析
└── html/    # HTML 解析 ✅ 已集成
```

### 文档转换器 (Transformer)
```
github.com/cloudwego/eino-ext/components/document/transformer/
├── splitter/
│   ├── recursive/   # 递归分块 ✅ 已集成
│   ├── semantic/    # 语义分块
│   ├── markdown/    # Markdown 分块
│   └── html/        # HTML 分块
└── reranker/
    └── score/       # 分数重排序
```

### Embedding 模型
```
github.com/cloudwego/eino-ext/components/embedding/
├── dashscope/   # 阿里云 DashScope ✅ 已集成
├── openai/      # OpenAI ✅ 已集成
├── ark/         # 火山引擎 Ark
├── ollama/      # Ollama 本地模型
├── gemini/      # Google Gemini
└── qianfan/     # 百度千帆
```

### 检索器 (Retriever)
```
github.com/cloudwego/eino-ext/components/retriever/
├── es7/         # Elasticsearch 7
├── es8/         # Elasticsearch 8
├── milvus/      # Milvus
├── qdrant/      # Qdrant
├── redis/       # Redis
├── opensearch2/ # OpenSearch 2
└── opensearch3/ # OpenSearch 3
```
> 注意：无 pgvector retriever，已自行实现

### 索引器 (Indexer)
```
github.com/cloudwego/eino-ext/components/indexer/
├── es7/         # Elasticsearch 7
├── es8/         # Elasticsearch 8
├── milvus/      # Milvus
├── qdrant/      # Qdrant
├── redis/       # Redis
└── volc_vikingdb/ # 火山引擎 VikingDB
```

### 工具 (Tool)
```
github.com/cloudwego/eino-ext/components/tool/
├── duckduckgo/  # DuckDuckGo 搜索 ✅ 已集成
├── mcp/         # MCP 工具 ✅ 已集成
├── bingsearch/  # Bing 搜索
└── browseruse/  # 浏览器自动化
```

### 模型 (Model)
```
github.com/cloudwego/eino-ext/components/model/
├── ark/         # 火山引擎 Ark ✅ 已集成
├── openai/      # OpenAI ✅ 已集成
├── claude/      # Anthropic Claude
├── deepseek/    # DeepSeek
├── gemini/      # Google Gemini
├── ollama/      # Ollama
├── qwen/        # 通义千问
└── qianfan/     # 百度千帆
```

---

## 实现建议

### 第一阶段：文档格式支持 ✅ 已完成
1. ✅ 集成 PDF Parser
2. ✅ 集成 DOCX Parser
3. ✅ 集成 File Loader
4. ✅ 更新导入 API 支持文件上传

### 第二阶段：检索质量优化 ✅ 已完成
1. ✅ 实现 BM25 全文搜索（PostgreSQL ts_rank_cd）
2. ✅ 实现混合检索（BM25 + 向量，加权融合）
3. ✅ 集成 Score Reranker 重排序（首尾效应优化）

### 第二阶段 B：数据分析师 ✅ 已完成
1. ✅ CSV/XLSX 原文件本地落盘（data/files/<kbID>/<docID>/）
2. ✅ 扩展 parseFile 支持 xlsx/csv 解析
3. ✅ 实现 DuckDB 工具（data_schema、data_analysis）
4. ✅ 新增内置 Agent 模板：builtin-data-analyst

### 第二阶段 C：多模式 Agent 架构 ✅ 已完成
1. ✅ 定义主控/子 Agent 角色（Orchestrator/Specialist）
2. ✅ 内置主控 Agent：Supervisor、Deep、Plan-Execute
3. ✅ 内置子 Agent：RAG、DataAnalyst
4. ✅ AgentFactory 支持根据类型创建对应 ADK Agent
5. ✅ 集成语义分块（Semantic Chunking）

### 第三阶段：高级功能
1. FAQ 知识库
2. 评估系统
3. 前端 UI
