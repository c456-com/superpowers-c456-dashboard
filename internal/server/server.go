// Package server 提供本地 HTTP 服务：静态前端 + /data(聚合 JSON) + /events(SSE 自动刷新) + watch。
package server

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"superpowers-c456-dashboard/internal/aggregate"
	"superpowers-c456-dashboard/internal/config"
)

//go:embed all:dist
var distFS embed.FS

// Config 服务配置。
type Config struct {
	Port       int
	Interval   int // watch 轮询秒数
	NoWatch    bool
	ConfigPath string // 项目清单路径（默认全局配置路径）
}

// Server 服务状态。
type Server struct {
	cfg          Config
	agg          *aggregate.Aggregate
	projectRoots map[string]string // 项目名 → 根目录（file API 用）
	sig          map[string]map[string]int64
	mu           sync.RWMutex
	clients      map[chan string]struct{}
	clientsMu    sync.Mutex
	configPath   string

	scanMu sync.Mutex
	scan   ScanState // 目录扫描运行状态（前端展示实时进度）
}

// ScanState 目录扫描的实时运行状态。
type ScanState struct {
	Running bool                `json:"running"`
	Current string              `json:"current"` // 正在扫描的目录
	Found   []config.Discovered `json:"found"`   // 已发现的候选
	Done    bool                `json:"done"`
	Error   string              `json:"error,omitempty"`
}

// New 创建服务并做首次扫描。
func New(cfg Config) (*Server, error) {
	s := &Server{
		cfg:        cfg,
		configPath: cfg.ConfigPath,
		clients:    map[chan string]struct{}{},
	}
	specs, err := config.LoadSpecs(cfg.ConfigPath)
	if err != nil {
		return nil, err
	}
	_ = config.EnsureDir()
	s.recompute(&aggregate.Config{Projects: specs})
	return s, nil
}

// CurrentConfig 返回当前加载的项目配置路径与覆盖。
func (s *Server) recompute(cfg *aggregate.Config) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.agg = aggregate.ScanAll(cfg)
	s.projectRoots = map[string]string{}
	for _, p := range cfg.Projects {
		s.projectRoots[p.Name] = p.Path
	}
	s.sig = aggregate.CollectSignature(cfg)
}

// reloadFromDisk 从当前配置路径重读项目清单并重扫。
func (s *Server) reloadFromDisk() {
	specs, err := config.LoadSpecs(s.configPath)
	if err != nil {
		slog.Error("reload config", "err", err)
		return
	}
	s.recompute(&aggregate.Config{Projects: specs})
	s.broadcast("refresh")
}

// Reload 重新加载配置并重扫。
func (s *Server) Reload() {
	s.reloadFromDisk()
}

// handler 主 HTTP 处理器。
func (s *Server) handler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /data", func(w http.ResponseWriter, r *http.Request) {
		s.mu.RLock()
		data, _ := json.Marshal(s.agg)
		s.mu.RUnlock()
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store")
		w.Write(data)
	})

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Write([]byte(`{"ok":true}`))
	})

	// /api/file：安全读取项目内文件（文档里的文件路径链接闭环）。
	// 只允许读取「已配置项目根目录内」的文件（resolve 后校验前缀，防目录穿越）。
	mux.HandleFunc("GET /api/file", s.fileHandler)

	// 项目管理：添加/移除/扫描目录识别（写全局配置 → 重扫 → SSE 刷新）
	mux.HandleFunc("POST /api/projects", s.addProjectHandler)
	mux.HandleFunc("DELETE /api/projects/{name}", s.removeProjectHandler)
	mux.HandleFunc("POST /api/scan-dir", s.scanDirHandler)
	mux.HandleFunc("GET /api/scan/status", s.scanStatusHandler)

	mux.HandleFunc("GET /events", s.sseHandler)

	// 静态资源 + SPA fallback
	sub, _ := fs.Sub(distFS, "dist")
	fileServer := http.FileServer(http.FS(sub))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		p := r.URL.Path
		if p == "/" || p == "/index.html" {
			fileServer.ServeHTTP(w, r)
			return
		}
		trimmed := p[1:]
		if _, err := fs.Stat(sub, trimmed); err == nil {
			fileServer.ServeHTTP(w, r)
			return
		}
		// SPA fallback → index.html
		index, _ := fs.ReadFile(sub, "index.html")
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(index)
	})

	return s.withLogging(mux)
}

func (s *Server) withLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		slog.Info("http", "method", r.Method, "path", r.URL.Path, "dur", time.Since(start).Round(time.Millisecond))
	})
}

func (s *Server) sseHandler(w http.ResponseWriter, r *http.Request) {
	fl, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)

	ch := make(chan string, 8)
	s.clientsMu.Lock()
	s.clients[ch] = struct{}{}
	s.clientsMu.Unlock()
	defer func() {
		s.clientsMu.Lock()
		delete(s.clients, ch)
		s.clientsMu.Unlock()
	}()

	// 初始连接消息
	_, _ = w.Write([]byte("data: connected\n\n"))
	fl.Flush()

	for {
		select {
		case msg := <-ch:
			_, _ = w.Write([]byte("data: " + msg + "\n\n"))
			fl.Flush()
		case <-r.Context().Done():
			return
		}
	}
}

