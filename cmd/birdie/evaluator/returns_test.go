package evaluator

import "testing"

func TestEvalReturns(t *testing.T) {
	tests := []struct{ name, input, want string }{
		{"stop-after-return", `fn answer() int:
    return 42
    value = 1 // 0
answer()`, "42"},
		{"return-through-nested-blocks", `fn answer() int:
    if true:
        if true:
            return 7
        value = 1 // 0
    return 9
answer()`, "7"},
		{"return-stops-at-call-boundary", `fn inner() int:
    return 3
fn outer() int:
    value = inner()
    return value + 4
outer() + 1`, "8"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, d := testEval(t, tt.input)
			requireNoRuntimeDiagnostic(t, tt.input, d)
			testIntegerObject(t, value, tt.want)
		})
	}
}
