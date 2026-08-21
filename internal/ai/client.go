package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Message 对话消息。
type Message struct {
	Role       string     `json:"role"`
	Content    string     `json:"content,omitempty"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
	ToolCallID string     `json:"tool_call_id,omitempty"`
	Name       string     `json:"name,omitempty"`
}

// ToolCall 模型请求的工具调用。
type ToolCall struct {
	ID       string       `json:"id"`
	Type     string       `json:"type"`
	Function ToolCallFunc `json:"function"`
}

// ToolCallFunc 工具调用的名称与参数。
type ToolCallFunc struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
}

// Tool 描述一个可供模型调用的工具。
type Tool struct {
	Type     string     `json:"type"`
	Function ToolSchema `json:"function"`
}

// ToolSchema 工具的 JSON Schema 描述。
type ToolSchema struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

// ChatResp 兼容 OpenAI chat completions 响应。
type ChatResp struct {
	Choices []struct {
		Message Message `json:"message"`
	} `json:"choices"`
}

// Client 泛型 OpenAI 兼容客户端。
type Client struct {
	BaseURL string
	APIKey  string
	Model   string
	HTTP    *http.Client
}

// NewClient 构建设置好的客户端。
func NewClient(cfg Config) *Client {
	base := strings.TrimRight(cfg.BaseURL, "/")
	if !strings.HasSuffix(base, "/chat/completions") {
		base += "/chat/completions"
	}
	return &Client{
		BaseURL: base,
		APIKey:  cfg.APIKey,
		Model:   cfg.Model,
		HTTP:    &http.Client{Timeout: 120 * time.Second},
	}
}

// Chat 发送一轮带工具定义的对话（不执行工具，由调用者负责 agent loop）。
func (c *Client) Chat(ctx context.Context, messages []Message, tools []Tool, maxTokens int) (Message, error) {
	payload := map[string]interface{}{
		"model":    c.Model,
		"messages": messages,
	}
	if len(tools) > 0 {
		payload["tools"] = tools
	}
	if maxTokens > 0 {
		payload["max_tokens"] = maxTokens
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return Message{}, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL, bytes.NewReader(body))
	if err != nil {
		return Message{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	if c.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.APIKey)
	}
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return Message{}, fmt.Errorf("AI 请求失败: %w", err)
	}
	defer resp.Body.Close()
	rb, err := io.ReadAll(resp.Body)
	if err != nil {
		return Message{}, err
	}
	if resp.StatusCode != http.StatusOK {
		return Message{}, fmt.Errorf("AI 返回 %d: %s", resp.StatusCode, truncate(string(rb), 300))
	}
	var cr ChatResp
	if err := json.Unmarshal(rb, &cr); err != nil {
		return Message{}, err
	}
	if len(cr.Choices) == 0 {
		return Message{}, fmt.Errorf("AI 无返回内容")
	}
	return cr.Choices[0].Message, nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
