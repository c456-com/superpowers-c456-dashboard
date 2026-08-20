package aggregate

import (
	"os"
	"path/filepath"
	"testing"
)

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// 构造两个测试项目 + 一个 projects.yaml。
func makeFixture(t *testing.T) (configPath string) {
	t.Helper()
	base := t.TempDir()
	p1 := filepath.Join(base, "projA")
	p2 := filepath.Join(base, "projB")
	mk := func(root string) {
		if err := os.MkdirAll(filepath.Join(root, "docs", "specs"), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	mk(p1)
	mk(p2)
	writeFile(t, filepath.Join(p1, "docs", "specs", "2026-08-01-a-design.md"),
		"# A 设计 (spec)\n\n> **状态：** approved\n\n- [x] 完成\n- [ ] 待做\n")
	writeFile(t, filepath.Join(p2, "docs", "2026-08-02-roadmap.md"),
		"# B 路线 (roadmap)\n\n### 阶段① 起步\n基础。\n")
	cfg := filepath.Join(base, "projects.yaml")
	writeFile(t, cfg, `projects:
  - name: 项目A
    path: `+p1+`
    type: 产品
    status: 开发中
  - name: 项目B
    path: `+p2+`
    type: 工具
`)
	return cfg
}

func TestLoadConfigAndScanAll(t *testing.T) {
	cfgPath := makeFixture(t)
	cfg, err := LoadConfig(cfgPath, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.Projects) != 2 {
		t.Fatalf("projects = %d, 期望 2", len(cfg.Projects))
	}
	if cfg.Projects[0].Name != "项目A" {
		t.Errorf("name0 = %q, 期望 项目A", cfg.Projects[0].Name)
	}
	if cfg.Projects[1].Status != "" {
		t.Errorf("status1 应为空")
	}

	agg := ScanAll(cfg)
	if agg.Total != 2 {
		t.Errorf("total = %d, 期望 2", agg.Total)
	}
	if len(agg.Projects) != 2 {
		t.Errorf("projects len = %d, 期望 2", len(agg.Projects))
	}
	// 项目A 名字用了配置里的展示名
	found := map[string]bool{}
	for _, p := range agg.Projects {
		found[p.Name] = true
	}
	if !found["项目A"] || !found["项目B"] {
		t.Errorf("展示名映射失败: %v", found)
	}
	if agg.GlobalTasks != 2 || agg.GlobalDone != 1 {
		t.Errorf("global tasks = %d/%d, 期望 1/2", agg.GlobalDone, agg.GlobalTasks)
	}
	if agg.GlobalCompletion != 50 {
		t.Errorf("global completion = %d, 期望 50", agg.GlobalCompletion)
	}
}

func TestOverridePaths(t *testing.T) {
	cfgPath := makeFixture(t)
	// 用 override 路径扫描，忽略 config 里的 path
	dir := filepath.Dir(cfgPath)
	projA := filepath.Join(dir, "projA")
	cfg, err := LoadConfig("", []string{projA})
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.Projects) != 1 {
		t.Fatalf("override projects = %d, 期望 1", len(cfg.Projects))
	}
	agg := ScanAll(cfg)
	if agg.Total != 1 {
		t.Errorf("total = %d, 期望 1", agg.Total)
	}
}
