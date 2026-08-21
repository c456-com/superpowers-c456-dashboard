# 子项目③ — AI 建议引擎（L3 agent loop）+ 文档状态识别

- 日期：2026-08-20
- 状态：草稿（待辉哥批准）
- 关联：superpowers-c456-dashboard（依赖子项目①的 8 阶段类型 + docs/models）

## 背景

辉哥定（Q2=a 手动+文档变化自动；Q3=L3 多步 agent 循环；A=设计先行）。
dashboard 内置 AI 能力：配置模型后，能理解项目 → 给建议。
本规格把三块建议动作**具体化**，并纳入**文档状态 emoji 识别**。

## 一、AI 模型配置（设置里）

- 设置新增「AI 配置」：API Base URL + 模型名 + API Key
- Key 存本地 `auth.json`（`os.UserConfigDir()/superpowers-c456-dashboard/auth.json`，权限 600，`.gitignore` 排除）
- 未配置 → 面板 AI 功能提示「请先配置 AI 模型」

## 二、触发方式（Q2 = a + b）

| 触发 | 说明 |
|------|------|
| a. 手动触发 | 面板点「AI 分析」按钮 → 对当前项目跑一轮 agent |
| b. 文档变化自动 | 复用 watch：检测到项目 superpowers 文档增/改/删 → 自动触发一轮（**30s 去抖**合并，防连续保存狂打 API）|

定时触发（c）留后续。

## 三、L3 agent loop（Q3）

后端实现轻量多步 agent：
- **循环**：模型思考 → 调用内置工具 → 拿到结果 → 再思考 → …直到产出建议
- **内置工具集**（function calling / MCP 风格）：
  - `list_docs(project, type)` — 列出某类型文档
  - `read_doc(project, path)` — 读文档内容/frontmatter
  - `read_agents(project)` — 读 AGENTS.md
  - `check_models(project)` — 检查 docs/models 是否存在/过期
  - `has_dir(project, dir)` — 检查目录存在
- 模型用这些工具自主探索，最终产出**结构化建议**（JSON：action 类型 + 说明 + 建议文案 + 建议执行）

## 四、三块建议动作（具体化）

### 1. AGENTS.md 规则注入
- **检测**：项目无 AGENTS.md，或 AGENTS.md 缺少关键规则标记（如 `.superpowers-c456-workflow` 标记、`docs/models` 规范字样）
- **内置规则集**（dashboard 自带模板 `AGENTS.md.tpl`）：8 阶段工作流约定、数据模型同步规范（写设计须同步 schema.md）、验证驱动、可追溯性元数据
- **建议**：「项目缺少开发最佳实践规则」→ 用户点「一键注入」→ agent 读现有 AGENTS.md + 模板 → 生成/追加规则 → 写回
- **成果标记**：注入后 AGENTS.md 加 `.superpowers-c456-workflow` 标记，避免重复提示

### 2. 缺文档提示
- **检测**：按项目当前阶段判断某类型 0 文档 → 提示创建。优先级参考工作流顺序：
  - 有 product 无 requirement → 提示「缺客户需求来源」
  - 有 spec 无 roadmap/plan → 提示「缺路线图/开发计划」
  - 有 plan 无 sprint → 提示「缺冲刺拆解」
- **建议**：「建议创建《xxx》文档」→ agent 用模板生成骨架（frontmatter + 大纲）→ 写回

### 3. 数据模型缺失/过期提示
- **检测**：
  - 有 spec 但无 `docs/models/` → 提示「缺数据模型目录」
  - `schema.md` 的 `updated` / git 变更时间 早于最近 spec 变更 → 提示「数据模型可能过期，请同步」
- **建议**：「同步数据模型」→ agent 读相关 spec 的数据模型设计 → 合并/更新 schema.md 与 er.mmd

## 五、建议呈现 UX

- 建议以**通知条/卡片**呈现（dashboard 页面底部或右上角浮层），每条含：类型图标 + 说明 + 「执行」按钮（一键让 agent 执行）
- 建议按项目聚合，可一键全部忽略/标记已读
- 手动触发在「AI 分析」按钮；自动触发在有建议时浮出

## 六、文档状态识别 + emoji 前缀（辉哥定）

列表项识别文档状态，标题前加对应 emoji，快速识别草稿/已批准等。

### 状态归一化映射（scan parseMeta 已解析 `> 状态：`，前端做归一化）

| 规范状态 | 匹配（取原文，大小写/中英归一） | emoji | 色标 |
|---------|----------------------------|-------|------|
| 已批准 | approved / APPROVED / 正式方案 / 已验证 / 已定稿 | ✅ | green |
| 草稿 | draft / DRAFT / appended草案 / 未验证 / 未定稿 | 🔶 | amber |
| 提案 | 提案 / proposal / 建议 | 💡 | blue |
| 已废弃 | 废弃 / deprecated / 已删除 | 🗑️ | gray |
| 未知/其他 | 其余 | 📄 | neutral |

### 实现
- 前端新建 `STATUS_META`（映射：匹配 → {emoji, label, color}）
- 文档卡片/列表标题前渲染 emoji + label；原 status 文本可保留为 tooltip
- 归一化在**前端**（`web/src/lib/types.ts` + 卡片组件），后端 `Document.status` 保持原文

## 七、范围

包含：AI 配置、L3 agent loop（工具集）、三块建议动作、建议 UX、状态 emoji 识别。
依赖：子项目①（8 阶段类型 + docs/models）+ ②（多视图 UX 的列表）。
不包含：定时触发(c)、趋势曲线（子项目④）。

## 八、验证

- go 侧：agent 工具集单测（假模型/假工具）
- 前端：`bunx tsc --noEmit` + build + Playwright（配置 AI / 手动触发 / 建议条 / 状态 emoji 渲染）
- 状态映射：各状态值进入对应 emoji/label 的单元测试
