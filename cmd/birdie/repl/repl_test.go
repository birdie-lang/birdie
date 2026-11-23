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
		{"indentation", "  x = 1", []string{"error[invalid-indentation]", " --> <repl>:1:1"}},
		{"source-order", "x =\n@", []string{
			"error[expected-expression]", " --> <repl>:1:4",
			"error[unexpected-character]", " --> <repl>:2:1",
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
			if strings.Count(output, "error[") != len(tt.want)/2 {
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

func TestEvaluationModes(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "default-mode",
			input: "1 + 2 * 3\ntrue\n1 > 2\n",
			want:  "> 7\n> true\n> false\n> ",
		},
		{
			name:  "switch-through-all-modes",
			input: "/tokens\n/parse\n1 + 2\n/eval\n1 + 2\n",
			want:  "> tokens> parse> (1 + 2)\nparse> > 3\n> ",
		},
		{
			name:  "empty-line",
			input: "\n1 + 2\n",
			want:  "> > 3\n> ",
		},
		{
			name:  "list-indexing",
			input: "[10][0]\nx = [10]\nx\nx[0]\n",
			want:  "> 10\n> > [10]\n> 10\n> ",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out bytes.Buffer
			Start(strings.NewReader(tt.input), &out)
			if got := out.String(); got != tt.want {
				t.Fatalf("output = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestEvaluationDiagnostics(t *testing.T) {
	tests := []struct {
		name  string
		input string
		code  string
	}{
		{"lexer", "5 + true @\n2 + 3\n", "unexpected-character"},
		{"parser", "x =\n2 + 3\n", "expected-expression"},
		{"runtime", "5 + true\n2 + 3\n", "unknown-operator"},
		{"out-of-bounds-list-index", "[10][40]\n2 + 3\n", "index-out-of-bounds"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out bytes.Buffer
			Start(strings.NewReader(tt.input), &out)
			output := out.String()
			if strings.Count(output, "error[") != 1 || !strings.Contains(output, "error["+tt.code+"]") {
				t.Fatalf("expected one %s diagnostic:\n%s", tt.code, output)
			}
			if !strings.Contains(output, " --> <repl>:1:") {
				t.Fatalf("diagnostic is missing its source location:\n%s", output)
			}
			if !strings.HasSuffix(output, "> 5\n> ") {
				t.Fatalf("REPL did not evaluate the next input after the error:\n%s", output)
			}
		})
	}
}

func TestHelp(t *testing.T) {
	var out bytes.Buffer
	handled, m := handleCommand("/help", modeParse, &out)
	if !handled || m != modeParse {
		t.Fatal("help changed the current mode")
	}
	output := out.String()
	previous := -1
	for _, label := range []string{"Commands:", "/eval", "/tokens", "/parse", "/help", "/exit, /quit", "Editing:", "Enter", "Empty line", "Tab", "Shift+Tab", "Up / Down", "Ctrl+C", "Ctrl+D"} {
		pos := strings.Index(output, label)
		if pos <= previous {
			t.Fatalf("missing or out-of-order %q in help:\n%s", label, output)
		}
		previous = pos
	}
	for _, unsupported := range []string{"/compile", "/run"} {
		if strings.Contains(output, unsupported) {
			t.Fatalf("help advertises unsupported command %s", unsupported)
		}
	}
}

func TestUnknownCommands(t *testing.T) {
	for _, command := range []string{"/foo", "/compile", "/run", "/foo(", "/foo:", "/tokens 1 + 2"} {
		t.Run(command, func(t *testing.T) {
			var out bytes.Buffer
			Start(strings.NewReader("/parse\n"+command+"\n1 + 2\n"), &out)
			want := "> parse> Unknown command \"" + command + "\". Use /help for commands.\nparse> (1 + 2)\nparse> "
			if got := out.String(); got != want {
				t.Fatalf("output = %q, want %q", got, want)
			}
		})
	}
}
