# 子项目① — 8 阶段文档类型扩展 + 数据模型目录（ER 图）

- 日期：2026-08-20
- 状态：已批准（辉哥确认）
- 关联：superpowers-c456-dashboard（c456 生态开源工具）

## 背景与目标

dashboard 当前只能识别 4+1 类文档（spec/plan/roadmap/sprint/research + doc），
缺少完整 8 阶段体系的上游三阶段（customer requirement / user story / product design）。
同时需要一个**数据模型目录**维护 schema 及其演进变更记录，用 ER 图可视化并可切换显示。

本子项目是地基：既不引入 AI，又是后续「L2/L3 agent 理解项目 + 多视图 UX」的前提。

## 一、8 阶段文档类型扩展

### 1. 英文路径约定（辉哥定：磁盘/URL 全英文）

一律用英文目录，组织在 `docs/superpowers/` 下：

| 阶段 | 类型代码 | 英文目录 | 中文标签 |
|------|---------|---------|---------|
| 客户需求 | `requirement` | `requirements/` | 客户需求 |
| 调研 | `research` | `research/` | 调研 |
| 用户故事 | `story` | `stories/` | 用户故事 |
| 产品设计 | `product` | `product/` | 产品设计 |
| 功能设计 | `spec` | `specs/` | 功能设计 |
| 路线图 | `roadmap` | `roadmap/` | 路线图 |
| 开发计划 | `plan` | `plans/` | 开发计划 |
| 冲刺 | `sprint` | `sprints/` | 冲刺 |

> **工作流顺序（首页展示的「开发最佳实践工作流」）**：
> 客户需求 → 调研 → 用户故事 → 产品设计 → 功能设计 → 路线图 → 开发计划 → 冲刺
> 调研放在用户故事之上（辉哥定：需求之后先调研，再形成用户故事）。

### 2. 类型判定扩展（internal/scan/scan.go classify）

新增 3 种类型判定，规则（路径 + 标题关键词）：

- `requirement`：路径含 `/requirements/`，或标题含「需求」「需求规格」「客户需求」
- `story`：路径含 `/stories/`，或标题含「用户故事」「story」「user story」
- `product`：路径含 `/product/`，或标题含「产品设计」「UX」「信息架构」「交互设计」
- 原有 spec/plan/roadmap/sprint/research 判定保留不变
- 判定顺序：roadmap → sprint → requirement → research → story → product → spec → plan → doc
  - requirement/research/story/product 需在 spec 之前（避免「产品设计」被 spec 误判——spec 判定含「设计」）
  - research 判定在 requirement 之后（research 目录明确时优先）

### 3. 前端口径

- `web/src/lib/types.ts`：
  - `TYPE_LABEL` 增加 requirement=客户需求 / story=用户故事 / product=产品设计
  - `DOC_TYPES` 扩展为完整 8 类有序数组
  - 新增 `TYPE_ORDER`（工作流展示顺序：requirement→research→story→product→spec→roadmap→plan→sprint→doc）
- `docs/models/` 配套维护类型枚举与判定规则

### 4. 可追溯性元数据

文档 frontmatter / 引用块支持 `对应需求：` `对应故事：` `对应 Roadmap：` 键
（parseMeta 已支持任意键，规范用法即生效），为后续 AI 跨阶段串联打基础。

### 5. 首页「开发最佳实践工作流」形态（辉哥定）

首页不应只是 8 类文档罗列，而是表达**一条工作流最佳实践**：
客户端需求 → 调研 → 用户故事 → 产品设计 → 功能设计 → 路线图 → 开发计划 → 冲刺。

- 初始（子项目①范围）先在规格层面锁定 `TYPE_ORDER` 这条工作流顺序；
- 后续子项目②（多视图 UX）用「卡片 + 工作流连线」的展示形态实现首页（8 阶段卡片按流程串联，每卡片显示该阶段文档数/可点进）。

### 6. 回归测试

`internal/scan/scan_test.go` 增加用例：每种类型造文件名/标题断言 classify 结果正确；
含边界用例（如「产品设计」不被 spec 误判、`/requirements/` 目录被判定为 requirement）。

## 二、数据模型目录 + ER 图 + 变更记录

### 1. 目录位置（B2：作为项目文档体系一部分，不埋进设置）

数据模型目录放在**项目根 `docs/models/`**，与其他 superpowers 文档并列，被 scan 正常识别，
在 dashboard 左侧手风琴索引/文档流中随手可见、可点开渲染 ER 图——**不藏进设置**（辉哥 UX 判断：数据模型是高查看的开发信息，须随手可见）。

