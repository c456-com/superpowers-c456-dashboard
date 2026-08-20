package server

import (
	"encoding/json"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"superpowers-c456-dashboard/internal/aggregate"
)

// 构建一个带两个候选解析的文件服务，验证 dir（文档目录）解析 + 防目录穿越。
func newTestServer(t *testing.T) *Server {
	t.Helper()
	root := t.TempDir()
	// 项目根下放：根相对文件 + 文档目录下的文件
	if err := os.MkdirAll(filepath.Join(root, "docs"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "app.rb"), []byte("root-file"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "docs", "note.md"), []byte("doc-file"), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg := &aggregate.Config{
		Projects: []aggregate.ProjectSpec{{Name: "proj", Path: root}},
	}
	s := &Server{projectRoots: map[string]string{"proj": root}, agg: aggregate.ScanAll(cfg), clients: map[chan string]struct{}{}}
	return s
}

func doReq(s *Server, url string) (int, map[string]interface{}) {
	r := httptest.NewRequest("GET", url, nil)
	w := httptest.NewRecorder()
	s.fileHandler(w, r)
	var m map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &m)
	return w.Code, m
}

func TestFileHandlerRootRel(t *testing.T) {
	s := newTestServer(t)
	code, m := doReq(s, "/api/file?project=proj&path=app.rb")
	if code != 200 || !strings.Contains(m["content"].(string), "root-file") {
		t.Fatalf("根相对读取失败 code=%d m=%v", code, m)
	}
}

func TestFileHandlerDocDir(t *testing.T) {
	s := newTestServer(t)
	// dir=docs 且文件在 docs/note.md → 应成功
	code, m := doReq(s, "/api/file?project=proj&path=note.md&dir=docs")
	if code != 200 || !strings.Contains(m["content"].(string), "doc-file") {
		t.Fatalf("文档目录解析失败 code=%d m=%v", code, m)
	}
	// dir=docs 但文件在根 → 自动落到根版本
	code2, m2 := doReq(s, "/api/file?project=proj&path=app.rb&dir=docs")
	if code2 != 200 || !strings.Contains(m2["content"].(string), "root-file") {
		t.Fatalf("dir回退失败 code=%d m=%v", code2, m2)
	}
}

func TestFileHandlerTraversal(t *testing.T) {
	s := newTestServer(t)
	code, _ := doReq(s, "/api/file?project=proj&path=passwd")
	if code != 404 {
		t.Fatalf("文件不存在应404，得 %d", code)
	}
	// 目录穿越越界 → 400
	code2, m2 := doReq(s, "/api/file?project=proj&path=../../../etc/passwd&dir=..")
	if code2 != 400 {
		t.Fatalf("越界应400，得 %d m=%v", code2, m2)
	}
}

func TestFileHandlerUnknownProject(t *testing.T) {
	s := newTestServer(t)
	code, _ := doReq(s, "/api/file?project=nope&path=x")
	if code != 404 {
		t.Fatalf("未知项目应404，得 %d", code)
	}
}
