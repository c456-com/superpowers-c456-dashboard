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
| 用户故事 | `story` | `stories/` | 用户故事 |
| 产品设计 | `product` | `product/` | 产品设计 |
| 功能设计 | `spec` | `specs/` | 功能设计 |
| 路线图 | `roadmap` | `roadmap/` | 路线图 |
| 开发计划 | `plan` | `plans/` | 开发计划 |
| 冲刺 | `sprint` | `sprints/` | 冲刺 |
| 调研 | `research` | `research/` | 调研 |

### 2. 类型判定扩展（internal/scan/scan.go classify）

新增 3 种类型判定，规则（路径 + 标题关键词）：

- `requirement`：路径含 `/requirements/`，或标题含「需求」「需求规格」「客户需求」
- `story`：路径含 `/stories/`，或标题含「用户故事」「story」「user story」
- `product`：路径含 `/product/`，或标题含「产品设计」「UX」「信息架构」「交互设计」
- 原有 spec/plan/roadmap/sprint/research 判定保留不变
- 判定顺序：roadmap → sprint → requirement → story → product → spec → plan → research → doc
  （requirement/story/product 需在 spec 之前，避免「产品设计」被 spec 误判——因为 spec 判定含「设计」）

### 3. 前端口径

- `web/src/lib/types.ts`：
  - `TYPE_LABEL` 增加 requirement=客户需求 / story=用户故事 / product=产品设计
  - `DOC_TYPES` 扩展为完整 8 类有序数组
  - 新增 `TYPE_ORDER`（8 阶段展示顺序：requirement→story→product→spec→roadmap→plan→sprint→research→doc）
- `docs/models/` 配套维护类型枚举与判定规则

### 4. 可追溯性元数据

文档 frontmatter / 引用块支持 `对应需求：` `对应故事：` `对应 Roadmap：` 键
（parseMeta 已支持任意键，规范用法即生效），为后续 AI 跨阶段串联打基础。

### 5. 回归测试

`internal/scan/scan_test.go` 增加用例：每种类型造文件名/标题断言 classify 结果正确；
含边界用例（如「产品设计」不被 spec 误判、`/requirements/` 目录被判定为 requirement）。

## 二、数据模型目录 + ER 图 + 变更记录

### 1. 目录结构（仓库根 docs/models/）

```
docs/models/
  README.md            # 数据模型总览 + 导航
  schema.md            # 当前数据模型（Go struct / JSON 契约 + ER 图）
  changelog.md         # 数据模型变更记录（时间线，倒序）
  er.mmd               # mermaid erDiagram 源文件（可切换显示用）
```

### 2. ER 图（ER 图切换显示）

- `schema.md` 内嵌 **mermaid `erDiagram`** 展示实体关系：Project / Document / Section / Task / RoadmapStage
- 前端详情/文档渲染已支持 mermaid（react-markdown + 代码块），ER 图在 md 预览中可见
- 可切换显示：同一实体关系提供 **ER 图（mermaid）+ 字段表（markdown table）** 两种呈现，
  用户可切换（dashboard 文档视图支持 markdown 渲染即天然可切）
- `er.mmd` 独立文件便于外部工具/渲染

### 3. 数据模型变更记录（changelog.md）

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

## 三、范围

包含：
- scan classify 扩展 3 类型
- types.ts TYPE_LABEL/DOC_TYPES/TYPE_ORDER
- docs/models/ 目录（README/schema/changelog/er.mmd）
- 回归测试

不包含（后续子项目）：
- AI 接入（L3 agent loop / 建议引擎）
- 多视图 UX（路线图/图谱/时序/冲刺置顶）
- 7 天趋势曲线

## 四、验证

- `go test ./internal/scan/` 全绿（含新类型用例）
- `bunx tsc --noEmit`
- `bun run build` + 重启，逐步确认 8 类型出现在手风琴/统计
- docs/models 目录齐全、schema.md 的 erDiagram 可渲染
