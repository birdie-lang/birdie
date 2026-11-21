package evaluator

import (
	"strings"
	"testing"

	"github.com/birdie-lang/birdie/cmd/birdie/diagnostic"
	"github.com/birdie-lang/birdie/cmd/birdie/source"
)

func TestEvalFunctions(t *testing.T) {
	tests := []struct{ name, input, want string }{
		{"argument-expressions", `fn add(left: int, right: int) int:
    return left + right
add(1 + 2, 3 * 4)`, "15"},
		{"call-before-declaration", `answer = double(4)
fn double(value: int) int:
    return value * 2
answer`, "8"},
		{"local-parameters-per-call", `value = 10
fn twice(value: int) int:
    value = value + 1
    return value * 2
first = twice(2)
second = twice(4)
value + first + second`, "26"},
		{"lexical-not-caller-scope", `value = 10
fn read() int:
    return value
fn caller(value: int) int:
    return read()
caller(20)`, "10"},
		{"factorial", `fn factorial(n: int) int:
    if n == 0:
        return 1
    return n * factorial(n - 1)
factorial(5)`, "120"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, d := testEval(t, tt.input)
			requireNoRuntimeDiagnostic(t, tt.input, d)
			testIntegerObject(t, value, tt.want)
		})
	}
}

func TestEvalFunctionDiagnostics(t *testing.T) {
	for _, input := range []string{
		"fn divide(value: int) int:\n    return 10 // value\ndivide(0)",
		"fn add(left: int, right: int) int:\n    return left + right\nadd(1 // 0, 2 // 0)",
	} {
		t.Run(input, func(t *testing.T) {
			start := strings.Index(input, "//")
			testRuntimeDiagnostic(t, input, diagnostic.DivisionByZero, source.Span{Start: start, End: start + 2})
		})
	}
}

func TestEvalInvalidCalls(t *testing.T) {
	input := "fn identity(value: int) int:\n    return value\nidentity()"
	value, d := testEval(t, input)
	if d == nil {
		t.Fatal("expected a diagnostic for the missing argument")
	}
	if value != nil || d.Code == "" || d.Message == "" {
		t.Errorf("got value %+v and diagnostic %+v, want no value and a code/message", value, *d)
	}
	// The error should identify the call; its exact code is still undecided.
	callStart := strings.LastIndex(input, "\n") + 1
	if d.Span.Start < callStart || d.Span.End > len(input) || d.Span.Start >= d.Span.End {
		t.Errorf("diagnostic span %+v is outside the failing call", d.Span)
	}
}
