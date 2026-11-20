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
		{"binding", []step{{"count = 5\ncount", "5"}}},
		{"zero", []step{{"count = 0\ncount", "0"}}},
		{"negative", []step{{"count = -7\ncount", "-7"}}},
		{"expression-initializer", []step{{"count = 2 + 3 * 4\ncount", "14"}}},
		{"large-integer", []step{{"count = 9223372036854775808\ncount", "9223372036854775808"}}},
		{"derived-binding", []step{
			{"count = 5\ncount", "5"},
			{"double = count * 2\ndouble", "10"},
		}},
		{"reassignment", []step{
			{"count = 5\ncount", "5"},
			{"count = 12\ncount", "12"},
		}},
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
				if d != nil {
					t.Fatalf("evaluating %q: unexpected diagnostic %+v", step.input, *d)
				}
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
		{"true", "ready = true\nready", true},
		{"false", "ready = false\nready", false},
		{"comparison-initializer", "ready = 1 < 2\nready", true},
		{"reassignment", "ready = true\nready = false\nready", false},
		{"reassignment-preserves-copy", "ready = true\noriginal = ready\nready = false\noriginal", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, d := testEval(t, tt.input)
			if d != nil {
				t.Fatalf("unexpected runtime diagnostic: %+v", *d)
			}
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
		{"undefined-initializer", "value = missing", source.Span{Start: 8, End: 15}},
		{"self-initialization", "value = value", source.Span{Start: 8, End: 13}},
		{"read-before-assignment", "value\nvalue = 1", source.Span{Start: 0, End: 5}},
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
		{"new-binding", "value = missing", diagnostic.UnboundVariable, source.Span{Start: 8, End: 15}},
		{"reassignment", "saved = missing", diagnostic.UnboundVariable, source.Span{Start: 8, End: 15}},
		{"type-mismatch", "value = 5 + true", diagnostic.UnknownOperator, source.Span{Start: 10, End: 11}},
		{"division-by-zero", "saved = 1 // 0", diagnostic.DivisionByZero, source.Span{Start: 10, End: 12}},
		{"stop-after-error", "value = missing\nsaved = 0", diagnostic.UnboundVariable, source.Span{Start: 8, End: 15}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := object.NewEnvironment()
			value, d := testEvalWithEnvironment(t, "saved = 7\nsaved", env)
			if d != nil {
				t.Fatalf("unexpected setup diagnostic: %+v", *d)
			}
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
