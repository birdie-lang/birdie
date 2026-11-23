package evaluator

import (
	"testing"

	"github.com/birdie-lang/birdie/cmd/birdie/object"
)

func TestEvalListLiterals(t *testing.T) {
	tests := []struct {
		input string
		want  []string
	}{
		{"[]", nil},
		{"[1 + 2, 4 * 5]", []string{"3", "20"}},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			value, d := testEval(t, tt.input)
			requireNoRuntimeDiagnostic(t, tt.input, d)
			list, ok := value.(*object.Array)
			if !ok {
				t.Fatalf("expected a list, got %T (%+v)", value, value)
			}
			if len(list.Elements) != len(tt.want) {
				t.Fatalf("got %d elements, want %d", len(list.Elements), len(tt.want))
			}
			for i, want := range tt.want {
				testIntegerObject(t, list.Elements[i], want)
			}
		})
	}
}

func TestEvalListIndex(t *testing.T) {
	const input = "rows = [[10, 20], [30]]\nrows[0][9223372036854775808 - 9223372036854775807]"
	value, d := testEval(t, input)
	requireNoRuntimeDiagnostic(t, input, d)
	testIntegerObject(t, value, "20")
}

func TestEvalInvalidListIndex(t *testing.T) {
	tests := []struct{ name, input string }{
		{"at-length", "[10][1]"},
		{"negative", "[10][-1]"},
		{"larger-than-machine-integer", "[10][18446744073709551616]"},
		{"non-integer", "[10][false]"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, d := testEval(t, tt.input)
			if value != nil || d == nil || d.Code == "" || d.Message == "" {
				t.Fatalf("got value %+v and diagnostic %+v, want a runtime diagnostic", value, d)
			}
		})
	}
}