func (s *Server) broadcast(msg string) {
	s.clientsMu.Lock()
	defer s.clientsMu.Unlock()
	for ch := range s.clients {
		select {
		case ch <- msg:
		default:
		}
	}
}

// fileHandler 安全读取项目内文件（文档里文件路径链接闭环）。
// 请求：GET /api/file?project=<项目名>&path=<项目根内相对路径>[&dir=<文档所在目录>]
// dir 可选：markdown 相对链接语义是相对文档所在目录；先试 root/dir/path，再退回 root/path。
// 安全：resolve 后必须落在该项目根目录内，否则拒绝（防目录穿越）。
func (s *Server) fileHandler(w http.ResponseWriter, r *http.Request) {
	project := r.URL.Query().Get("project")
	rel := r.URL.Query().Get("path")
	dir := r.URL.Query().Get("dir") // 文档所在目录（相对项目根，如 docs）
	if project == "" || rel == "" {
		writeJSONError(w, 400, "缺少 project 或 path 参数")
		return
	}
	s.mu.RLock()
	root, ok := s.projectRoots[project]
	s.mu.RUnlock()
	if !ok {
		writeJSONError(w, 404, "未知项目: "+project)
		return
	}
	rootAbs, _ := filepath.Abs(root)

	resolve := func(rel string) (string, bool) {
		abs, err := filepath.Abs(filepath.Join(rootAbs, rel))
		if err != nil {
			return "", false
		}
		// 防目录穿越：abs 必须在 rootAbs 内
		relChk, err := filepath.Rel(rootAbs, abs)
		if err != nil || relChk == ".." || strings.HasPrefix(relChk, ".."+string(filepath.Separator)) {
			return "", false
		}
		return abs, true
	}

	// 候选取绝对路径：优先 root/dir/rel（markdown 相对文档目录），再退回 root/rel（相对项目根）
	// 逐个尝试，文件实际存在读取成功才用（dir 版本文件不存在时自动落到根版本）
	var abs string
	anyInside := false // 是否有候选解析到项目根内（用于区分 404 与 400）
	candidates := []string{rel}
	if dir != "" && dir != "." {
		candidates = []string{filepath.Join(dir, rel), rel}
	}
	read := func(s string) ([]byte, bool) {
		a, ok := resolve(s)
		if !ok {
			return nil, false
		}
		anyInside = true
		d, err := os.ReadFile(a)
		if err != nil {
			return nil, false
		}
		abs = a
		return d, true
	}
	var data []byte
	var found bool
	for _, c := range candidates {
		if d, ok := read(c); ok {
			data, found = d, true
			break
		}
	}
	if !found {
		// 有候选在根内（只是不存在）→ 404；全部越界 → 400
		if anyInside {
			writeJSONError(w, 404, "文件不存在: "+rel)
		} else {
			writeJSONError(w, 400, "路径超出项目根目录")
		}
		return
	}
	// 限制文件大小（防意外大文件），默认 1MB
	if len(data) > 1<<20 {
		writeJSONError(w, 413, "文件过大(>1MB)，仅支持查看小文件")
		return
	}
	resp := map[string]string{
		"project": project,
		"path":    rel,
		"abs":     abs,
		"content": string(data),
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(resp)
}

func writeJSONError(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

// addProjectHandler 添加一个项目到全局配置（POST /api/projects {path, name?}）。
func (s *Server) addProjectHandler(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Path string `json:"path"`
		Name string `json:"name,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil || in.Path == "" {
		writeJSONError(w, 400, "缺少 path")
		return
	}
	if _, err := os.Stat(in.Path); err != nil {
		writeJSONError(w, 400, "目录不存在或不可读: "+in.Path)
		return
	}
	abs, _ := filepath.Abs(in.Path)
	name := in.Name
	if name == "" {
		name = filepath.Base(abs)
	}
	spec := aggregate.ProjectSpec{Name: name, Path: abs, Type: "superpowers"}
	if err := s.addSpec(spec); err != nil {
		writeJSONError(w, 500, err.Error())
		return
	}
	writeJSON(w, map[string]string{"ok": "added", "name": name})
}

// removeProjectHandler 从全局配置移除项目（DELETE /api/projects/{name}）。
func (s *Server) removeProjectHandler(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if name == "" {
		writeJSONError(w, 400, "缺少项目名")
		return
	}
	if err := s.removeSpec(name); err != nil {
		writeJSONError(w, 500, err.Error())
		return
	}
	writeJSON(w, map[string]string{"ok": "removed", "name": name})
}

// scanDirHandler 异步递归扫描目录识别 superpowers 项目（不自动添加）。
// POST /api/scan-dir {path, max_depth?} → 立即返回 {started:true}；
// 扫描状态经 GET /api/scan/status 轮询（running/current/found/done）。
func (s *Server) scanDirHandler(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Path     string `json:"path"`
		MaxDepth int    `json:"max_depth,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil || in.Path == "" {
		writeJSONError(w, 400, "缺少 path")
		return
	}
	if _, err := os.Stat(in.Path); err != nil {
		writeJSONError(w, 400, "目录不存在: "+in.Path)
		return
	}
	depth := in.MaxDepth
	if depth <= 0 {
		depth = 3
	}

	s.scanMu.Lock()
	if s.scan.Running {
		s.scanMu.Unlock()
		writeJSONError(w, 409, "已有扫描任务进行中")
		return
	}
	s.scan = ScanState{Running: true, Current: in.Path, Found: []config.Discovered{}}
	s.scanMu.Unlock()

	go func() {
		root := in.Path
		defer func() {
			s.scanMu.Lock()
			s.scan.Running = false
			s.scan.Done = true
			if rec := recover(); rec != nil {
				s.scan.Error = "扫描异常: " + fmt.Sprint(rec)
			}
			s.scanMu.Unlock()
		}()
		found := config.ScanForProjects(root, depth, func(dir string) {
			s.scanMu.Lock()
			s.scan.Current = dir
			s.scanMu.Unlock()
		})
		s.scanMu.Lock()
		s.scan.Found = make([]config.Discovered, len(found))
		copy(s.scan.Found, found)
		s.scan.Current = ""
		s.scanMu.Unlock()
	}()

	writeJSON(w, map[string]bool{"started": true})
}

// scanStatusHandler 返回目录扫描实时状态（GET /api/scan/status）。
func (s *Server) scanStatusHandler(w http.ResponseWriter, r *http.Request) {
	s.scanMu.Lock()
	state := s.scan
	s.scanMu.Unlock()
	writeJSON(w, state)
}

// addSpec 添加项目配置并持久化 + 重扫 + 广播。
func (s *Server) addSpec(spec aggregate.ProjectSpec) error {
	s.mu.RLock()
	specs := s.specsSnapshot()
	s.mu.RUnlock()
	// 去重（按路径）
	for _, ex := range specs {
		if ex.Path == spec.Path {
			return nil // 已存在
		}
	}
	specs = append(specs, spec)
	if err := config.SaveSpecs(s.configPath, specs); err != nil {
		return err
	}
	_ = config.EnsureDir()
	s.recompute(&aggregate.Config{Projects: specs})
	s.broadcast("refresh")
	return nil
}

// removeSpec 移除项目配置并持久化。
func (s *Server) removeSpec(name string) error {
	s.mu.RLock()
	specs := s.specsSnapshot()
	s.mu.RUnlock()
	kept := specs[:0]
	removed := false
	for _, sp := range specs {
		if sp.Name != name && sp.Path != name {
			kept = append(kept, sp)
		} else {
			removed = true
		}
	}
	if !removed {
		return nil
	}
	if err := config.SaveSpecs(s.configPath, kept); err != nil {
		return err
	}
	s.recompute(&aggregate.Config{Projects: kept})
	s.broadcast("refresh")
	return nil
}

// specsSnapshot 从当前聚合结果反推项目清单（避免额外读盘）。
func (s *Server) specsSnapshot() []aggregate.ProjectSpec {
	out := make([]aggregate.ProjectSpec, 0, len(s.agg.Projects))
	for _, p := range s.agg.Projects {
		out = append(out, aggregate.ProjectSpec{
			Name:   p.Name,
			Path:   p.Root,
			Status: p.Status,
			Type:   p.Type,
		})
	}
	return out
}

func writeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(v)
}

