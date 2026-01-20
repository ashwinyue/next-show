# Agent 模板使用指南

## 概述

Agent 模板系统让用户可以快速创建预配置的 Agent，无需手动配置所有参数。系统提供以下模板：

### 主控 Agent（Orchestrator）

| 模板代码 | 名称 | 图标 | 说明 |
|---------|------|------|------|
| `supervisor` | 主控协调器 | 🎯 | 协调多个子 Agent 协作完成任务 |
| `deep` | 深度思考 | 🧠 | 通过深度思考和推理解决复杂问题 |
| `plan_execute` | 计划执行 | 📋 | 先制定计划，然后执行并根据情况调整 |
| `sequential` | 顺序执行 | ➡️ | 按固定顺序依次执行子 Agent |
| `loop` | 循环执行 | 🔁 | 循环执行子 Agent 直到任务完成 |

### 专家 Agent（Specialist）

| 模板代码 | 名称 | 图标 | 说明 |
|---------|------|------|------|
| `chat` | 对话助手 | 💬 | 基础对话，适合简单问答 |
| `rag` | 知识检索 | 📚 | 基于知识库的检索增强生成 |
| `data_analyst` | 数据分析 | 📊 | 使用 DuckDB 进行数据分析 |
| `react` | 反应式 | ⚡ | 根据情况动态决策和行动 |

---

## API 使用

### 1. 列出所有模板

```bash
GET /api/v1/agent-templates
```

**响应示例：**
```json
{
  "templates": [
    {
      "code": "supervisor",
      "name": "supervisor",
      "display_name": "主控协调器",
      "description": "协调多个子 Agent 协作完成任务",
      "agent_type": "supervisor",
      "agent_role": "orchestrator",
      "category": "orchestrator",
      "config": {
        "default_prompt": "你是一个智能任务协调器...",
        "default_iterations": 20
      },
      "icon": "🎯"
    }
  ],
  "total": 9
}
```

---

### 2. 从模板创建 Agent

```bash
POST /api/v1/agent-templates/create
```

**请求示例 1：创建主控协调器**

```json
{
  "template_code": "supervisor",
  "name": "my_supervisor",
  "display_name": "我的任务协调器",
  "provider_id": "your-provider-id",
  "model_name": "gpt-4",
  "sub_agent_ids": [
    "agent-rag-id",
    "agent-data-analyst-id",
    "agent-custom-id"
  ],
  "system_prompt": "你是我的个人助手，可以调用知识检索和数据分析工具",
  "max_iterations": 30
}
```

**请求示例 2：创建知识检索 Agent**

```json
{
  "template_code": "rag",
  "name": "my_knowledge_assistant",
  "display_name": "文档助手",
  "provider_id": "your-provider-id",
  "model_name": "gpt-4"
}
```

**请求示例 3：创建循环执行 Agent**

```json
{
  "template_code": "loop",
  "name": "data_processing_loop",
  "display_name": "数据处理循环",
  "provider_id": "your-provider-id",
  "model_name": "gpt-4",
  "sub_agent_ids": [
    "agent-collector-id",
    "agent-cleaner-id",
    "agent-saver-id"
  ],
  "max_iterations": 5
}
```

**响应示例：**
```json
{
  "id": "uuid",
  "name": "my_supervisor",
  "display_name": "我的任务协调器",
  "description": "协调多个子 Agent 协作完成任务",
  "agent_type": "supervisor",
  "agent_role": "orchestrator",
  "is_enabled": true,
  "created_at": "2026-01-20T10:00:00Z"
}
```

---

## 使用流程

### 场景 1：创建带子 Agent 的主控协调器

**步骤 1：创建子 Agent**

```bash
# 创建知识检索 Agent
POST /api/v1/agent-templates/create
{
  "template_code": "rag",
  "name": "knowledge_helper",
  "display_name": "知识助手",
  "provider_id": "provider-1",
  "model_name": "gpt-4"
}

# 创建数据分析 Agent
POST /api/v1/agent-templates/create
{
  "template_code": "data_analyst",
  "name": "data_expert",
  "display_name": "数据专家",
  "provider_id": "provider-1",
  "model_name": "gpt-4"
}
```

**步骤 2：创建主控协调器并关联子 Agent**

```bash
POST /api/v1/agent-templates/create
{
  "template_code": "supervisor",
  "name": "my_supervisor",
  "display_name": "智能协调器",
  "provider_id": "provider-1",
  "model_name": "gpt-4",
  "sub_agent_ids": [
    "knowledge-helper-id",
    "data-expert-id"
  ]
}
```

**步骤 3：开始对话**

