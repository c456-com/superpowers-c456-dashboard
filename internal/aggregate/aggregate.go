// Package aggregate 管理多项目配置与聚合：读取 projects.yaml，扫描所有项目，
// 生成聚合后的总览数据（所有项目 + 各自的完整数据）。
package aggregate

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	"project-dev-dashboard/internal/scan"
)

// Config 顶层配置。
type Config struct {
	Projects []ProjectSpec `yaml:"projects" json:"projects"`
}

// ProjectSpec 单个项目的配置项。
type ProjectSpec struct {
	// Name 展示名（可选，默认用目录名）
	Name string `yaml:"name" json:"name"`
	// Path 项目根目录（必填）
	Path string `yaml:"path" json:"path"`
	// Dirs 额外递归扫描的文档目录（默认 docs）
	Dirs []string `yaml:"dirs,omitempty" json:"dirs,omitempty"`
	// Exclude 额外忽略的目录
	Exclude []string `yaml:"exclude,omitempty" json:"exclude,omitempty"`
	// Status 项目状态（开发中/已上线/暂停 等，前端标签）
	Status string `yaml:"status,omitempty" json:"status,omitempty"`
	// Type 项目类型（产品/web服务/工具 等）
	Type string `yaml:"type,omitempty" json:"type,omitempty"`
}

// ProjectEntry 聚合后单个项目（扫描结果 + 配置信息）。
type ProjectEntry struct {
	scan.Project
	Status string `json:"status,omitempty"`
	Type   string `json:"type,omitempty"`
}

// Aggregate 聚合结果。
type Aggregate struct {
	Projects  []ProjectEntry `json:"projects"`
	Generated string         `json:"generated_at"`
	Total     int            `json:"total_projects"`
	// 全局统计（跨项目）
	GlobalTasks      int `json:"global_tasks_total"`
	GlobalDone       int `json:"global_tasks_done"`
	GlobalCompletion int `json:"global_completion"`
	GlobalDocs       int `json:"global_docs_total"`
}

// LoadConfig 读取 projects.yaml，返回配置。
// 若给出显式 overridePaths，则忽略配置文件里的 path，用这些 path 逐个扫描（供 -p 参数）。
func LoadConfig(path string, overridePaths []string) (*Config, error) {
	cfg := &Config{}
	if len(overridePaths) > 0 {
		for _, p := range overridePaths {
			abs, err := filepath.Abs(p)
			if err != nil {
				return nil, err
			}
			cfg.Projects = append(cfg.Projects, ProjectSpec{Path: abs})
		}
		return cfg, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("解析 %s: %w", path, err)
	}
	// 相对路径基于配置文件所在目录解析；~ 展开为 HOME
	base := filepath.Dir(path)
	for i := range cfg.Projects {
		p := cfg.Projects[i].Path
		if strings.HasPrefix(p, "~/") {
			home, _ := os.UserHomeDir()
			p = filepath.Join(home, p[2:])
		}
		if !filepath.IsAbs(p) {
			p = filepath.Join(base, p)
		}
		cfg.Projects[i].Path = p
	}
	return cfg, nil
}

// ScanAll 扫描配置中的所有项目，返回聚合结果。
func ScanAll(cfg *Config) *Aggregate {
	agg := &Aggregate{Generated: nowTs()}
	var gTasks, gDone, gDocs int
	for _, spec := range cfg.Projects {
		prj := scan.Scan(spec.Path, spec.Exclude, spec.Dirs)
		name := spec.Name
		if name == "" {
			name = prj.Name
		}
		entry := ProjectEntry{
			Project: *prj,
			Status:  spec.Status,
			Type:    spec.Type,
		}
		entry.Name = name
		agg.Projects = append(agg.Projects, entry)
		gTasks += prj.Stats.TasksTotal
		gDone += prj.Stats.TasksDone
		gDocs += prj.Stats.TotalDocs
		if spec.Name != "" {
			// Name 覆盖
		}
	}
	sort.Slice(agg.Projects, func(i, j int) bool {
		return strings.ToLower(agg.Projects[i].Name) < strings.ToLower(agg.Projects[j].Name)
	})
	agg.Total = len(agg.Projects)
	agg.GlobalTasks = gTasks
	agg.GlobalDone = gDone
	agg.GlobalDocs = gDocs
	if gTasks > 0 {
		agg.GlobalCompletion = gDone * 100 / gTasks
	}
	return agg
}

// CollectSignature 聚合签名：所有项目签名合并（用于 watch 判断变化）。
func CollectSignature(cfg *Config) map[string]map[string]int64 {
	out := map[string]map[string]int64{}
	for _, spec := range cfg.Projects {
		out[spec.Path] = scan.CollectSignature(spec.Path, spec.Exclude, spec.Dirs)
	}
	return out
}

// SignaturesEqual 判断两组聚合签名是否一致。
func SignaturesEqual(a, b map[string]map[string]int64) bool {
	if len(a) != len(b) {
		return false
	}
	for k, sa := range a {
		sb, ok := b[k]
		if !ok || len(sa) != len(sb) {
			return false
		}
		for p, t := range sa {
			if sb[p] != t {
				return false
			}
		}
	}
	return true
}

func nowTs() string {
	return time.Now().Format("2006-01-02 15:04:05")
}
