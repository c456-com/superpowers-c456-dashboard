package config

import (
	"os"
	"path/filepath"
	"testing"

	"superpowers-c456-dashboard/internal/aggregate"
)

// ScanForProjects 递归识别 superpowers 项目（specs/plans 文档）；只出项目根，
// 识别到某目录含 superpowers 文档后不再深入其内部（docs/superpowers 等不算独立项目）。
func TestScanForProjects(t *testing.T) {
	root := t.TempDir()
	// 一个真正的 superpowers 项目
	proj := filepath.Join(root, "app-a")
	mkdirAll(t, filepath.Join(proj, "docs", "superpowers", "specs"))
	writeFile(t, filepath.Join(proj, "docs", "superpowers", "specs", "2026-01-01-x.md"), "# spec 设计\n## 目标\n- [ ] a")
	writeFile(t, filepath.Join(proj, "docs", "superpowers", "plans", "2026-01-01-p.md"), "# 实现计划\n- [ ] b")
	// 一个非 superpowers 项目（普通 md，无 specs/plans）
	notProj := filepath.Join(root, "app-b")
	mkdirAll(t, notProj)
	writeFile(t, filepath.Join(notProj, "README.md"), "# 普通项目，无关")
	// 顶层另一项目
	proj2 := filepath.Join(root, "app-c")
	mkdirAll(t, filepath.Join(proj2, "docs", "plans"))
	writeFile(t, filepath.Join(proj2, "docs", "plans", "2026-03-01-p.md"), "# 计划\n- [ ] a")

	found := ScanForProjects(root, 4)
	got := map[string]int{}
	for _, d := range found {
		got[d.Name] = d.DocCount
		t.Logf("found: %s (docs=%d)", d.Name, d.DocCount)
	}
	// app-a 与 app-c 识别；app-b 不识别
	if got["app-a"] < 1 {
		t.Errorf("app-a 应被识别（docs=%d）", got["app-a"])
	}
	if got["app-c"] < 1 {
		t.Errorf("app-c 应被识别（docs=%d）", got["app-c"])
	}
	if _, bad := got["app-b"]; bad {
		t.Errorf("app-b 不应被识别为 superpowers 项目")
	}
	// 不出现 docs/superpowers 等里层目录作为独立项目
	for name := range got {
		if name != "app-a" && name != "app-c" {
			t.Errorf("出现意外的项目项: %s（应只出项目根）", name)
		}
	}
}

// LoadSpecs / SaveSpecs 往返。
func TestLoadSaveSpecs(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "projects.json")
	specs := []aggregate.ProjectSpec{
		{Name: "a", Path: "/tmp/a", Type: "superpowers"},
		{Name: "b", Path: "~/foo", Type: "superpowers"},
	}
	if err := SaveSpecs(path, specs); err != nil {
		t.Fatal(err)
	}
	got, err := LoadSpecs(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("应 2 个项目，得 %d", len(got))
	}
	if got[0].Name != "a" || got[0].Path != "/tmp/a" {
		t.Errorf("a 解析错误: %+v", got[0])
	}
	// ~ 应展开为 HOME
	if got[1].Name != "b" || got[1].Path == "~/foo" {
		t.Errorf("~ 未展开: %+v", got[1])
	}
}

// LoadSpecs 文件不存在时返回空清单（非错误）。
func TestLoadSpecsMissing(t *testing.T) {
	got, err := LoadSpecs(filepath.Join(t.TempDir(), "nope.json"))
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Fatalf("应为空，得 %d", len(got))
	}
}

func writeFile(t *testing.T, p, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func mkdirAll(t *testing.T, p string) {
	t.Helper()
	if err := os.MkdirAll(p, 0o755); err != nil {
		t.Fatal(err)
	}
}