### 2. 目录结构（项目根 docs/models/）

```
docs/models/
  README.md            # 数据模型总览 + 导航
  schema.md            # 当前数据模型（Go struct / JSON 契约 + ER 图）
  changelog.md         # 数据模型变更记录（时间线，倒序）
  er.mmd               # mermaid erDiagram 源文件（可切换显示用）
```

### 3. 多视图（总览 → 蚂蚁，dashboard 文档视图渲染）

- L1 总览：**ER 图**（mermaid erDiagram）——实体关系宏观
- L2 结构：**字段表**（markdown table）——每实体字段名/类型/说明
- L3 明细：**JSON 契约**（/data 返回样例）
- L4 蚂蚁：**类型/枚举明细**（doc.type 8 值 + 判定规则）
- L5 演进：**changelog.md** 变更记录时间线
- 因 dashboard 支持 markdown + mermaid 渲染，这些视图在文档打开时天然可看/可切换（ER 图块 + 字段表相邻）

### 4. 数据模型变更记录（changelog.md）

- 每次数据模型（struct/schema/类型判定规则）变更 → 在 changelog.md 新增一条
- 条目格式（倒序，最新在上）：
  ```
  ## YYYY-MM-DD（类型：add/change/remove）
  - **变更**：描述变更内容
  - **影响**：受影响文件 / 数据契约字段
  - **迁移**：是否需要重建快照 / 兼容处理
  - **决策人**：辉哥 / AI
  ```

### 4. 当前数据模型基线（schema.md 初始内容）

主实体关系：
```
erDiagram
  PROJECT ||--o{ DOCUMENT : contains
  DOCUMENT ||--o{ SECTION : has
  DOCUMENT ||--o{ TASK : has
  PROJECT ||--o{ ROADMAP_STAGE : has
```

字段契约（关键）：
- `Project { name, root, type, status, stats, documents[], roadmap_stages[] }`
- `Document { path, title, date, type, status, summary, sections[], tasks[], content, meta{} }`
  - `type` 枚举：requirement/story/product/spec/roadmap/plan/sprint/research/doc
- `Section { level, title }`
- `Task { text, done }`
- `RoadmapStage { id, title, desc }`

## 三、最佳实践注入（AGENTS.md 规则 / 数据模型同步规范）

辉哥定：我们的最佳实践规则要能被扫描出的项目遵循。dashboard 作为工具，
应让项目 AI 知道写设计时要同步维护最终版数据模型。

### 1. AGENTS.md 规则注入（建议引擎场景之一，落在 AI 子项目）

- dashboard 内置**开发最佳实践规则集**（模板）：8 阶段工作流、数据模型目录规范、
  验证驱动、可追溯性元数据等（参照 athena AGENTS.md 的精炼版）
- 扫描出的项目若 AGENTS.md 缺失这些规则 → **提示**：可一键让 AI 基于
  项目已有 AGENTS.md 附加我们的规则（此能力在子项目 AI(3) 的 L3 agent 实现）

### 2. 数据模型同步规范（写设计时必须维护最终版数据模型）

- 功能设计文档常有「单功能描述 + 数据模型设计」
- 规范：**任何功能设计/spec 涉及数据模型变更时，须同步更新 `docs/models/schema.md`**
  （各设计文档里的模型是过程稿，schema.md 是唯一最终版）
- dashboard/项目扫描到 spec 但 `docs/models/` 缺失或过期 → 提示（AI 子项目实现检测）

> 本子项目①先落地：**规则集模板 + docs/models 目录作为项目规范的一部分**（作为模板/文档存在），
> 注入与检测的动作交给后续 AI 子项目(3)。

## 四、范围

包含：
- scan classify 扩展 3 类型
- types.ts TYPE_LABEL/DOC_TYPES/TYPE_ORDER
- docs/models/ 目录（README/schema/changelog/er.mmd）作为项目最佳实践规范
- 最佳实践规则集模板（AGENTS.md 建议用的规则文档）
- 回归测试

不包含（后续子项目）：
- AI 接入（L3 agent loop / 建议引擎 / 规则自动注入）
- 多视图 UX（工作流首页/图谱/时序/冲刺置顶）
- 7 天趋势曲线

## 五、验证

- `go test ./internal/scan/` 全绿（含新类型用例）
- `bunx tsc --noEmit`
- `bun run build` + 重启，逐步确认 8 类型出现在手风琴/统计
- docs/models 目录齐全、schema.md 的 erDiagram 可渲染