```bash
# 创建会话
POST /api/v1/sessions
{
  "agent_id": "my-supervisor-id"
}

# 流式对话
POST /api/v1/chat/stream
{
  "session_id": "session-id",
  "message": "帮我分析一下最近的销售数据，并查询产品文档中的相关信息"
}
```

---

### 场景 2：创建数据处理循环

**步骤 1：创建流程中的各个 Agent**

```bash
# 数据采集 Agent
POST /api/v1/agents
{
  "name": "data_collector",
  "display_name": "数据采集器",
  "agent_type": "chat",
  "system_prompt": "你负责从 API 采集数据",
  ...
}

# 数据清洗 Agent
POST /api/v1/agents
{
  "name": "data_cleaner",
  "display_name": "数据清洗器",
  "agent_type": "chat",
  "system_prompt": "你负责清洗和格式化数据",
  ...
}

# 数据存储 Agent
POST /api/v1/agents
{
  "name": "data_saver",
  "display_name": "数据存储器",
  "agent_type": "chat",
  "system_prompt": "你负责将数据保存到数据库",
  ...
}
```

**步骤 2：创建循环 Agent**

```bash
POST /api/v1/agent-templates/create
{
  "template_code": "loop",
  "name": "data_pipeline",
  "display_name": "数据处理管道",
  "provider_id": "provider-1",
  "model_name": "gpt-4",
  "sub_agent_ids": [
    "data-collector-id",
    "data-cleaner-id",
    "data-saver-id"
  ],
  "max_iterations": 10
}
```

---

## 参数说明

### CreateAgentFromTemplateRequest

| 字段 | 类型 | 必填 | 说明 |
|-----|------|------|------|
| `template_code` | string | ✅ | 模板代码（如 `supervisor`, `rag`） |
| `name` | string | ✅ | Agent 名称（唯一标识） |
| `display_name` | string | ✅ | Agent 显示名称 |
| `provider_id` | string | ✅ | 模型提供商 ID |
| `model_name` | string | ✅ | 模型名称（如 `gpt-4`） |
| `system_prompt` | string | ❌ | 自定义系统提示词（覆盖模板默认值） |
| `sub_agent_ids` | []string | ❌ | 子 Agent ID 列表 |
| `max_iterations` | int | ❌ | 最大迭代次数（覆盖模板默认值） |

---

## 模板配置覆盖

### 1. 系统提示词

如果未提供 `system_prompt`，使用模板的 `default_prompt`。

```json
{
  "template_code": "chat",
  "system_prompt": "你是一个专业的客服助手"  // 覆盖模板默认值
}
```

### 2. 最大迭代次数

如果未提供 `max_iterations`，使用模板的 `default_iterations`。

```json
{
  "template_code": "supervisor",
  "max_iterations": 50  // 覆盖默认值 20
}
```

### 3. 子 Agent 配置

通过 `sub_agent_ids` 字段关联子 Agent：

```json
{
  "template_code": "supervisor",
  "sub_agent_ids": [
    "rag-agent-id",    // 内置 Agent
    "custom-agent-id"  // 用户自定义 Agent
  ]
}
```

---

## 最佳实践

### 1. 主控 Agent 选择

| 需求 | 推荐模板 |
|-----|---------|
| 需要灵活协调多个专家 | `supervisor` |
| 需要深度推理和思考 | `deep` |
| 需要制定和执行计划 | `plan_execute` |
| 固定流程执行 | `sequential` |
| 需要重复执行流程 | `loop` |

### 2. 子 Agent 组合

**示例组合 1：研究助手**
```
主控: supervisor
├─ rag (文档检索)
├─ data_analyst (数据分析)
└─ chat (总结输出)
```

**示例组合 2：客服系统**
```
主控: supervisor
├─ rag (知识库查询)
├─ chat (常规对话)
└─ custom (订单查询)
```

**示例组合 3：数据管道**
```
主控: loop (循环 5 次)
├─ collector (数据采集)
├─ cleaner (数据清洗)
└─ saver (数据存储)
```

---

## 常见问题

### Q: 可以修改模板创建的 Agent 吗？

A: 可以。通过 `PUT /api/v1/agents/:id` 更新 Agent 配置。

### Q: 如何查看 Agent 的子 Agent 关系？

A: 使用 `GET /api/v1/agents/:id/relations`。

### Q: 子 Agent 的执行顺序是如何确定的？

A: 通过 `AgentRelation` 表的 `sort_order` 字段控制。

### Q: 可以嵌套主控 Agent 吗？

A: 可以。子 Agent 本身也可以是主控 Agent，形成多层结构。

---

**🎉 开始使用 Agent 模板，快速构建您的多 Agent 系统！**
