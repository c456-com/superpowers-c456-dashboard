# project-dev-dashboard — 多项目开发监控面板（Go 单二进制）

扫描多个项目的 superpowers 扁平 markdown 文档（specs/plans/roadmap/sprint），
生成可视化 Web 仪表盘：总览 / Roadmap 时间线 / 功能设计图谱 / 开发计划甘特图 / 任务清单，
支持**多项目切换**与 SSE 自动刷新。

AI 驱动开发时，用它随时看每个项目的全貌与细节。

## 为什么 Go 单二进制（辉哥定）

- 一个可执行文件跨平台即用（macOS / Linux / Windows），前端用 `go:embed` 打进二进制
- 零依赖、零构建、`可复制分发`：拷贝一个文件到任何机器就能跑
- 原生支持多项目聚合（`projects.yaml` 清单），不用每项目配一个看板

## 快速开始

```bash
# 1. 建 projects.yaml（可复制 projects.yaml.example）
cp projects.yaml.example projects.yaml   # 改成你的项目路径

# 2. 跑起来
./project-dash --config projects.yaml
# 打开 http://127.0.0.1:8642

# 或直接扫单个项目（无需配置文件）
./project-dash --no-watch -p ~/Codes/athena --port 8642
```

## 参数

| 参数 | 默认 | 说明 |
|------|------|------|
| `--config` | `./projects.yaml` | 项目清单配置文件路径 |
| `-p` | — | 扫描单个项目路径（可多次；指定后忽略 config 里的 path） |
| `--port` | 8642 | HTTP 端口 |
| `--interval` | 2 | watch 轮询秒数（文档变化自动重扫） |
| `--no-watch` | false | 禁用自动刷新 watch |

## projects.yaml

见 `projects.yaml.example`。每项：
- `name` 展示名（可选，默认目录名）
- `path` 项目根目录（必填）
- `type` / `status` 标签（可选）
- `dirs` 额外递归扫描的文档目录（默认 `["docs"]`）

## 构建

```bash
go build -o project-dash ./cmd/project-dash
# 交叉编译
env GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o project-dash-linux ./cmd/project-dash
env GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -o project-dash.exe ./cmd/project-dash
```

## 目录结构

```
cmd/project-dash/     入口（参数解析 + 启动）
internal/scan/        markdown 解析（移植自 Python 版 scan.py）
internal/aggregate/   多项目配置 + 聚合
internal/server/      HTTP + SSE + watch + go:embed 前端
  └─ dist/            前端（index.html + vendored mermaid/marked）
```

## 数据契约（/data）

```json
{
  "projects": [
    {
      "name":"athena","root":"/path","type":"","status":"",
      "generated_at":"2026-08-20 10:00:00",
      "stats": {"total_docs":12,"tasks_done":8,"tasks_total":10,"completion":80,...},
      "documents":[{"path","title","date","type","status","summary","sections","tasks","content"}],
      "roadmap_stages":[...]
    }
  ],
  "total_projects":3,
  "generated_at":"...",
  "global_tasks_total":30,"global_tasks_done":20,"global_completion":66,"global_docs_total":45
}
```

## 原生前端 vs 建库前端

本工具前端是**原生 HTML+JS 单文件**（不引 React 构建链）——因为核心诉求是
「可复制分发 + 零构建」，一个二进制内嵌前端即可，无需 node 环境。
若某日需要复杂交互再升级为 React/shadcn（参考 `xiaohui-tech-stack`）。

## 关联

- 旧 Python 版（单项目）：`~/.hermes/skills/devops/project-dev-dashboard/`
- Go 开发规范：`go-development-best-practices` 技能
