package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunFile(t *testing.T) {
	const program = `fn factorial(n: int) int:
    if n == 0:
        return 1
    return n * factorial(n - 1)

print(factorial(5))
42
`
	path := filepath.Join(t.TempDir(), "factorial.bd")
	if err := os.WriteFile(path, []byte(program), 0o600); err != nil {
		t.Fatal(err)
	}

	var out, errOut bytes.Buffer
	code := Execute([]string{"run", path}, strings.NewReader(""), &out, &errOut)
	if code != 0 || out.String() != "120\n" || errOut.Len() != 0 {
		t.Fatalf("got exit %d, stdout %q, stderr %q", code, &out, &errOut)
	}
}

func TestRunStdin(t *testing.T) {
	tests := []struct{ name, input, output, diagnostic string }{
		{"silent-result", "2 + 3", "", ""},
		{"parse-error-before-output", "print(1)\nx =\n", "", "error[expected-expression]"},
		{"runtime-error-after-output", "print(1)\n1 // 0", "1\n", "error[division-by-zero]"},
		{"print-arity", "print()", "", "error[argument-count-mismatch]"},
		{"no-value-in-expression", "print(1) + 2", "1\n", "error[expected-value]"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out, errOut bytes.Buffer
			code := Execute([]string{"run", "-"}, strings.NewReader(tt.input), &out, &errOut)
			if out.String() != tt.output {
				t.Errorf("stdout = %q, want %q", &out, tt.output)
			}
			if tt.diagnostic == "" {
				if code != 0 || errOut.Len() != 0 {
					t.Fatalf("got exit %d and stderr %q, want success", code, &errOut)
				}
			} else if code != 1 || !strings.Contains(errOut.String(), tt.diagnostic) ||
				!strings.Contains(errOut.String(), "<stdin>:") || !strings.Contains(errOut.String(), "^") {
				t.Fatalf("got exit %d and stderr %q, want a source diagnostic for %s", code, &errOut, tt.diagnostic)
			}
			if strings.Contains(errOut.String(), "\x1b[") {
				t.Errorf("stderr contains ANSI escapes without a terminal: %q", &errOut)
			}
		})
	}
}
