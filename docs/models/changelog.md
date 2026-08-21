# 数据模型变更记录（Changelog）

> 每次数据模型（struct / schema / 类型判定规则 / 状态映射）变更，在此新增一条（倒序，最新在上）。
> 多视图明细见 `schema.md`，ER 图源文件 `er.mmd`。

## 2026-08-20（add/change：8 阶段文档类型 + 状态归一化）
- **变更**：文档类型从 5 类扩展为 9 类（新增 requirement / story / product）；
  判定顺序改为 roadmap→sprint→requirement→research→story→product→spec→plan→doc；
  前端新增 STATUS_META 状态归一化 emoji 映射 + TYPE_ORDER 工作流顺序。
- **影响**：`internal/scan/scan.go`（classify）、`web/src/lib/types.ts`（TYPE_LABEL/DOC_TYPES/TYPE_ORDER/STATUS_META）、
  `web/src/components/*`（文档列表 status emoji）、`internal/scan/scan_test.go`。
- **迁移**：无需重建快照；新增类型判定向后兼容（旧文档类型不变），可直接重扫。
- **决策人**：辉哥

## 2026-08-20（add：数据模型目录）
- **变更**：建立 `docs/models/` 数据模型目录（schema.md + er.mmd + changelog.md + README），
  作为项目最佳实践规范的一部分（写设计须同步维护最终版 schema.md）。
- **影响**：`docs/models/*`。
- **迁移**：无。
- **决策人**：辉哥
