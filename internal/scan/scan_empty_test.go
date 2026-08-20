package scan

import (
	"os"
	"path/filepath"
	"testing"
)

// 回归：项目无 roadmap 文档（或扫描不到 stage）时，roadmap_stages / documents 必须是空数组而非 nil。
// 曾因 `var stages []RoadmapStage` 返回 nil → JSON 序列化为 null → 前端 .length 崩溃白屏。
func TestNoNilSlicesWhenEmpty(t *testing.T) {
	root := t.TempDir()
	// 只放一个普通 md，无 roadmap、无 spec 任务
	p := filepath.Join(root, "README.md")
	if err := os.WriteFile(p, []byte("# 项目说明\n\n普通文档，无 roadmap。\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	prj := Scan(root, nil, nil)

	if prj.Documents == nil {
		t.Error("Documents 必须是非 nil 空数组，不能是 null")
	}
	if prj.Roadmap == nil {
		t.Error("Roadmap 必须是非 nil 空数组，不能是 null")
	}
	if len(prj.Roadmap) != 0 {
		t.Errorf("Roadmap 应为空，得 %d", len(prj.Roadmap))
	}
	// 文档内的 tasks/sections 也必须非 nil
	for _, d := range prj.Documents {
		if d.Tasks == nil {
			t.Errorf("文档 %s 的 tasks 是 nil（应为空数组）", d.Path)
		}
		if d.Sections == nil {
			t.Errorf("文档 %s 的 sections 是 nil（应为空数组）", d.Path)
		}
	}
}

// 回归：有 roadmap 文档但无「阶段X」小节也无箭头序列时，roadmap_stages 应为空数组而非 null
func TestRoadmapEmptyNotNil(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "docs")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	// 一个 roadmap 类型文档但没有可解析的阶段
	if err := os.WriteFile(filepath.Join(dir, "2026-08-20-roadmap.md"),
		[]byte("# 开发路线 (roadmap)\n\n这是路线说明，没有阶段小节。\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	prj := Scan(root, nil, nil)
	if prj.Roadmap == nil {
		t.Fatal("Roadmap 是 nil")
	}
	if len(prj.Roadmap) != 0 {
		t.Errorf("Roadmap 应为 0 个阶段，得 %d", len(prj.Roadmap))
	}
}
