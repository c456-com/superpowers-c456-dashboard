package server

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"superpowers-c456-dashboard/internal/ai"
	"superpowers-c456-dashboard/internal/scan"
)

// aiConfigHandler GET /api/ai/config → 返回 AI 配置（不含 key 明文，仅返是否已配 + base/model）。
func (s *Server) aiConfigHandler(w http.ResponseWriter, r *http.Request) {
	cfg, _ := ai.LoadConfig(s.authPath)
	writeJSON(w, map[string]interface{}{
		"base_url": cfg.BaseURL,
		"model":    cfg.Model,
		"has_key":  cfg.APIKey != "",
	})
}

// aiSaveConfigHandler POST /api/ai/config {base_url, model, api_key} → 保存到 auth.json。
func (s *Server) aiSaveConfigHandler(w http.ResponseWriter, r *http.Request) {
	var cfg ai.Config
	if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
		writeJSONError(w, 400, "参数错误: "+err.Error())
		return
	}
	cfg.BaseURL = strings.TrimSpace(cfg.BaseURL)
	cfg.Model = strings.TrimSpace(cfg.Model)
	if cfg.BaseURL == "" {
		writeJSONError(w, 400, "缺少 base_url")
		return
	}
	if err := ai.SaveConfig(s.authPath, cfg); err != nil {
		writeJSONError(w, 500, "保存失败: "+err.Error())
		return
	}
	writeJSON(w, map[string]bool{"ok": true, "has_key": cfg.APIKey != ""})
}

// aiAnalyseHandler POST /api/ai/analyse [{project}] → 手动触发对项目跑一次 L3 agent。
// 异步：立即返回 started；结果写入 s.suggestions。前端轮询 GET /api/ai/suggestions。
func (s *Server) aiAnalyseHandler(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Project string `json:"project"`
	}
	_ = json.NewDecoder(r.Body).Decode(&in)
	cfg, err := ai.LoadConfig(s.authPath)
	if err != nil || cfg.BaseURL == "" || cfg.Model == "" {
		writeJSONError(w, 400, "未配置 AI 模型，请先在设置里配置")
		return
	}
	project := in.Project
	// 支持缺省：取第一个项目（总览上下文）
	var root string
	if project != "" {
		s.mu.RLock()
		root = s.projectRoots[project]
		s.mu.RUnlock()
		if root == "" {
			writeJSONError(w, 404, "项目不存在: "+project)
			return
		}
	} else {
		s.mu.RLock()
		for _, p := range s.agg.Projects {
			root = p.Root
			project = p.Name
			break
		}
		s.mu.RUnlock()
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 600*time.Second)
		defer cancel()
		s.broadcast(`{"type":"ai","event":"running","project":"` + project + `"}`)
		sugs, err := runAgentOnce(ctx, cfg, project, root)
		s.sugMu.Lock()
		s.suggestionsAt = time.Now().Format("15:04:05")
		if err != nil {
			s.suggestions = []ai.Suggestion{{
				Type: "ok", Severity: "warning", Title: "AI 分析失败",
				Detail: err.Error(), Action: "检查 AI 配置后重试",
			}}
		} else {
			s.suggestions = sugs
		}
		done := s.suggestions
		s.sugMu.Unlock()
		s.broadcast(`{"type":"ai","event":"done","project":"` + project + `"}`)
		_ = done
	}()
	writeJSON(w, map[string]bool{"started": true})
}

// aiSuggestionsHandler GET /api/ai/suggestions → 返回最近分析结果。
func (s *Server) aiSuggestionsHandler(w http.ResponseWriter, r *http.Request) {
	s.sugMu.Lock()
	at := s.suggestionsAt
	out := make([]ai.Suggestion, len(s.suggestions))
	copy(out, s.suggestions)
	s.sugMu.Unlock()
	if at == "" {
		writeJSON(w, map[string]interface{}{"at": "", "suggestions": []ai.Suggestion{}})
		return
	}
	writeJSON(w, map[string]interface{}{"at": at, "suggestions": out})
}

