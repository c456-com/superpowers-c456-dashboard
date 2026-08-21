package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ToolContext 提供给工具的执行上下文（项目根 + 已扫描结果）。
type ToolContext struct {
	Root      string // 项目根目录
	DocTree   string // 简化文档树描述（8 阶段类型 → 文档列表）
	HasModels bool   // 是否已有 docs/models
	// 模板内容（供建议注入）
	agentsTemplate string
}

// ExecuteAnalyzers 内置工具集定义（模型可调用）。
func ApplyTools() []Tool {
	tools := []Tool{
		{
			Type: "function",
			Function: ToolSchema{
				Name:        "list_docs",
				Description: "列出项目某个文档类型的文档（类型：requirement/research/story/product/spec/roadmap/plan/sprint/doc）。调用后你会知道该阶段有哪些文档、各自标题与日期。",
				Parameters: map[string]interface{}{
					"type": map[string]interface{}{"type": "string", "description": "文档类型"},
				},
			},
		},
		{
			Type: "function",
			Function: ToolSchema{
				Name:        "read_agents",
				Description: "读取项目的 AGENTS.md 内容（若存在）。用于判断是否已有开发最佳实践规则。",
				Parameters:  map[string]interface{}{},
			},
		},
		{
			Type: "function",
			Function: ToolSchema{
				Name:        "check_models",
				Description: "检查项目是否有 data/models 数据模型目录及 schema.md。返回是否存在。",
				Parameters:  map[string]interface{}{},
			},
		},
	}
	return tools
}

// 8 阶段类型中文标签（给模型提示用）。
var typeLabels = map[string]string{
	"requirement": "客户需求", "research": "调研", "story": "用户故事", "product": "产品设计",
	"spec": "功能设计", "roadmap": "路线图", "plan": "开发计划", "sprint": "冲刺", "doc": "文档",
}

// ExecuteOne 执行单个工具调用，返回结果字符串。
func (tc *ToolContext) ExecuteOne(ctx context.Context, name string, args json.RawMessage) (string, error) {
	switch name {
	case "list_docs":
		var a struct {
			Type string `json:"type"`
		}
		_ = json.Unmarshal(args, &a)
		return tc.listDocs(a.Type), nil
	case "read_agents":
		return tc.readAgents(), nil
	case "check_models":
		b := tc.HasModels
		if b {
			return "docs/models 存在，已有数据模型目录。", nil
		}
		return "项目缺少 docs/models 数据模型目录。", nil
	default:
		return "", fmt.Errorf("未知工具: %s", name)
	}
}

func (tc *ToolContext) listDocs(t string) string {
	if t == "" {
		return "请指定类型参数 type。类型: requirement/research/story/product/spec/roadmap/plan/sprint/doc"
	}
	found := []string{}
	if tc.DocTree != "" {
		for _, line := range strings.Split(tc.DocTree, "\n") {
			if strings.Contains(line, typeLabels[t]) {
				found = append(found, line)
			}
		}
	}
	if len(found) == 0 {
		return fmt.Sprintf("类型「%s」暂无文档。", typeLabels[t])
	}
	return fmt.Sprintf("类型「%s」文档:\n%s", typeLabels[t], strings.Join(found, "\n"))
}

func (tc *ToolContext) readAgents() string {
	p := filepath.Join(tc.Root, "AGENTS.md")
	b, err := os.ReadFile(p)
	if err != nil {
		return "项目无 AGENTS.md。"
	}
	return "项目 AGENTS.md 存在，前 1200 字:\n" + truncate(string(b), 1200)
}
