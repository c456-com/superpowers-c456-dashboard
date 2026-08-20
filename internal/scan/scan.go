// Package scan 扫描解析 superpowers 扁平 markdown 文档，生成结构化项目数据。
//
// 移植自 Python 版 project-dev-dashboard 的 scan.py，逻辑保持一致：
//   - 类型判定：spec / plan / roadmap / sprint / research / doc
//   - 日期：文件名前缀 YYYY-MM-DD，回退到引用块日期
//   - 元数据：文件开头 `> ` 引用块里 `**键：**值`
//   - 任务：`- [ ] x`(未完成) / `- [x] x`(完成)
//   - Roadmap 阶段：`### 阶段X 标题` 小节
package scan

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

// Project 单项目的扫描结果。
type Project struct {
	Name        string         `json:"name"`
	Root        string         `json:"root"`
	GeneratedAt string         `json:"generated_at"`
	Stats       Stats          `json:"stats"`
	Documents   []Document     `json:"documents"`
	Roadmap     []RoadmapStage `json:"roadmap_stages"`
}

// Stats 项目统计。
type Stats struct {
	TotalDocs     int            `json:"total_docs"`
	ByType        map[string]int `json:"by_type"`
	TasksTotal    int            `json:"tasks_total"`
	TasksDone     int            `json:"tasks_done"`
	Completion    int            `json:"completion"`
	SpecsTotal    int            `json:"specs_total"`
	PlansTotal    int            `json:"plans_total"`
	SprintsTotal  int            `json:"sprints_total"`
	RoadmapsTotal int            `json:"roadmaps_total"`
	LastScan      string         `json:"last_scan"`
}

// Document 单篇文档解析结果。
type Document struct {
	Path     string            `json:"path"`
	Title    string            `json:"title"`
	Date     string            `json:"date"`
	Type     string            `json:"type"`
	Status   string            `json:"status"`
	Meta     map[string]string `json:"meta"`
	Summary  string            `json:"summary"`
	Sections []Section         `json:"sections"`
	Tasks    []Task            `json:"tasks"`
	Content  string            `json:"content"`
	Mtime    int64             `json:"mtime"`
}

// Section 文档小节。
type Section struct {
	Level int    `json:"level"`
	Title string `json:"title"`
}

// Task 任务项。
type Task struct {
	Text string `json:"text"`
	Done bool   `json:"done"`
}

// RoadmapStage 开发阶段。
type RoadmapStage struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Desc  string `json:"desc"`
}

// 默认忽略目录。
var defaultIgnores = map[string]bool{
	".git": true, ".hg": true, ".svn": true, "node_modules": true,
	"vendor": true, "dist": true, "build": true, ".next": true, ".nuxt": true,
	"target": true, "__pycache__": true, ".hermes": true, ".idea": true,
	".vscode": true, "coverage": true, "venv": true, ".venv": true,
	".tox": true, "data": true,
}

// DefaultIgnores 暴露默认忽略集（供 signature/watch 复用）。
func DefaultIgnores() map[string]bool { return defaultIgnores }

var (
	titleRe      = regexp.MustCompile(`^#\s+(.+)$`)
	metaRe       = regexp.MustCompile(`\*\*([^*]+?)[：:]\*\*\s*(.*)$`)
	plainMetaRe  = regexp.MustCompile(`^([\p{Han}\w][^：:]{0,20})[：:]\s*(.+)$`)
	dateRe       = regexp.MustCompile(`(\d{4}-\d{2}-\d{2})`)
	datePrefixRe = regexp.MustCompile(`^(\d{4}-\d{2}-\d{2})`)
	taskRe       = regexp.MustCompile(`^\s*[-*]\s*\[([ xX])\]\s*(.+)$`)
	sectionRe    = regexp.MustCompile(`^(#{2,6})\s+(.+)$`)
	stageRe      = regexp.MustCompile(`^\s*(#{2,6})\s*阶段\s*([①-⑳0-9一二三四五六七八九十]+)[、.\s]*\s*(.*)$`)
	arrowRe      = regexp.MustCompile(`^\s*([①-⑳0-9一二三四五六七八九十]+)\s+([^\s→]+.*?)\s*(?:→|$)`)
)

