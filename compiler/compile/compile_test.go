package compile

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGarnetExample(t *testing.T) {
	srcPath := filepath.Join("..", "..", "examples", "tour", "tour.gn")
	if err := Build(srcPath); err != nil {
		t.Fatal(err)
	}
	outPath := filepath.Join("..", "..", "examples", "tour", "garnet-out", "tour.go")
	src, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatal(err)
	}
	text := string(src)
	if strings.Contains(text, ":=") {
		t.Fatal("生成代码使用了 :=")
	}
	for _, want := range []string{
		"[]byte(",
		"`json:\"id\"`",
		"Status_ready",
		"else if",
		"garnetBindStudentParser",
		"type Celsius float64",
		"type Kelvin = float64",
		"png.BestCompression",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("生成代码缺少 %s", want)
		}
	}
	if _, err := parser.ParseFile(token.NewFileSet(), outPath, src, 0); err != nil {
		t.Fatal(err)
	}
}

func TestFailFixtures(t *testing.T) {
	dir := filepath.Join("..", "..", "tests", "fail")
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) == 0 {
		t.Fatal("tests/fail 是空的")
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".gn") {
			continue
		}
		t.Run(e.Name(), func(t *testing.T) {
			path := filepath.Join(dir, e.Name())
			src, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			line, _, _ := strings.Cut(string(src), "\n")
			line = strings.TrimPrefix(strings.TrimSpace(line), "// error:")
			want := strings.TrimSpace(line)
			if want == "" {
				t.Fatal("缺少 // error: 说明")
			}
			err = BuildSource(filepath.Join(t.TempDir(), e.Name()), src)
			if err == nil {
				t.Fatal("期望编译失败")
			}
			if !strings.Contains(err.Error(), want) {
				t.Fatalf("错误里没有 %q\n%s", want, err)
			}
		})
	}
}

// tests/pass 和 examples 下的每个 .gn 都必须能编译通过。
func TestPassFixtures(t *testing.T) {
	var paths []string
	for _, pattern := range []string{
		filepath.Join("..", "..", "tests", "pass", "*.gn"),
		filepath.Join("..", "..", "examples", "*", "*.gn"),
	} {
		found, err := filepath.Glob(pattern)
		if err != nil {
			t.Fatal(err)
		}
		paths = append(paths, found...)
	}
	if len(paths) == 0 {
		t.Fatal("没有找到 tests/pass 或 examples 下的 .gn 文件")
	}
	for _, path := range paths {
		t.Run(filepath.Base(path), func(t *testing.T) {
			src, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if err := BuildSource(filepath.Join(t.TempDir(), filepath.Base(path)), src); err != nil {
				t.Fatal(err)
			}
		})
	}
}
