<div align="center">

# ⚡ superpowers-c456-dashboard

**多项目开发监控面板 · AI 驱动开发的指挥官视图**

扫描多个项目的 **superpowers 扁平 Markdown 文档**（specs / plans / roadmap / sprint），
一键聚合生成可视化 Web 仪表盘——总览 / Roadmap 时间线 / 功能图谱 / 开发计划 / 任务清单。

单二进制 · 跨平台 · 零依赖 · 全局配置 · Web 面板管理

<a href="#sparkles-特性"><b>特性</b></a> ·
<a href="#rocket-快速开始"><b>快速开始</b></a> ·
<a href="#gear-命令行"><b>命令行</b></a> ·
<a href="#books-文档格式约定"><b>文档约定</b></a> ·
<a href="#hammer_and_wrench-构建"><b>构建</b></a>

</div>

---

## ✨ 特性

- **多项目一屏总览**：全局任务/文档统计，各项目进度一目了然
- **8 阶段工作流首页**：客户需求→调研→用户故事→产品设计→功能设计→路线图→开发计划→冲刺 卡片串联，缺文档一目了然
- **五视图深挖**：总览(工作流) · Roadmap 时间线 · 功能设计图谱(React Flow 可拖拽) · 开发计划甘特图 · 任务清单（checkbox 进度）
- **AI 开发顾问**：配置任意 OpenAI 兼容模型 → L3 agent 自动探索项目（读文档/AGENTS.md/数据模型）→ 给出开发最佳实践建议（缺规则/缺文档/模型过期一键提示）
- **详情抽屉**：点任意文档看完整 md 渲染（加粗/表格/代码高亮/文件路径链接 → 源码预览）
- **状态 emoji**：文档标题前默认标 ✅已批准 / 🔶草稿 / 💡提案，一眼识别
- **数据模型目录**：`docs/models/`（ER 图 + 字段表 + 变更记录），写设计同步维护
- **SSE 自动刷新**：文档变化即时重扫，面板实时更新
- **Web 面板管理项目**：一键扫描目录识别 superpowers 项目、添加、移除，全局配置自动读写
- **自发现**：`--scan` 递归扫描目录，自动识别符合 superpowers 格式的项目并入库
- **单二进制可复制分发**：前端 `go:embed` 进二进制，一个文件跨平台即用（macOS / Linux / Windows）

## 🚀 快速开始

### 方式一：直接运行（推荐）

```bash
# 直接运行，无需任何参数——全局配置自动读写
./superpowers-c456-dashboard
# 打开 http://127.0.0.1:8642
```

首次启动自动在全局配置目录创建 `projects.json`（位于 `~/.config/superpowers-c456-dashboard/`，
macOS 为 `~/Library/Application Support/`，Windows 为 `%AppData%\superpowers-c456-dashboard\`）。

### 方式二：一键扫描目录识别项目

```bash
# 递归扫描 ~/Codes，自动识别其中的 superpowers 项目并加入面板
./superpowers-c456-dashboard --scan ~/Codes
```

### 方式三：Web 面板管理

启动后在浏览器打开面板，右上角 **「管理项目」**→ 输入一个目录 → **「扫描」**，
识别出的 superpowers 项目会列出，点「添加」即入面板；也可手动填路径添加、逐项目移除。

## ⚙️ 命令行

| 参数 | 默认 | 说明 |
|------|------|------|
| `--scan <dir>` | — | 递归扫描该目录，自动识别 superpowers 项目并加入配置 |
| `--config <path>` | 全局路径 | 自定义项目清单路径（默认自动读全局配置） |
| `--port` | 8642 | HTTP 端口 |
| `--interval` | 2 | watch 轮询秒数（文档变化自动重扫） |
| `--no-watch` | false | 禁用自动刷新 watch |
| `-h` / `--help` | — | 帮助 |

`--config` 是可选的：默认全自动，无需指定配置文件。

## 📖 文档格式约定

面板为 **superpowers 开发流程**产生的扁平 Markdown 文档做展示。自动识别/分类规则：

- **类型（8 阶段工作流）**：路径含 `requirements` 或标题含「需求」→ requirement；含「调研」→ research；含 `stories` 或「用户故事」→ story；含 `product` 或「产品设计/UX」→ product；含 `specs` 或「设计/方案」→ spec；含 `roadmap/路线图` → roadmap；含「实现计划」→ plan；含 `sprint/冲刺` → sprint；其余 → doc
- **元数据**：标题后 `> 状态：` / `> **键：** 值` 引用块（支持两种键风格）；可追溯字段 `对应需求：` `对应故事：` `对应 Roadmap：`
- **日期**：文件名前缀 `YYYY-MM-DD` 或引用块日期
- **任务/checkbox**：`- [ ]` 未完成、`- [x]` 完成
- **Roadmap 阶段**：`阶段[①②...0-9]` 标题 → 时间线
- **文档状态**：`> 状态：` 取值建议 draft(草稿🔶)/approved(批准✅)/proposal(提案💡)/deprecated(废弃🗑️)

## 🛠️ 构建

```bash
go build -o superpowers-c456-dashboard ./cmd/superpowers-c456-dashboard
# 交叉编译
env GOOS=linux   GOARCH=amd64 CGO_ENABLED=0 go build -o superpowers-c456-dashboard-linux   ./cmd/superpowers-c456-dashboard
env GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -o superpowers-c456-dashboard.exe    ./cmd/superpowers-c456-dashboard
```

> 前端（React + shadcn/ui + Vite + Tailwind）构建需 **bun**：`cd web && bun install && bun run build`，
> 产物自动进 `internal/server/dist` 由 `go:embed` 打进二进制。若只改后端 Go 代码，直接 `go build` 即可。

## 🗂️ 目录结构

```
cmd/superpowers-c456-dashboard/  入口（命令行 + 全局配置 + 扫描入库）
internal/scan/                   superpowers markdown 解析
internal/aggregate/              多项目聚合
internal/config/                 全局配置目录 + 项目扫描识别
internal/ai/                     AI 建议引擎（OpenAI 兼容客户端 + L3 agent loop + 工具集）
internal/server/                 HTTP + SSE + watch + go:embed 前端
  └─ dist/                       前端构建产物（vite 输出，go:embed 打入）
web/                             React 前端源码（bun 构建）
docs/models/                     数据模型目录（ER 图 + 字段表 + changelog）
docs/AGENTS.md.tpl               最佳实践规则集模板（AI 注入用）
```

## 🗃️ 数据契约（GET /data）

```json
{
  "projects": [
    {
      "name": "athena", "root": "/path", "type": "", "status": "",
      "stats": { "total_docs": 12, "tasks_done": 8, "tasks_total": 10, "completion": 80 },
      "documents": [ { "path", "title", "date", "type", "status", "summary", "sections", "tasks", "content" } ],
      "roadmap_stages": [ ... ]
    }
  ],
  "total_projects": 3,
  "global_tasks_total": 30, "global_tasks_done": 20, "global_completion": 66, "global_docs_total": 45
}
```

## 🤝 生态

superpowers-c456-dashboard 是 **c456 生态工具**之一，专为 superpowers 风格 AI 驱动开发流程服务。
欢迎使用、反馈、贡献。

## 📄 License

MIT
