package evaluator

import (
	"strings"
	"testing"

	"github.com/birdie-lang/birdie/cmd/birdie/diagnostic"
	"github.com/birdie-lang/birdie/cmd/birdie/source"
)

func TestEvalConditionals(t *testing.T) {
	tests := []struct{ name, input, want string }{
		{"true-branch-shares-scope", `if true:
    value = 4
value = value + 1
value`, "5"},
		{"false-branch-is-skipped", `value = 3
if false:
    value = 1 // 0
value`, "3"},
		{"else-branch", `if false:
    value = 1
else:
    value = 2
value`, "2"},
		{"first-matching-else-if", `if false:
    value = 1
else if 2 < 3:
    value = 2
else if 1 // 0 == 0:
    value = 3
else:
    value = 4
value`, "2"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, d := testEval(t, tt.input)
			requireNoRuntimeDiagnostic(t, tt.input, d)
			testIntegerObject(t, value, tt.want)
		})
	}
}

func TestEvalConditionalDiagnostics(t *testing.T) {
	for _, input := range []string{
		"if 1 // 0 == 0:\n    value = 1",
		"if true:\n    value = 1 // 0\nvalue = 2",
	} {
		t.Run(input, func(t *testing.T) {
			start := strings.Index(input, "//")
			testRuntimeDiagnostic(t, input, diagnostic.DivisionByZero, source.Span{Start: start, End: start + 2})
		})
	}
}
