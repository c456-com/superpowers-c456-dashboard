package ai

import (
	"context"
	"testing"
)

func TestExecuteOneArgsShapes(t *testing.T) {
	tc := &ToolContext{Root: "/nonexist", DocTree: "- 功能设计: 搜索（立即解析不受影响）\n- 功能设计: 资产导入\n"}

	cases := []struct{ name, args, want string }{
		{"object形式", `{"type":"spec"}`, "功能设计"},
		{"字符串重编码", `"{\"type\":\"spec\"}"`, "功能设计"},
	}
	for _, c := range cases {
		got, err := tc.ExecuteOne(context.Background(), "list_docs", []byte(c.args))
		if err != nil {
			t.Fatalf("%s err: %v", c.name, err)
		}
		if !contains(got, c.want) {
			t.Errorf("%s: 得 %q, 期望含 %q", c.name, got, c.want)
		}
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || indexOf(s, sub) >= 0)
}
func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
