// Package server 提供本地 HTTP 服务：静态前端 + /data(聚合 JSON) + /events(SSE 自动刷新) + watch。
package server

import (
	"embed"
	"encoding/json"
	"io/fs"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"project-dev-dashboard/internal/aggregate"
)

//go:embed all:dist
var distFS embed.FS

// Config 服务配置。
type Config struct {
	Port       int
	Interval   int // watch 轮询秒数
	NoWatch    bool
	ConfigPath string
	Overrides  []string // -p 参数指定的项目路径（跳过 projects.yaml 的 path）
}

// Server 服务状态。
type Server struct {
	cfg        Config
	agg        *aggregate.Aggregate
	sig        map[string]map[string]int64
	mu         sync.RWMutex
	clients    map[chan string]struct{}
	clientsMu  sync.Mutex
	configPath string
}

// New 创建服务并做首次扫描。
func New(cfg Config) (*Server, error) {
	loaded, err := aggregate.LoadConfig(cfg.ConfigPath, cfg.Overrides)
	if err != nil {
		return nil, err
	}
	s := &Server{
		cfg:        cfg,
		configPath: cfg.ConfigPath,
		clients:    map[chan string]struct{}{},
	}
	s.recompute(loaded)
	return s, nil
}

// CurrentConfig 返回当前加载的项目配置路径与覆盖。
func (s *Server) recompute(cfg *aggregate.Config) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.agg = aggregate.ScanAll(cfg)
	s.sig = aggregate.CollectSignature(cfg)
}

// Reload 重新加载配置并重扫。
func (s *Server) Reload() {
	loaded, err := aggregate.LoadConfig(s.configPath, s.cfg.Overrides)
	if err != nil {
		slog.Error("reload config", "err", err)
		return
	}
	s.recompute(loaded)
	s.broadcast("refresh")
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
			// 配置可能包含覆盖路径，这里用跳过配置文件重载的方式只做重扫
			loaded, err := aggregate.LoadConfig(s.configPath, s.cfg.Overrides)
			if err != nil {
				slog.Error("watch load config", "err", err)
				continue
			}
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
