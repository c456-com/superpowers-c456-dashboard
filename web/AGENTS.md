# AGENTS.md — superpowers-c456-dashboard web 前端（shadcn 铁律）

> 本文件约束本仓库前端对 AI 代理的硬性规范。**只许用 shadcn 生态组件 + theme 主题样式，禁止自由发挥原生组件/硬编码颜色。**

## 🚫 一、禁自由发挥原生组件（辉哥定，铁律）

- **禁止** 手写原生组件去实现 shadcn 生态已有的东西（Dialog/Sheet/Select/Accordion/Button/Badge/Input/Tabs 等一律用 `web/src/components/ui/` 里的 shadcn 组件）
- **禁止** 用原生 `<select>`、自造 modal、手写弹层
- **只有 shadcn 生态确实没有合适组件时**才允许自己写，且写成符合 shadcn 风格的组件（CSS 变量驱动、可放 ui/ 目录复用）
- 新增 UI 前先查 `web/src/components/ui/` + shadcn blocks（`bunx --bun shadcn@latest add <组件>`）
- 图标用 `lucide-react`，禁 emoji 装饰、禁手写 SVG 图标（lucide 已有）

## 🎨 二、样式必须走 shadcn theme（CSS 变量），禁止硬编码颜色

- **文字色**：正文 `text-foreground`、次要 `text-muted-foreground`、强调 `text-primary` —— **禁 `text-gray-400/500/600/700`、`text-black`、`text-[#...]`**
- **背景**：卡片 `bg-card`、面板底 `bg-background`、hover `bg-muted` / `hover:bg-accent` —— **禁 `bg-white`、`bg-gray-50/100`、`bg-[#f7f7f8]`**
- **边框**：`border-border` —— 禁 `border-gray-200`
- 所有颜色引用 theme 变量（`--background`/`--foreground`/`--card`/`--muted`/`--primary` 等），词法在 `web/src/index.css` 的 `:root` / `.dark`
- **markdown/代码块文字看不到 = 用了硬编码浅灰**（`text-muted/0.922`、`gray-*`）——必须 `text-foreground`（深色），代码块 `bg-muted/60 text-foreground border-border`
- 唯一例外：**类型语义色**（doc type Tag：rose/teal/fuchsia/amber 等标识某阶段类型）是有意区分，保留，但卡片底色仍用 `bg-card`，Tag 底色用 `bg-*\/10`（透明）而非 `bg-*-50`

## 🌗 三、明暗切换（dark mode）

- index.css 已有 `:root`（亮）+ `.dark`（暗）CSS 变量
- 切换方式：给 `<html>` 加/去 `class="dark"`
- **面板必须有明暗切换入口**（Header 右上角 sun/moon 按钮，`lucide-react` 的 `Sun`/`Moon`）
- 初始值跟随 `window.matchMedia('(prefers-color-scheme: dark)')`，切换后存 `localStorage('theme')`，下次启动恢复
- 组件里 `bg-card`/`text-foreground` 等会自动适配明暗；任何用 `bg-white`/`text-gray-*` 的都会在暗色下错乱

## ⚙️ 构建

- **bun 禁 npm**：依赖/构建用 `bun` / `bunx`
- 前端 build：`cd web && bun run build`（**必须单独跑，勿串 &&**，否则 bun 打 help 假成功）；产物自动进 `internal/server/dist` 由 go:embed 打入
- 改前端 → `bun run build` → `go build -o superpowers-c456-dashboard ./cmd/superpowers-c456-dashboard` → 重启，三步缺一不可
- 验证：`bun run typecheck` + `bun run build` + Playwright 端到端；检查 UI 有无 `bg-white`/`text-gray-` 残留

## 📐 布局规范（辉哥定）

- 首页＝多项目看板（无左侧导航）；进单项目＝左侧 shadcn Accordion 索引 + 顶部视图页签
- 项目切换下拉用 shadcn `Select`（禁原生 select）；点击交互必须同步 hash 路由
- modal 容器两段式：toolbar（shrink-0 px-6 py-4 border-b）+ content（flex-1 overflow-y-auto 独立 padding）
- 详情抽屉 DocDrawer：react-markdown + rehype-highlight，markdown 文字 `text-foreground`