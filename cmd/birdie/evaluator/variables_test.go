package evaluator

import (
	"testing"

	"github.com/birdie-lang/birdie/cmd/birdie/diagnostic"
	"github.com/birdie-lang/birdie/cmd/birdie/object"
	"github.com/birdie-lang/birdie/cmd/birdie/source"
)

// Each scenario reuses an environment across evaluations, as the REPL does.
func TestEvalIntegerVariables(t *testing.T) {
	type step struct{ input, want string }
	tests := []struct {
		name  string
		steps []step
	}{
		{"expression-initializer", []step{{"count = 2 + 3 * 4\ncount", "14"}}},
		{"update-from-previous-value", []step{
			{"count = 5\ncount", "5"},
			{"count = count + 3\ncount", "8"},
		}},
		{"reassignment-preserves-copy", []step{
			{"count = 9223372036854775808\ncount", "9223372036854775808"},
			{"original = count\noriginal", "9223372036854775808"},
			{"count = count + 1\ncount", "9223372036854775809"},
			{"original", "9223372036854775808"},
		}},
		{"negation-preserves-operand", []step{
			{"count = 7\ncount", "7"},
			{"-count", "-7"},
			{"count", "7"},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := object.NewEnvironment()
			for _, step := range tt.steps {
				value, d := testEvalWithEnvironment(t, step.input, env)
				requireNoRuntimeDiagnostic(t, step.input, d)
				testIntegerObject(t, value, step.want)
			}
		})
	}
}

func TestEvalBooleanVariables(t *testing.T) {
	tests := []struct {
		name, input string
		want        bool
	}{
		{"false", "ready = false\nready", false},
		{"reassignment-preserves-copy", "ready = 1 < 2\noriginal = ready\nready = false\nready == original", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, d := testEval(t, tt.input)
			requireNoRuntimeDiagnostic(t, tt.input, d)
			testBooleanObject(t, value, tt.want)
		})
	}
}

func TestUnboundVariableDiagnostics(t *testing.T) {
	tests := []struct {
		name, input string
		span        source.Span
	}{
		{"undefined-name", "missing", source.Span{Start: 0, End: 7}},
		{"self-initialization", "value = value", source.Span{Start: 8, End: 13}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testRuntimeDiagnostic(t, tt.input, diagnostic.UnboundVariable, tt.span)
		})
	}
}

func TestFailedAssignmentPreservesEnvironment(t *testing.T) {
	tests := []struct {
		name, input string
		code        diagnostic.Code
		span        source.Span
	}{
		{"new-binding-stops-after-error", "value = missing\nsaved = 0", diagnostic.UnboundVariable, source.Span{Start: 8, End: 15}},
		{"failed-reassignment", "saved = 1 // 0", diagnostic.DivisionByZero, source.Span{Start: 10, End: 12}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := object.NewEnvironment()
			const setup = "saved = 7\nsaved"
			value, d := testEvalWithEnvironment(t, setup, env)
			requireNoRuntimeDiagnostic(t, setup, d)
			testIntegerObject(t, value, "7")

			value, d = testEvalWithEnvironment(t, tt.input, env)
			if d == nil {
				t.Fatal("expected a diagnostic from the failed assignment")
			}
			if value != nil || d.Code != tt.code || d.Span != tt.span {
				t.Errorf("got value %+v and diagnostic %+v, want no value and %s at %+v", value, *d, tt.code, tt.span)
			}
			saved, _ := env.Get("saved")
			testIntegerObject(t, saved, "7")
			if _, exists := env.Get("value"); exists {
				t.Error("failed assignment created a binding")
			}
		})
	}
}
