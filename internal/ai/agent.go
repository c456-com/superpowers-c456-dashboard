package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// Suggestion 一条 AI 建议。
type Suggestion struct {
	ID       string `json:"id"`
	Type     string `json:"type"`     // agents_missing / agents_rule / doc_missing / models_missing / ok
	Severity string `json:"severity"` // info / warning / success
	Title    string `json:"title"`
	Detail   string `json:"detail"`
	Action   string `json:"action"` // 建议执行的动作描述
}

// AgentCtx L3 agent 的一次运行上下文。
type AgentCtx struct {
	Client        *Client
	Tools         []Tool
	ToolContext   *ToolContext
	SystemPrompt  string
	MaxIterations int
}

// Run 执行 L3 agent loop：模型思考→调工具→拿结果→再思考，直到产出 JSON 建议。
// 返回建议列表。
func (a *AgentCtx) Run(ctx context.Context) ([]Suggestion, error) {
	messages := []Message{
		{Role: "system", Content: a.SystemPrompt},
	}

	history := []string{} // 工具调用历史（用于提示模型已探索内容）
	for i := 0; i < a.MaxIterations; i++ {
		// 累积工具探索结果喂给模型（轻量上下文合并，避免无限增长）
		user := "请基于项目情况，用工具探查后给出建议。已探索结果:\n" + strings.Join(history, "\n")
		if i > 0 {
			messages = append(messages, Message{Role: "user", Content: user})
		}
		msg, err := a.Client.Chat(ctx, messages, a.Tools, 1024)
		if err != nil {
			return nil, err
		}
		// 追加模型回复
		messages = append(messages, msg)

		// 模型要求调用工具
		if len(msg.ToolCalls) > 0 {
			for _, tc := range msg.ToolCalls {
				args := tc.Function.Arguments
				res, err := a.ToolContext.ExecuteOne(ctx, tc.Function.Name, args)
				if err != nil {
					res = "工具错误: " + err.Error()
				}
				history = append(history, fmt.Sprintf("[%s] %s", tc.Function.Name, truncate(res, 800)))
				messages = append(messages, Message{
					Role:       "tool",
					ToolCallID: tc.ID,
					Content:    res,
				})
			}
			continue
		}

		// 无工具调用 → 视为最终建议，尝试解析 JSON
		if s, ok := parseSuggestions(msg.Content); ok {
			return s, nil
		}
		// 无法解析则继续一轮（给模型一次重试机会）
		if i == a.MaxIterations-1 {
			return defaultSuggestion(msg.Content), nil
		}
		messages = append(messages, Message{Role: "user", Content: "请直接输出 JSON 数组建议，不要用其他格式。"})
	}
	return nil, fmt.Errorf("agent 达到最大迭代次数未产出建议")
}

// parseSuggestions 尝试从模型输出解析 JSON 建议数组（容忍 ```json 围栏）。
func parseSuggestions(s string) ([]Suggestion, bool) {
	s = strings.TrimSpace(s)
	// 去掉 ```json ... ``` 围栏
	if strings.HasPrefix(s, "```") {
		if i := strings.Index(s, "\n"); i > 0 {
			s = s[i+1:]
		}
		if i := strings.LastIndex(s, "```"); i > 0 {
			s = s[:i]
		}
	}
	// 找到第一个 [ 到最后一个 ]
	start := strings.Index(s, "[")
	end := strings.LastIndex(s, "]")
	if start < 0 || end <= start {
		return nil, false
	}
	var list []Suggestion
	if err := json.Unmarshal([]byte(s[start:end+1]), &list); err != nil {
		return nil, false
	}
	return list, true
}

func defaultSuggestion(content string) []Suggestion {
	return []Suggestion{{
		Type: "ok", Severity: "info", Title: "AI 分析完成",
		Detail: content, Action: "",
	}}
}

// DefaultSystemPrompt 构建 agent 系统提示（描述工具用法与输出格式）。
func DefaultSystemPrompt(project, root string) string {
	return fmt.Sprintf(`你是 superpowers 开发最佳实践顾问。请分析项目「%s」（根目录 %s），
通过工具了解项目情况后，给出改进建议。

可用的工具：list_docs（列某类文档）、read_agents（读 AGENTS.md）、check_models（检查数据模型目录）。

检查要点：
1. 项目是否缺失 AGENTS.md，或 AGENTS.md 缺开发最佳实践规则（8阶段工作流/数据模型同步/验证驱动）→ 建议补充
2. 工作流各阶段（客户需求→调研→用户故事→产品设计→功能设计→路线图→开发计划→冲刺）是否有缺口 → 建议创建缺的文档类型
3. 是否有功能设计(spec) 但缺 docs/models 数据模型目录 → 建议同步数据模型

先调用工具探索，最后输出 JSON 数组，每个元素字段：
{"type": "agents_missing|agents_rule|doc_missing|models_missing|ok", "severity": "info|warning|success", "title": "简短标题", "detail": "具体说明", "action": "建议执行的行动（如：创建文档 / 注入规则）"}
只输出 JSON 数组，不要其他文字。`, project, root)
}
