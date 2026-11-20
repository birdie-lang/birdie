package repl

import (
	"bytes"
	"strings"
	"testing"
)

func TestParseLineDiagnostics(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{"lexer", "@", []string{"<repl>:1:1: error[unexpected-character]"}},
		{"parser", "x =", []string{"<repl>:1:4: error[expected-expression]"}},
		{"indentation", "  x = 1", []string{"<repl>:1:1: error[invalid-indentation]"}},
		{"source-order", "x =\n@", []string{
			"<repl>:1:4: error[expected-expression]",
			"<repl>:2:1: error[unexpected-character]",
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out bytes.Buffer
			program, ok := parseLine(tt.input, &out)
			if ok || program != nil {
				t.Fatal("invalid input was reported as a successful parse")
			}
			output := out.String()
			if strings.Count(output, "error[") != len(tt.want) {
				t.Fatalf("unexpected diagnostic count in output:\n%s", output)
			}
			previous := -1
			for _, want := range tt.want {
				position := strings.Index(output, want)
				if position <= previous {
					t.Fatalf("missing or out-of-order %q in output:\n%s", want, output)
				}
				previous = position
			}
		})
	}
}

func TestParseLineSuccess(t *testing.T) {
	var out bytes.Buffer
	program, ok := parseLine("x = 1\ny = 2\n", &out)
	if !ok || program == nil || len(program.Statements) != 2 || out.Len() != 0 {
		t.Fatalf("valid input failed: program=%+v, ok=%t, output=%q", program, ok, out.String())
	}
}

func TestDiagnosticModes(t *testing.T) {
	var out bytes.Buffer
	Start(strings.NewReader("/parse\nx =\n/tokens\n@\n"), &out)
	for _, code := range []string{"expected-expression", "unexpected-character"} {
		if strings.Count(out.String(), "error["+code+"]") != 1 {
			t.Errorf("expected one %s diagnostic in output:\n%s", code, out.String())
		}
	}
}