var noiseSections = []string{"记录", "实测", "附录", "总结", "复盘"}

// classify 判定文档类型。
func classify(path, title string) string {
	p := "/" + strings.ToLower(strings.ReplaceAll(path, "\\", "/"))
	t := strings.ToLower(title)
	if strings.Contains(t, "roadmap") || strings.Contains(t, "路线图") || strings.Contains(t, "开发路线") {
		return "roadmap"
	}
	if strings.Contains(t, "sprint") || strings.Contains(t, "冲刺") {
		return "sprint"
	}
	if strings.Contains(p, "/specs/") || strings.Contains(t, "(spec") ||
		(strings.Contains(t, "设计") && !strings.Contains(t, "实现")) ||
		strings.Contains(t, "架构") || strings.Contains(t, "architecture") {
		return "spec"
	}
	if strings.Contains(t, "实现计划") || strings.Contains(t, "开发计划") ||
		(strings.Contains(t, "计划") && !strings.Contains(t, "设计")) {
		return "plan"
	}
	if strings.Contains(p, "/research/") || strings.Contains(t, "调研") ||
		strings.Contains(t, "探测") || strings.Contains(t, "investigation") {
		return "research"
	}
	return "doc"
}

func parseMeta(blockquote []string) map[string]string {
	meta := map[string]string{}
	for _, line := range blockquote {
		// 支持 `**键：**值`（加粗）和 `键：值`（普通）两种风格
		if m := metaRe.FindStringSubmatch(line); m != nil {
			meta[strings.TrimSpace(m[1])] = strings.TrimSpace(m[2])
			continue
		}
		if m := plainMetaRe.FindStringSubmatch(line); m != nil {
			meta[strings.TrimSpace(m[1])] = strings.TrimSpace(m[2])
		}
	}
	return meta
}

func parseDate(fname string, meta map[string]string) string {
	if m := datePrefixRe.FindStringSubmatch(fname); m != nil {
		return m[1]
	}
	for _, v := range meta {
		if m := dateRe.FindStringSubmatch(v); m != nil {
			return m[1]
		}
	}
	return ""
}

func extractTasks(lines []string) []Task {
	tasks := make([]Task, 0)
	for _, ln := range lines {
		m := taskRe.FindStringSubmatch(ln)
		if m != nil {
			tasks = append(tasks, Task{
				Text: strings.TrimSpace(m[2]),
				Done: strings.ToLower(m[1]) == "x",
			})
		}
	}
	return tasks
}

func extractSections(lines []string) []Section {
	sections := make([]Section, 0)
	for _, ln := range lines {
		m := sectionRe.FindStringSubmatch(ln)
		if m != nil {
			sections = append(sections, Section{Level: len(m[1]), Title: strings.TrimSpace(m[2])})
		}
	}
	return sections
}

func extractRoadmapStages(lines []string) []RoadmapStage {
	var stages []RoadmapStage
	var cur *RoadmapStage
	for _, ln := range lines {
		m := stageRe.FindStringSubmatch(ln)
		if m != nil {
			title := strings.TrimSpace(m[3])
			skip := false
			for _, k := range noiseSections {
				if strings.Contains(title, k) {
					skip = true
					break
				}
			}
			if skip {
				continue
			}
			if cur != nil {
				stages = append(stages, *cur)
			}
			id := m[2]
			tt := title
			if tt == "" {
				tt = "阶段" + id
			}
			cur = &RoadmapStage{ID: id, Title: tt, Desc: ""}
			continue
		}
		if cur != nil && cur.Desc == "" && strings.TrimSpace(ln) != "" &&
			!strings.HasPrefix(strings.TrimSpace(ln), "```") &&
			!strings.HasPrefix(strings.TrimSpace(ln), "|") {
			cur.Desc = strings.TrimSpace(ln)
		}
	}
	if cur != nil {
		stages = append(stages, *cur)
	}
	if len(stages) > 0 {
		return stages
	}
	// 回退：箭头开发顺序
	var seq []RoadmapStage
	for _, ln := range lines {
		am := arrowRe.FindStringSubmatch(ln)
		if am != nil {
			seq = append(seq, RoadmapStage{ID: am[1], Title: strings.TrimSpace(am[2]), Desc: ""})
		}
	}
	return seq
}