// runAgentOnce 构建工具上下文并跑一次 L3 agent。
func runAgentOnce(ctx context.Context, cfg ai.Config, project, root string) ([]ai.Suggestion, error) {
	client := ai.NewClient(cfg)
	toolCtx := &ai.ToolContext{
		Root:      root,
		DocTree:   buildDocTree(root),
		HasModels: hasModelsDir(root),
	}
	agent := &ai.AgentCtx{
		Client:        client,
		Tools:         ai.ApplyTools(),
		ToolContext:   toolCtx,
		SystemPrompt:  ai.DefaultSystemPrompt(project, root),
		MaxIterations: 6,
	}
	return agent.Run(ctx)
}

// buildDocTree 生成项目的文档树描述（8 阶段类型 → 文档标题/日期）。
func buildDocTree(root string) string {
	// 依赖 scan：直接扫项目根，构造不同类型 → 文档行
	prj := scanProject(root)
	if prj == nil {
		return ""
	}
	byType := map[string][]map[string]string{}
	for _, d := range prj.Documents {
		byType[d.Type] = append(byType[d.Type], map[string]string{
			"title": d.Title, "date": d.Date,
		})
	}
	var sb strings.Builder
	for _, t := range []string{"requirement", "research", "story", "product", "spec", "roadmap", "plan", "sprint", "doc"} {
		label := map[string]string{
			"requirement": "客户需求", "research": "调研", "story": "用户故事", "product": "产品设计",
			"spec": "功能设计", "roadmap": "路线图", "plan": "开发计划", "sprint": "冲刺", "doc": "文档",
		}[t]
		for _, d := range byType[t] {
			sb.WriteString("- ")
			sb.WriteString(label)
			sb.WriteString(": ")
			sb.WriteString(d["title"])
			if d["date"] != "" {
				sb.WriteString(" (" + d["date"] + ")")
			}
			sb.WriteString("\n")
		}
	}
	return sb.String()
}

// hasModelsDir 判断项目是否有 docs/models 目录。
func hasModelsDir(root string) bool {
	_, err := os.Stat(filepath.Join(root, "docs", "models", "schema.md"))
	return err == nil
}

// scanProject 对单个项目路径执行一次扫描（复用 scan 包）。
func scanProject(root string) *scan.Project {
	return scan.Scan(root, nil, nil)
}

// DebounceAutoMs 自动触发的防抖时长（3 分钟）。
const DebounceAutoMs = 3 * 60 * 1000

// MaybeAutoAnalyse 文档变化后自动触发 AI 分析（防抖 3 分钟；未配置 AI 则跳过）。
// 用于 Watch 检测到项目文档变化时调用。取变化项目（或缺省第一个）跑一次 L3 agent。
func (s *Server) MaybeAutoAnalyse(name string) {
	s.aiAutoMu.Lock()
	defer s.aiAutoMu.Unlock()
	if !s.aiLastRun.IsZero() && time.Since(s.aiLastRun).Milliseconds() < DebounceAutoMs {
		return // 防抖：距上次运行不足 3 分钟，跳过
	}
	cfg, err := ai.LoadConfig(s.authPath)
	if err != nil || cfg.BaseURL == "" || cfg.Model == "" {
		return // 未配置 AI，跳过
	}
	s.aiLastRun = time.Now()

	// 解析项目根
	var project, root string
	if name != "" {
		s.mu.RLock()
		root = s.projectRoots[name]
		s.mu.RUnlock()
		if root == "" {
			project, root = name, ""
		} else {
			project = name
		}
	}
	if root == "" {
		s.mu.RLock()
		for _, p := range s.agg.Projects {
			root, project = p.Root, p.Name
			break
		}
		s.mu.RUnlock()
	}
	if root == "" {
		return
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 600*time.Second)
		defer cancel()
		s.broadcast(`{"type":"ai","event":"running","project":"` + project + `","auto":true}`)
		sugs, aerr := runAgentOnce(ctx, cfg, project, root)
		s.sugMu.Lock()
		s.suggestionsAt = time.Now().Format("15:04:05")
		if aerr != nil {
			s.suggestions = []ai.Suggestion{{
				Type: "ok", Severity: "warning", Title: "AI 分析失败",
				Detail: aerr.Error(), Action: "检查 AI 配置后重试",
			}}
		} else {
			s.suggestions = sugs
		}
		s.sugMu.Unlock()
		s.broadcast(`{"type":"ai","event":"done","project":"` + project + `","auto":true}`)
	}()
}
