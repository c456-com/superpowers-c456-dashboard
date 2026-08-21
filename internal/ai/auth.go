// Package ai 提供 AI 建议引擎：配置管理 + 通用 OpenAI 兼容客户端 + L3 agent loop。
package ai

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Config AI 模型配置（存全局配置目录 auth.json，600 权限，不随项目库）。
type Config struct {
	BaseURL string `json:"base_url"` // OpenAI 兼容端点基址，如 http://127.0.0.1:8888/v1
	Model   string `json:"model"`    // 模型名
	APIKey  string `json:"api_key"`  // API Key（本地端点可空）
}

// 已配置文件标记（避免 dashboard 把读取到的值误当敏感外泄）
type stored struct {
	AI Config `json:"ai"`
}

// DefaultAuthPath 返回全局配置目录下的 auth.json。
func DefaultAuthPath() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		dir = ".config"
	}
	return filepath.Join(dir, "superpowers-c456-dashboard", "auth.json")
}

// LoadConfig 读取 AI 配置；不存在返回零值（未配置）。
func LoadConfig(path string) (Config, error) {
	var s stored
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Config{}, nil
		}
		return Config{}, err
	}
	if err := json.Unmarshal(b, &s); err != nil {
		return Config{}, err
	}
	return s.AI, nil
}

// SaveConfig 保存 AI 配置（0600 权限，防读取 key）。
func SaveConfig(path string, cfg Config) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	s := stored{AI: cfg}
	b, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o600)
}