func firstSentence(meta map[string]string, content string) string {
	for _, k := range []string{"辉哥需求", "目的", "背景", "里程碑", "目标"} {
		if v, ok := meta[k]; ok && v != "" {
			return v
		}
	}
	for _, ln := range strings.Split(content, "\n") {
		s := strings.TrimSpace(ln)
		if s != "" && !strings.HasPrefix(s, "#") && !strings.HasPrefix(s, ">") &&
			!strings.HasPrefix(s, "```") && !strings.HasPrefix(s, "---") &&
			!strings.HasPrefix(s, "- [") {
			if len(s) > 200 {
				return s[:200]
			}
			return s
		}
	}
	return ""
}

func parseDoc(path, root string) *Document {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		rel = path
	}
	contentBytes, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	content := string(contentBytes)
	lines := strings.Split(content, "\n")

	title := ""
	for _, ln := range lines {
		m := titleRe.FindStringSubmatch(ln)
		if m != nil {
			title = strings.TrimSpace(m[1])
			break
		}
	}

	var blockquote []string
	sawContent := false
	for _, ln := range lines {
		if strings.HasPrefix(ln, ">") {
			blockquote = append(blockquote, strings.TrimSpace(ln[1:]))
		} else if strings.TrimSpace(ln) == "" {
			continue
		} else if !sawContent {
			// 首个非空非引用行：若是标题(#)则跳过继续收集；否则停止
			if strings.HasPrefix(strings.TrimSpace(ln), "# ") || strings.HasPrefix(ln, "---") {
				sawContent = true
				continue
			}
			break
		} else {
			break
		}
	}
	meta := parseMeta(blockquote)
	fname := filepath.Base(path)
	date := parseDate(fname, meta)
	docType := classify(rel, title)

	info, _ := os.Stat(path)
	mtime := int64(0)
	if info != nil {
		mtime = info.ModTime().Unix()
	}

	return &Document{
		Path:     rel,
		Title:    titleOr(title, fname),
		Date:     date,
		Type:     docType,
		Status:   meta["状态"],
		Meta:     meta,
		Summary:  firstSentence(meta, content),
		Sections: extractSections(lines),
		Tasks:    extractTasks(lines),
		Content:  content,
		Mtime:    mtime,
	}
}

func titleOr(title, fname string) string {
	if title != "" {
		return title
	}
	return fname
}

// Collect ext 排除规则：ext 是额外的忽略目录名。
func walk(base string, extIgnore map[string]bool, root string, fn func(doc *Document)) {
	filepath.WalkDir(base, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			if path != base {
				name := d.Name()
				if defaultIgnores[name] || extIgnore[name] || strings.HasPrefix(name, ".") {
					return filepath.SkipDir
				}
			}
			return nil
		}
		if !strings.HasSuffix(strings.ToLower(d.Name()), ".md") {
			return nil
		}
		if doc := parseDoc(path, root); doc != nil {
			fn(doc)
		}
		return nil
	})
}

