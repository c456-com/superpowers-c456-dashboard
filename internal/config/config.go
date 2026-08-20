// Package config 管理全局配置与项目目录扫描。
// 全局配置目录：os.UserConfigDir()/superpowers-c456-dashboard/（类似 opencode 的 ~/.config/）。
// 统一提供：默认配置路径、加载/保存项目清单（JSON）、递归扫描目录识别 superpowers 项目。
package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"superpowers-c456-dashboard/internal/aggregate"
)

// appName 全局配置目录名（贴合 os.UserConfigDir 下的子目录）。
const appName = "superpowers-c456-dashboard"

// GlobalDir 返回全局配置目录（os.UserConfigDir()/superpowers-c456-dashboard）。
func GlobalDir() string {
	base, err := os.UserConfigDir()
	if err != nil {
		base = "."
	}
	return filepath.Join(base, appName)
}

// GlobalFile 返回全局配置文件路径（projects.json）。
func GlobalFile() string {
	return filepath.Join(GlobalDir(), "projects.json")
}

// EnsureDir 确保全局配置目录存在。
func EnsureDir() error {
	return os.MkdirAll(GlobalDir(), 0o755)
}

// LoadSpecs 从指定路径读取项目清单（JSON）。
// 若文件不存在返回空清单（非错误）。
func LoadSpecs(path string) ([]aggregate.ProjectSpec, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return []aggregate.ProjectSpec{}, nil
	}
	if err != nil {
		return nil, err
	}
	if len(strings.TrimSpace(string(data))) == 0 {
		return []aggregate.ProjectSpec{}, nil
	}
	var cfg struct {
		Projects []aggregate.ProjectSpec `json:"projects"`
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	if cfg.Projects == nil {
		return []aggregate.ProjectSpec{}, nil
	}
	// 路径规范化：~ 展开、相对转为绝对
	for i := range cfg.Projects {
		cfg.Projects[i].Path = normalizePath(cfg.Projects[i].Path)
	}
	return cfg.Projects, nil
}

// MigrateLegacyYAML 若 target 尚未存在、且 legacyFile（当前目录的旧 projects.yaml）存在，
// 则把旧 yaml 项目导入为全局 json 配置（一次性迁移）。
func MigrateLegacyYAML(target, legacyFile string) error {
	if _, err := os.Stat(target); err == nil {
		return nil // 全局配置已存在，不迁移
	}
	if _, err := os.Stat(legacyFile); err != nil {
		return nil // 无旧配置
	}
	cfg, err := aggregate.LoadConfig(legacyFile, nil)
	if err != nil {
		return err
	}
	if err := SaveSpecs(target, cfg.Projects); err != nil {
		return err
	}
	return nil
}

// SaveSpecs 保存项目清单到指定路径（JSON，缩进）。
func SaveSpecs(path string, specs []aggregate.ProjectSpec) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	cfg := struct {
		Projects []aggregate.ProjectSpec `json:"projects"`
	}{Projects: specs}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

// NormalizePath 展开 ~ 并转绝对路径（供 API 输入路径规范化）。
func NormalizePath(p string) string { return normalizePath(p) }

// normalizePath 展开 ~ 并转绝对路径。
func normalizePath(p string) string {
	if strings.HasPrefix(p, "~/") {
		home, _ := os.UserHomeDir()
		p = filepath.Join(home, p[2:])
	}
	if !filepath.IsAbs(p) {
		if abs, err := filepath.Abs(p); err == nil {
			p = abs
		}
	}
	return p
}

// Discovered 单个被扫描识别出的候选项目。
type Discovered struct {
	Path     string `json:"path"`
	Name     string `json:"name"`
	DocCount int    `json:"doc_count"`
}

// ScanForProjects 递归扫描 root 目录，识别「看起来像 superpowers 项目」的子目录。
// 识别特征：目录内存在 *_spec/*_plan/roadmap/sprint 等 superpowers 文档，或 docs/superpowers 结构。
// maxDepth 限制递归深度（防深挖失控）；返回候选列表（已排序）。
// onVisit 可选回调：每访问一个目录时调用（用于前端实时显示扫描进度）。
func ScanForProjects(root string, maxDepth int, onVisit ...func(dir string)) []Discovered {
	rootAbs, err := filepath.Abs(root)
	if err != nil {
		return nil
	}
	var visit func(string)
	if len(onVisit) > 0 && onVisit[0] != nil {
		visit = onVisit[0]
	}
	var out []Discovered
	seen := map[string]bool{}

	filepath.WalkDir(rootAbs, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return nil // 跳过不可读
		}
		if visit != nil {
			visit(p)
		}
		if !d.IsDir() || p == rootAbs {
			return nil
		}
		// 计算深度
		rel, _ := filepath.Rel(rootAbs, p)
		depth := len(strings.Split(rel, string(filepath.Separator)))
		if depth > maxDepth {
			return filepath.SkipDir
		}
		base := d.Name()
		if isIgnoredDir(base) {
			return filepath.SkipDir
		}
		count := countSuperpowersDocs(p)
		if count > 0 {
			// 若某祖先目录已被识别为项目 → 跳过（避免把 docs/superpowers 等里层当独立项目）
			ancestor := p
			underExisting := false
			for {
				parent := filepath.Dir(ancestor)
				if parent == rootAbs {
					break
				}
				if seen[parent] {
					underExisting = true
					break
				}
				ancestor = parent
			}
			if !underExisting {
				name := strings.TrimPrefix(rel, string(filepath.Separator))
				if name == "" {
					name = filepath.Base(rootAbs)
				}
				if !seen[p] {
					seen[p] = true
					out = append(out, Discovered{Path: p, Name: name, DocCount: count})
				}
			} else {
				// 停止深入（不会再有新项目）
				return filepath.SkipDir
			}
		}
		return nil
	})
	sort.Slice(out, func(i, j int) bool {
		return out[i].Name < out[j].Name
	})
	return out
}

// countSuperpowersDocs 统计目录下（浅扫该目录自身）符合 superpowers 类型的 md 文档数。
// 提高命中：docs/superpowers 结构或路径含 specs/plans/roadmap/sprint 或文件名带 spec/plan/roadmap。
func countSuperpowersDocs(dir string) int {
	count := 0
	_ = filepath.WalkDir(dir, func(p string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(strings.ToLower(d.Name()), ".md") {
			return nil
		}
		lower := strings.ToLower(p)
		if strings.Contains(lower, "docs/superpowers") ||
			strings.Contains(lower, "specs") ||
			strings.Contains(lower, "plans") ||
			strings.Contains(lower, "roadmap") ||
			strings.Contains(lower, "sprint") ||
			strings.HasSuffix(strings.ToLower(d.Name()), "_spec.md") ||
			strings.HasSuffix(strings.ToLower(d.Name()), "_plan.md") {
			count++
		}
		if count >= 3 {
			return filepath.SkipAll // 够判定了，提前停
		}
		return nil
	})
	return count
}

// isIgnoredDir 判断是否应跳过的目录。
func isIgnoredDir(name string) bool {
	switch name {
	case ".git", "node_modules", "vendor", "dist", "build", ".worktrees", ".idea", ".vscode", "target", "coverage":
		return true
	}
	return false
}