// Watch 后台轮询文档签名变化，变化则重扫并广播。
func (s *Server) Watch() {
	if s.cfg.NoWatch {
		return
	}
	interval := time.Duration(s.cfg.Interval) * time.Second
	if interval <= 0 {
		interval = 2 * time.Second
	}
	go func() {
		for {
			time.Sleep(interval)
			// 重读配置清单 + 重扫
			specs, err := config.LoadSpecs(s.configPath)
			if err != nil {
				slog.Error("watch load config", "err", err)
				continue
			}
			loaded := &aggregate.Config{Projects: specs}
			sig := aggregate.CollectSignature(loaded)
			s.mu.RLock()
			old := s.sig
			s.mu.RUnlock()
			if !aggregate.SignaturesEqual(sig, old) {
				s.recompute(loaded)
				s.broadcast("refresh")
				slog.Info("watch", "msg", "项目文档变化，已重扫")
			}
		}
	}()
}

// Run 启动 HTTP 服务并阻塞。
func (s *Server) Run() error {
	srv := &http.Server{
		Addr:    "127.0.0.1:" + itoa(s.cfg.Port),
		Handler: s.handler(),
	}
	slog.Info("dashboard 启动", "url", "http://127.0.0.1:"+itoa(s.cfg.Port))
	return srv.ListenAndServe()
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var b [20]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		b[i] = '-'
	}
	return string(b[i:])
}
