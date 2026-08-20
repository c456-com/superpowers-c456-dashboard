// superpowers-c456-dashboard — 多项目开发监控面板（Go 单二进制，c456 生态工具）。
//
// 扫描多个项目的 superpowers 扁平 markdown 文档，生成可视化 Web 仪表盘
// （总览 / Roadmap 时间线 / 功能图谱 / 开发计划 / 任务清单 + 详情抽屉），
// 支持项目切换，SSE 自动刷新。可复制分发：一个二进制跨平台即用。
// 全局配置自动读取（os.UserConfigDir()/superpowers-c456-dashboard），可经 web 面板管理项目。
package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"superpowers-c456-dashboard/internal/aggregate"
	"superpowers-c456-dashboard/internal/config"
	"superpowers-c456-dashboard/internal/server"
)

func main() {
	var (
		port     = flag.Int("port", 8642, "HTTP 端口")
		interval = flag.Int("interval", 2, "watch 轮询秒数")
		noWatch  = flag.Bool("no-watch", false, "禁用自动刷新 watch")
		confPath = flag.String("config", "", "项目清单路径（默认全局配置路径；留空自动读取）")
		scanRoot = flag.String("scan", "", "递归扫描该目录并展示识别出的 superpowers 项目")
	)
	flag.Parse()

	// 默认用全局配置路径（无需指定参数）
	if *confPath == "" {
		*confPath = config.GlobalFile()
	}

	// 一次迁移：若全局配置不存在，但当前目录有旧的 projects.yaml → 自动转成全局 json
	if *confPath == config.GlobalFile() {
		if err := config.MigrateLegacyYAML(*confPath, "projects.yaml"); err != nil {
			fmt.Fprintf(os.Stderr, "迁移旧配置提示: %v\n", err)
		}
	}

	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo})))

	// -scan：扫描根目录，把识别出的 superpowers 项目加入配置，然后启动
	if *scanRoot != "" {
		if err := scanAndAppend(*confPath, *scanRoot); err != nil {
			fmt.Fprintf(os.Stderr, "扫描失败: %v\n", err)
			os.Exit(1)
		}
	}

	srv, err := server.New(server.Config{
		Port:       *port,
		Interval:   *interval,
		NoWatch:    *noWatch,
		ConfigPath: *confPath,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "启动失败: %v\n", err)
		os.Exit(1)
	}
	srv.Watch()
	if err := srv.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "服务错误: %v\n", err)
		os.Exit(1)
	}
}

// aggregateSpec 从扫描到的目录构造一个项目配置项（名称用目录名）。
func aggregateSpec(path, dir string) aggregate.ProjectSpec {
	return aggregate.ProjectSpec{
		Name: dir,
		Path: filepath.Clean(path),
		Type: "superpowers",
	}
}

// scanAndAppend 扫描 root 下识别出的 superpowers 项目，合并进现有配置。
func scanAndAppend(confPath, root string) error {
	existing, err := config.LoadSpecs(confPath)
	if err != nil {
		return err
	}
	existingMap := map[string]bool{}
	for _, s := range existing {
		existingMap[s.Path] = true
	}
	// 递归扫描，深度 4 层
	found := config.ScanForProjects(root, 4)
	added := 0
	for _, d := range found {
		if existingMap[d.Path] {
			continue
		}
		existing = append(existing, aggregateSpec(d.Path, d.Name))
		existingMap[d.Path] = true
		added++
	}
	_ = config.EnsureDir()
	if err := config.SaveSpecs(confPath, existing); err != nil {
		return err
	}
	slog.Info("scan", "root", root, "found", len(found), "added", added)
	return nil
}
