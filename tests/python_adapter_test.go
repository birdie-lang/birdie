package tests

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestPythonNodeAdapter(t *testing.T) {
	python := os.Getenv("BIRDIE_TEST_PYTHON")
	if python == "" {
		t.Skip("set BIRDIE_TEST_PYTHON to check the Python node adapter")
	}
	python, err := exec.LookPath(python)
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name, body, input, wantError string
	}{
		{"existing-reference", "return p if p is root.left else root", `[[2,1,3],1]`, ""},
		{"fresh-copy", "return TreeNode(p.val)", `[[2,1,3],1]`, "original input node"},
		{"missing-target", "return p", `[[2,1,3],4]`, "node selector not found"},
		{"ambiguous-values", "return p", `[[2,1,1],1]`, "unique integer values"},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			source := filepath.Join(t.TempDir(), "solution.py")
			code := "class Solution:\n    def select(self, root, p):\n        " + test.body + "\n"
			if err := os.WriteFile(source, []byte(code), 0o644); err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
			defer cancel()
			cmd := exec.CommandContext(ctx, python, filepath.Join("testdata", "python_runner.py"), source)
			cmd.Env = append(os.Environ(), "PYTHONDONTWRITEBYTECODE=1", "PYTHONUTF8=1")
			cmd.Stdin = strings.NewReader(`{"entry":"select","argAdapters":["tree","node"],"resultAdapter":"node","args":` + test.input + `}`)
			output, err := cmd.CombinedOutput()
			if test.wantError != "" {
				if err == nil || !strings.Contains(string(output), test.wantError) {
					t.Fatalf("expected %q; error=%v; output=%s", test.wantError, err, output)
				}
			} else if err != nil || strings.TrimSpace(string(output)) != "1" {
				t.Fatalf("expected the existing left child; error=%v; output=%s", err, output)
			}
		})
	}
}
