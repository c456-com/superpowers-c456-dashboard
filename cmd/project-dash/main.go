// project-dash — 多项目开发监控面板（Go 单二进制）。
//
// 扫描多个项目的 superpowers 扁平 markdown 文档，生成可视化 Web 仪表盘
// （总览 / Roadmap 时间线 / 功能图谱 / 开发计划 / 任务清单 + 详情抽屉），
// 支持项目切换，SSE 自动刷新。可复制分发：一个二进制跨平台即用。
package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"project-dev-dashboard/internal/server"
)

func main() {
	var (
		port     = flag.Int("port", 8642, "HTTP 端口")
		interval = flag.Int("interval", 2, "watch 轮询秒数")
		noWatch  = flag.Bool("no-watch", false, "禁用自动刷新 watch")
		confPath = flag.String("config", "", "projects.yaml 路径（默认 ./projects.yaml）")
		paths    multiFlag
	)
	flag.Var(&paths, "p", "扫描的单个项目路径（可多次；指定后忽略 config 里的 path）")
	flag.Parse()

	if *confPath == "" {
		*confPath = "projects.yaml"
	}
	// 相对 config 路径基于 cwd
	if !filepath.IsAbs(*confPath) {
		if abs, err := filepath.Abs(*confPath); err == nil {
			*confPath = abs
		}
	}
	// 若给定 -p，则无需 projects.yaml 存在；否则必须存在
	if len(paths) == 0 {
		if _, err := os.Stat(*confPath); err != nil {
			fmt.Fprintf(os.Stderr, "未找到配置文件 %s\n", *confPath)
			fmt.Fprintln(os.Stderr, "用法示例:")
			fmt.Fprintln(os.Stderr, "  1) projects.yaml 配置多项目（推荐）:")
			fmt.Fprintln(os.Stderr, "     project-dash --config /path/projects.yaml")
			fmt.Fprintln(os.Stderr, "  2) 直接扫单个项目:")
			fmt.Fprintln(os.Stderr, "     project-dash --no-watch -p /path/to/project [--port 8642]")
			os.Exit(1)
		}
	}

	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo})))

	srv, err := server.New(server.Config{
		Port:       *port,
		Interval:   *interval,
		NoWatch:    *noWatch,
		ConfigPath: *confPath,
		Overrides:  []string(paths),
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

// multiFlag 支持 -p 多次指定。
type multiFlag []string

func (m *multiFlag) String() string { return fmt.Sprint([]string(*m)) }
func (m *multiFlag) Set(v string) error {
	abs, err := filepath.Abs(v)
	if err != nil {
		return err
	}
	*m = append(*m, abs)
	return nil
}
