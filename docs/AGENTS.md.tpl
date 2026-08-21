# AGENTS.md — 开发最佳实践（superpowers-c456-dashboard 规则模板）

> 将此规则注入被扫描项目的 AGENTS.md，让项目 AI / 开发者遵循统一开发最佳实践。
> 注入后请保留标记 `.superpowers-c456-workflow` 避免重复提示。

## 开发工作流（8 阶段，遵循此顺序产出文档）

客户需求 → 调研 → 用户故事 → 产品设计 → 功能设计 → 路线图 → 开发计划 → 冲刺

每阶段讨论产生多份文档，存放于 `docs/superpowers/` 对应英文目录：
`requirements/` · `research/` · `stories/` · `product/` · `specs/` · `roadmap/` · `plans/` · `sprints/`
文件名用日期前缀 `YYYY-MM-DD-<topic>.md`。

## 数据模型同步规范（强制）

- **任何功能设计/spec 涉及数据模型变更时，须同步更新 `docs/models/schema.md`**
  （各设计文档里的模型是过程稿，schema.md 是唯一最终版）
- 变更后在 `docs/models/changelog.md` 记录一条（变更/影响/迁移/决策人）
- ER 图用 mermaid `erDiagram` 维护在 `docs/models/er.mmd`

## 验证驱动

- 方案先存「未验证」状态，数据验证合格后转「正式/已批准」
- 找不到内容不乱返回（相关性阈值）
- 数据本地处理，零上传（客户保密数据）

## 可追溯性元数据

文档 frontmatter / 引用块使用 `对应需求：` `对应故事：` `对应 Roadmap：`
等键关联上下游（需求→用户故事→设计→路线图），保证可追溯。

## 状态标记

文档状态写在 `> 状态：` 引用块，取值建议：`draft`（草稿）/ `approved`（已批准）/ `proposal`（提案）/ `deprecated`（已废弃）。
<!-- markdown: 中间由 AI 依据项目实际 AGENTS.md 合并附加 -->
