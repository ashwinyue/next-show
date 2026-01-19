# Next-Show 待实现功能

## 已完成功能 ✅

### 核心架构
- [x] 主控/子 Agent 架构定义
- [x] eino ADK Agent 模式集成（Supervisor/Deep/PlanExecute）
- [x] AgentFactory 支持多种 Agent 模式

### API 模块
- [x] Agent 配置 HTTP API（CRUD、内置 Agent、主控/子关系）
- [x] Agent Tool 配置 API（工具绑定、MCP/内置/自定义工具）
- [x] Provider 管理 API（对话模型/Rerank/Embedding）
- [x] MCP Server/Tool 管理 API
- [x] 网络搜索配置 API（Tavily/Bing/Google/DuckDuckGo）
- [x] 系统设置 API（键值对配置、系统信息）
- [x] 标签管理 API
- [x] 分块管理 API（CRUD、标签关联）
- [x] 租户管理 API
- [x] API Key 管理

### 内置工具
- [x] knowledge_search - 知识库语义搜索
- [x] web_search - 网络搜索（DuckDuckGo）
- [x] web_fetch - 网页抓取
- [x] grep_chunks - 关键词搜索
- [x] list_knowledge_chunks - 列出分块
- [x] sequential_thinking - 顺序思考
- [x] todo_write - TODO 管理
- [x] data_analysis - DuckDB 数据分析

---

## 待实现功能 🔴

### 高优先级

#### 1. FAQ 知识库
**描述**: 企业问答场景核心功能，支持标准问/相似问/答案管理
**参考**: WeKnora `/internal/handler/faq.go`

功能点:
- [ ] FAQ 条目模型（标准问、相似问法、答案）
- [ ] FAQ 批量导入（支持 dry_run 验证模式）
- [ ] FAQ 列表（分页、标签筛选、关键词搜索）
- [ ] FAQ CRUD
- [ ] FAQ 匹配搜索

#### 2. 评估系统
**描述**: RAG 效果评测，支持数据集管理和评估报告
**参考**: WeKnora `/internal/handler/evaluation.go`

功能点:
- [ ] 评估数据集模型
- [ ] 评估任务创建和执行
- [ ] 评估结果查询
- [ ] 评估指标（准确率、召回率等）

#### 3. 认证系统
**描述**: 生产环境必需，用户注册/登录、JWT、多租户隔离
**参考**: WeKnora `/internal/handler/auth.go`

功能点:
- [ ] 用户模型（用户名、密码哈希、角色）
- [ ] 用户注册/登录
- [ ] JWT Token 生成和验证
- [ ] 认证中间件
- [ ] 密码重置

---

### 中优先级

#### 4. 文件存储
**描述**: MinIO/S3 文件上传和管理
**参考**: WeKnora `/internal/handler/system.go` (MinIO 部分)

功能点:
- [ ] MinIO 客户端集成
- [ ] 文件上传 API
- [ ] 文件下载 API
- [ ] 存储桶管理

#### 5. 自定义 Agent
**描述**: 用户自定义 Agent 配置，支持更灵活的配置
**参考**: WeKnora `/internal/handler/custom_agent.go`

功能点:
- [ ] 自定义 Agent 模型扩展
- [ ] 系统提示词模板
- [ ] 工具组合配置
- [ ] Agent 克隆/导出

#### 6. 更多内置工具

| 工具 | 说明 | 优先级 |
|------|------|--------|
| database_query | 数据库查询工具 | 中 |
| query_knowledge_graph | 知识图谱查询 | 低 |
| transfer_task | Agent 任务转接 | 低 |
| get_document_info | 获取文档信息 | 低 |

---

### 低优先级

#### 7. 消息反馈
- [ ] 点赞/点踩
- [ ] 反馈收集和统计

#### 8. 事件系统
- [ ] 异步任务队列
- [ ] 事件驱动处理
- [ ] 任务进度查询

#### 9. 审计日志
- [ ] API 调用日志
- [ ] 用户操作记录
- [ ] 安全审计

---

## API 统计

| 模块 | API 数量 | 状态 |
|------|----------|------|
| Agent 管理 | 15 | ✅ |
| Provider 管理 | 8 | ✅ |
| MCP Server/Tool | 10 | ✅ |
| 网络搜索配置 | 7 | ✅ |
| 系统设置 | 7 | ✅ |
| 标签管理 | 6 | ✅ |
| 分块管理 | 7 | ✅ |
| 租户管理 | 5 | ✅ |
| API Key 管理 | 5 | ✅ |
| **总计** | **70+** | |

---

## 技术债务

- [ ] `factory.go` 中的 AgentFactory 重复声明问题需要清理
- [ ] 添加单元测试
- [ ] 添加 API 文档（Swagger）
- [ ] 数据库迁移脚本

---

*最后更新: 2026-01-20*
