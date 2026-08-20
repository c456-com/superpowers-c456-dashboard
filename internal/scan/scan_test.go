package scan

import (
	"os"
	"path/filepath"
	"testing"
)

// helper: 在临时目录构造一个 superpowers 风格的文档树。
func makeFixture(t *testing.T) (root string) {
	t.Helper()
	root = t.TempDir()
	write := func(rel, content string) {
		p := filepath.Join(root, rel)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("README.md", "# 项目主文档\n\n说明文字。")
	write("docs/2026-08-15-roadmap.md", `# 开发路线图 (roadmap)

> **状态：** 进行中
> **目的：** 产品演进

### 阶段① 地基
搭建基础能力。
### 阶段② 核心
核心功能。
`)
	write("docs/specs/2026-08-15-search-design.md", `# 搜索功能设计 (spec)

> **状态：** approved
> **目的：** 统一检索

## 架构
## 检索
- [x] 设计
- [ ] 实现
`)
	write("docs/plans/2026-08-16-implementation-plan.md", `# 实施计划 (plan)

> **状态：** 待开始

- [ ] 任务一
- [x] 任务二
`)
	write("docs/2026-08-14-调研报告.md", `# 调研 (research)
内容。
`)
	return root
}

func TestScanTypes(t *testing.T) {
	root := makeFixture(t)
	prj := Scan(root, nil, nil)

	byPath := map[string]string{}
	for _, d := range prj.Documents {
		byPath[d.Path] = d.Type
	}

	cases := map[string]string{
		"README.md":                                    "doc",
		"docs/2026-08-15-roadmap.md":                   "roadmap",
		"docs/specs/2026-08-15-search-design.md":       "spec",
		"docs/plans/2026-08-16-implementation-plan.md": "plan",
		"docs/2026-08-14-调研报告.md":                      "research",
	}
	for path, want := range cases {
		got, ok := byPath[path]
		if !ok {
			t.Errorf("缺少文档: %s", path)
			continue
		}
		if got != want {
			t.Errorf("类型 %s: 得 %q, 期望 %q", path, got, want)
		}
	}
}

func TestScanStatsAndTasks(t *testing.T) {
	root := makeFixture(t)
	prj := Scan(root, nil, nil)

	if prj.Stats.TotalDocs != 5 {
		t.Errorf("文档总数 = %d, 期望 5", prj.Stats.TotalDocs)
	}
	if prj.Stats.TasksTotal != 4 {
		t.Errorf("任务总数 = %d, 期望 4", prj.Stats.TasksTotal)
	}
	if prj.Stats.TasksDone != 2 {
		t.Errorf("已完成任务 = %d, 期望 2", prj.Stats.TasksDone)
	}
	if prj.Stats.Completion != 50 {
		t.Errorf("完成率 = %d, 期望 50", prj.Stats.Completion)
	}
	if prj.Stats.SpecsTotal != 1 || prj.Stats.PlansTotal != 1 || prj.Stats.RoadmapsTotal != 1 {
		t.Errorf("类型统计异常: %+v", prj.Stats.ByType)
	}
}

func TestScanRoadmapStages(t *testing.T) {
	root := makeFixture(t)
	prj := Scan(root, nil, nil)
	if len(prj.Roadmap) != 2 {
		t.Fatalf("roadmap 阶段数 = %d, 期望 2", len(prj.Roadmap))
	}
	if prj.Roadmap[0].ID != "①" {
		t.Errorf("stage[0].id = %q, 期望 ①", prj.Roadmap[0].ID)
	}
	if prj.Roadmap[0].Title != "地基" {
		t.Errorf("stage[0].title = %q, 期望 地基", prj.Roadmap[0].Title)
	}
}

func TestScanMetaAndSummary(t *testing.T) {
	root := makeFixture(t)
	prj := Scan(root, nil, nil)
	for _, d := range prj.Documents {
		if d.Path == "docs/specs/2026-08-15-search-design.md" {
			if d.Status != "approved" {
				t.Errorf("spec status = %q, 期望 approved", d.Status)
			}
			if d.Summary != "统一检索" {
				t.Errorf("spec summary = %q, 期望 统一检索", d.Summary)
			}
			if len(d.Sections) != 2 {
				t.Errorf("spec sections = %d, 期望 2", len(d.Sections))
			}
		}
	}
}

func TestCollectSignature(t *testing.T) {
	root := makeFixture(t)
	sig1 := CollectSignature(root, nil, nil)
	if len(sig1) != 5 {
		t.Errorf("签名文档数 = %d, 期望 5", len(sig1))
	}
	// 追加一个文档 → 签名变化
	p := filepath.Join(root, "docs", "2026-08-20-new.md")
	if err := os.WriteFile(p, []byte("# 新文档\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	sig2 := CollectSignature(root, nil, nil)
	if len(sig2) != 6 {
		t.Errorf("签名文档数(追加后) = %d, 期望 6", len(sig2))
	}
}

func TestIgnoreDirs(t *testing.T) {
	root := makeFixture(t)
	// 放一个 node_modules 里不该被扫的文档
	p := filepath.Join(root, "docs", "node_modules", "x.md")
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte("# x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	prj := Scan(root, nil, nil)
	for _, d := range prj.Documents {
		if d.Path == "docs/node_modules/x.md" {
			t.Errorf("node_modules 里的文档不应被扫描: %s", d.Path)
		}
	}
}
