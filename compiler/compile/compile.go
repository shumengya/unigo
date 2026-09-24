package compile

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"garnet/compiler/check"
	"garnet/compiler/gogen"
	"garnet/compiler/parser"
)

func Build(path string) error {
	src, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return BuildSource(path, src)
}

func BuildSource(filename string, src []byte) error {
	file, errs := parser.Parse(filename, src)
	if err := errs.Err(); err != nil {
		return err
	}
	info, errs := check.Check(file)
	if err := errs.Err(); err != nil {
		return err
	}
	code, err := gogen.Generate(file, info)
	if err != nil {
		return err
	}
	return writeAndBuild(filename, code)
}

func writeAndBuild(srcPath string, code []byte) error {
	outDir := filepath.Join(filepath.Dir(srcPath), "garnet-out")
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return err
	}
	entries, err := os.ReadDir(outDir)
	if err != nil {
		return err
	}
	for _, e := range entries {
		name := e.Name()
		if strings.HasSuffix(name, ".go") || name == "go.mod" || strings.HasSuffix(name, ".exe") {
			if err := os.Remove(filepath.Join(outDir, name)); err != nil {
				return err
			}
		}
	}
	base := strings.TrimSuffix(filepath.Base(srcPath), filepath.Ext(srcPath))
	if err := os.WriteFile(filepath.Join(outDir, base+".go"), code, 0o644); err != nil {
		return err
	}
	mod := []byte("module garnetout\n\ngo 1.22\n")
	if err := os.WriteFile(filepath.Join(outDir, "go.mod"), mod, 0o644); err != nil {
		return err
	}
	bin := base
	if runtime.GOOS == "windows" {
		bin += ".exe"
	}
	cmd := exec.Command("go", "build", "-o", bin, ".")
	cmd.Dir = outDir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("go build 失败: %w\n%s", err, out)
	}
	return nil
}