// Scan 扫描一个项目根目录，返回项目数据。
// root 项目根；extIgnore 额外忽略目录；docDirs 额外递归扫描文档目录（默认 docs）。
func Scan(root string, extIgnore []string, docDirs []string) *Project {
	rootAbs, _ := filepath.Abs(root)
	extIgnoreSet := map[string]bool{}
	for _, d := range extIgnore {
		extIgnoreSet[d] = true
	}
	incDirs := []string{"docs"}
	incDirs = append(incDirs, docDirs...)

	var docs []Document
	seen := map[string]bool{}
	collect := func(doc *Document) {
		if !seen[doc.Path] {
			seen[doc.Path] = true
			docs = append(docs, *doc)
		}
	}

	// 顶层 *.md（非递归）
	entries, _ := os.ReadDir(rootAbs)
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(strings.ToLower(e.Name()), ".md") {
			if doc := parseDoc(filepath.Join(rootAbs, e.Name()), rootAbs); doc != nil {
				collect(doc)
			}
		}
	}
	// 文档目录（递归）
	for _, d := range incDirs {
		if d == "" {
			continue
		}
		p := filepath.Join(rootAbs, d)
		if st, err := os.Stat(p); err == nil && st.IsDir() {
			walk(p, extIgnoreSet, rootAbs, collect)
		}
	}

	sort.Slice(docs, func(i, j int) bool {
		if docs[i].Date != docs[j].Date {
			return docs[i].Date < docs[j].Date
		}
		return docs[i].Path < docs[j].Path
	})

	stats := buildStats(docs)
	roads := filterType(docs, "roadmap")
	var stages []RoadmapStage
	for _, r := range roads {
		stages = append(stages, extractRoadmapStages(strings.Split(r.Content, "\n"))...)
	}

	return &Project{
		Name:        filepath.Base(strings.TrimRight(rootAbs, string(filepath.Separator))),
		Root:        rootAbs,
		GeneratedAt: time.Now().Format("2006-01-02 15:04:05"),
		Stats:       stats,
		Documents:   docs,
		Roadmap:     stages,
	}
}

func filterType(docs []Document, t string) []Document {
	var out []Document
	for _, d := range docs {
		if d.Type == t {
			out = append(out, d)
		}
	}
	return out
}

func buildStats(docs []Document) Stats {
	byType := map[string]int{}
	for _, d := range docs {
		byType[d.Type]++
	}
	var allTasks []Task
	for _, d := range docs {
		allTasks = append(allTasks, d.Tasks...)
	}
	done := 0
	for _, t := range allTasks {
		if t.Done {
			done++
		}
	}
	completion := 0
	if len(allTasks) > 0 {
		completion = done * 100 / len(allTasks)
	}
	return Stats{
		TotalDocs:     len(docs),
		ByType:        byType,
		TasksTotal:    len(allTasks),
		TasksDone:     done,
		Completion:    completion,
		SpecsTotal:    byType["spec"],
		PlansTotal:    byType["plan"],
		SprintsTotal:  byType["sprint"],
		RoadmapsTotal: byType["roadmap"],
		LastScan:      time.Now().Format("2006-01-02 15:04:05"),
	}
}

// CollectSignature 轻量签名：列出 (relpath, mtime)，用于判断是否变化。
func CollectSignature(root string, extIgnore []string, docDirs []string) map[string]int64 {
	rootAbs, _ := filepath.Abs(root)
	extIgnoreSet := map[string]bool{}
	for _, d := range extIgnore {
		extIgnoreSet[d] = true
	}
	incDirs := []string{"docs"}
	incDirs = append(incDirs, docDirs...)
	sig := map[string]int64{}

	addFile := func(path string) {
		if !strings.HasSuffix(strings.ToLower(filepath.Base(path)), ".md") {
			return
		}
		info, err := os.Stat(path)
		if err != nil {
			return
		}
		rel, _ := filepath.Rel(rootAbs, path)
		sig[rel] = info.ModTime().Unix()
	}

	entries, _ := os.ReadDir(rootAbs)
	for _, e := range entries {
		if !e.IsDir() {
			addFile(filepath.Join(rootAbs, e.Name()))
		}
	}
	for _, d := range incDirs {
		if d == "" {
			continue
		}
		p := filepath.Join(rootAbs, d)
		if st, err := os.Stat(p); err == nil && st.IsDir() {
			filepath.WalkDir(p, func(path string, de os.DirEntry, err error) error {
				if err != nil {
					return nil
				}
				if de.IsDir() {
					if path != p {
						name := de.Name()
						if defaultIgnores[name] || extIgnoreSet[name] || strings.HasPrefix(name, ".") {
							return filepath.SkipDir
						}
					}
					return nil
				}
				addFile(path)
				return nil
			})
		}
	}
	return sig
}
