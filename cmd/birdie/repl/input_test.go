package repl

import (
	"bytes"
	"errors"
	"io"
	"slices"
	"strings"
	"testing"
	"testing/iotest"

	"github.com/reeflective/readline"
)

func TestInputComplete(t *testing.T) {
	tests := []struct {
		name string
		text string
		want bool
	}{
		{"empty", "", true},
		{"expression", "1 + 2", true},
		{"invalid-expression", "1 +", true},
		{"open-parenthesis", "(1 +", false},
		{"closed-parenthesis", "(1 +\n    2)", true},
		{"nested-delimiters", "[1,\n    (2", false},
		{"closed-nested-delimiters", "[1,\n    (2)]", true},
		{"braces", "{1:\n    2}", true},
		{"quoted-delimiters", "text = \"[(:\"", true},
		{"mismatched-delimiters", "([)]", true},
		{"unmatched-close-before-open", ") (", true},
		{"lexical-error-inside-delimiters", "(@", true},
		{"block-header", "fn add(a: int, b: int) int:", false},
		{"block-body", "if ready:\n    1", false},
		{"block-empty-line", "if ready:\n    1\n", true},
		{"block-whitespace-line", "if ready:\n    1\n    ", true},
		{"invalid-block-indentation", "if ready:\n  1", true},
		{"nested-block", "if ready:\n    if found:\n        1", false},
		{"block-with-open-list", "if ready:\n    values = [\n", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := inputComplete(tt.text); got != tt.want {
				t.Errorf("inputComplete(%q) = %t, want %t", tt.text, got, tt.want)
			}
		})
	}
}

func TestStreamInput(t *testing.T) {
	var out bytes.Buffer
	read := streamInput(strings.NewReader("(\n    1 +\n    2\n)\n3\n"), &out)
	for _, want := range []string{"(\n    1 +\n    2\n)", "3"} {
		got, err := read(prompt)
		if err != nil || got != want {
			t.Fatalf("read() = %q, %v; want %q, nil", got, err, want)
		}
	}
	if got, err := read(prompt); got != "" || !errors.Is(err, io.EOF) {
		t.Fatalf("final read() = %q, %v; want EOF", got, err)
	}
	if got, want := out.String(), "> | | | > > "; got != want {
		t.Errorf("prompts = %q, want %q", got, want)
	}
}

func TestBlockInput(t *testing.T) {
	var out bytes.Buffer
	read := streamInput(strings.NewReader("if ready:\n    1\n\n3\n"), &out)
	for _, want := range []string{"if ready:\n    1\n", "3"} {
		got, err := read(prompt)
		if err != nil || got != want {
			t.Fatalf("read() = %q, %v; want %q, nil", got, err, want)
		}
	}
	if got, want := out.String(), "> | | > "; got != want {
		t.Errorf("prompts = %q, want %q", got, want)
	}
}

func TestMultilineTokenMode(t *testing.T) {
	var out bytes.Buffer
	Start(strings.NewReader("/tokens\n(\n    1,\n    2\n)\n"), &out)
	output := out.String()
	if strings.Contains(output, "error[") || strings.Count(output, continuationPrompt) != 3 {
		t.Fatalf("multiline token input failed:\n%s", output)
	}
	if strings.Contains(output, "Type:INDENT") || strings.Contains(output, "Type:DEDENT") {
		t.Fatalf("continuation indentation became block tokens:\n%s", output)
	}
}

func TestCommandsInsideMultilineInput(t *testing.T) {
	var out bytes.Buffer
	Start(strings.NewReader("/tokens\n(\n/eval\n)\n2\n"), &out)
	output := out.String()
	if strings.Contains(output, "Unknown command") || strings.Contains(output, "error[") || !strings.Contains(output, "Type:INT Literal:2") {
		t.Fatalf("a continuation line was interpreted as a REPL command:\n%s", output)
	}
}

func TestMultilineLexicalDiagnostic(t *testing.T) {
	var out bytes.Buffer
	Start(strings.NewReader("/tokens\n(\n    @\n1\n"), &out)
	output := out.String()
	if strings.Count(output, "error[") != 1 || !strings.Contains(output, "error[unexpected-character]") || !strings.Contains(output, " --> <repl>:2:5") {
		t.Fatalf("missing diagnostic for the continuation line:\n%s", output)
	}
	if !strings.Contains(output, "Type:INT Literal:1") {
		t.Fatalf("input was not reset after the lexical error:\n%s", output)
	}
}

func TestIncompleteInputEOF(t *testing.T) {
	var out bytes.Buffer
	Start(strings.NewReader("/tokens\n(\n1"), &out)
	output := out.String()
	if !strings.Contains(output, "error[unexpected-eof]") || !strings.Contains(output, " --> <repl>:2:2") {
		t.Fatalf("missing EOF diagnostic and source location:\n%s", output)
	}
	if strings.Contains(output, "Type:(") {
		t.Fatalf("incomplete input was submitted:\n%s", output)
	}
}

func TestEOFClosesBlockInput(t *testing.T) {
	var out bytes.Buffer
	Start(strings.NewReader("/tokens\nif ready:\n    1"), &out)
	output := out.String()
	if strings.Contains(output, "error[") || !strings.Contains(output, "Type:INDENT") || !strings.Contains(output, "Type:DEDENT") {
		t.Fatalf("EOF did not finish the block entry:\n%s", output)
	}
}

func TestInputError(t *testing.T) {
	var out bytes.Buffer
	Start(iotest.ErrReader(errors.New("read failed")), &out)
	if output := out.String(); !strings.Contains(output, "input error: read failed") {
		t.Fatalf("input error was not reported:\n%s", output)
	}
}

func TestInterruptPreservesMode(t *testing.T) {
	var out bytes.Buffer
	calls := 0
	var prompts []string
	run(func(prompt string) (string, error) {
		prompts = append(prompts, prompt)
		calls++
		switch calls {
		case 1:
			return "/parse", nil
		case 2:
			return "(", readline.ErrInterrupt
		case 3:
			return "1 + 2", nil
		default:
			return "", io.EOF
		}
	}, &out)
	if got, want := out.String(), "(1 + 2)\n"; got != want {
		t.Fatalf("output after cancellation = %q, want %q", got, want)
	}
	if want := []string{"> ", "parse> ", "parse> ", "parse> "}; !slices.Equal(prompts, want) {
		t.Fatalf("prompts after cancellation = %q, want %q", prompts, want)
	}
}

func TestExitCommands(t *testing.T) {
	for _, command := range []string{"/exit", "/quit"} {
		t.Run(command, func(t *testing.T) {
			var out bytes.Buffer
			Start(strings.NewReader(command+"\n999\n"), &out)
			if got, want := out.String(), "> Goodbye!\n"; got != want {
				t.Fatalf("output = %q, want %q", got, want)
			}
		})
	}
}
