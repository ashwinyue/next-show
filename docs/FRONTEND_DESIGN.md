# Next-Show 前端设计文档

## 1. 设计理念 (Design Philosophy)

基于 `frontend-design` 技能的指导，我们将采用 **"Refined Industrial" (精致工业风)** 作为核心视觉风格。

- **核心关键词**: 精准 (Precision)、原始 (Raw)、功能主义 (Functionalism)。
- **视觉特征**:
  - **色彩**: 默认深色模式 (Dark Mode)。主色调为深灰/黑，强调色采用高饱和度的 **琥珀色 (Amber)** 或 **青色 (Cyan)**，营造一种专业工具的氛围。
  - **排版**: 标题和数据展示采用等宽字体 (Monospaced，如 JetBrains Mono)，正文采用高可读性的无衬线字体 (Sans-serif)。
  - **材质**: 使用微弱的网格背景 (Grid Background)，磨砂玻璃 (Glassmorphism) 用于层级区分，边框清晰锐利，避免过度的圆角。
  - **动效**: 快速、干脆的微交互。加载状态使用技术感强的 Loading 动画（如数据流、光标闪烁）。

## 2. 技术栈 (Tech Stack)

- **构建工具**: Vite
- **框架**: React 18+ (TypeScript)
- **状态管理**: Zustand (轻量级，适合复杂交互)
- **UI 组件库**: Shadcn UI (基于 Radix UI，高可定制)
- **样式**: Tailwind CSS
- **路由**: React Router v6
- **数据请求**: TanStack Query (React Query) + Axios
- **图标**: Lucide React
- **Markdown 渲染**: React Markdown + Syntax Highlighter

## 3. 目录结构 (Directory Structure)

```
web/
├── public/
├── src/
│   ├── api/              # API 接口定义 (对应 internal/handler)
│   ├── assets/           # 静态资源
│   ├── components/       # 组件
│   │   ├── ui/           # Shadcn 基础组件 (Button, Input, etc.)
│   │   ├── common/       # 通用业务组件 (Layout, Sidebar)
│   │   └── features/     # 功能特定组件 (ChatBubble, AgentCard)
│   ├── hooks/            # 自定义 Hooks
│   ├── lib/              # 工具函数 (utils, axios instance)
│   ├── pages/            # 页面组件
│   │   ├── auth/         # 登录/注册
│   │   ├── chat/         # 对话主界面
│   │   ├── agent/        # Agent 管理
│   │   └── knowledge/    # 知识库管理
│   ├── stores/           # Zustand 状态管理
│   ├── types/            # TypeScript 类型定义 (对应 internal/model)
│   ├── App.tsx
│   └── main.tsx
├── index.html
├── package.json
├── tailwind.config.js
└── vite.config.ts
```

## 4. 功能模块设计

### 4.1 布局 (Layout)
- **侧边栏 (Sidebar)**:
  - 顶部: 应用 Logo (Next-Show)。
  - 导航区: "对话 (Chat)", "智能体 (Agents)", "知识库 (Knowledge)", "设置 (Settings)"。
  - 底部: 用户信息、主题切换、折叠按钮。
- **顶栏 (Header)** (部分页面): 显示当前页面标题或面包屑，操作按钮。

### 4.2 核心对话 (Chat Interface)
- **会话列表**: 左侧次级侧边栏，显示历史会话。
- **对话区域**:
  - **消息气泡**: 区分用户 (右侧) 和 Agent (左侧)。
  - **思考过程 (Thinking)**: Agent 的思考过程使用可折叠的 "手风琴" 样式展示，标记为 "Thinking..."，展开可查看详细步骤。
  - **工具调用 (Tool Call)**: 显示工具调用的参数和结果，采用代码块风格。
  - **流式响应**: 兼容 `next-show` 的 SSE 协议，实时打字机效果。
- **输入区域**: 支持多行文本，文件上传按钮 (用于知识库临时挂载)，模型/Agent 切换下拉框。

### 4.3 智能体管理 (Agent Management)
- **列表页**: 卡片式布局展示 Agent。区分 "Orchestrator" (主控) 和 "Specialist" (专家)。
  - 卡片显示: 头像、名称、描述、模型、Type。
- **编辑/创建页**:
  - 表单: 名称、描述、系统提示词 (System Prompt)。
  - 模型选择: 下拉选择 Provider 和 Model。
  - 工具配置: 复选框选择关联的 Tools。
  - 编排配置: (高级) 配置 Agent 之间的调用关系。

### 4.4 知识库 (Knowledge Base)
- **知识库列表**: 表格或列表展示。
- **详情页**:
  - **文档列表**: 显示已上传的文件，解析状态。
  - **上传**: 拖拽上传区域。
  - **切片预览**: 点击文档可查看分块 (Chunk) 详情和向量化结果。
  - **检索测试**: 提供一个搜索框，测试召回效果。

## 5. 交互细节
- **Loading**: 使用骨架屏 (Skeleton) 代替简单的 Spinner。
- **Error Handling**: Toast 提示 (右上角弹出) + 错误边界 (Error Boundary)。
- **Responsive**: 适配桌面端为主，兼顾平板尺寸。

## 6. 接口对接 (API Integration)
- 配置 Vite Proxy 代理 `/api` 请求到 `http://localhost:8080`。
- 实现 Axios Interceptor 处理 JWT Token 自动添加和 401 过期跳转。
