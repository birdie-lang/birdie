package evaluator

import (
	"math/big"
	"testing"

	"github.com/birdie-lang/birdie/cmd/birdie/diagnostic"
	"github.com/birdie-lang/birdie/cmd/birdie/lexer"
	"github.com/birdie-lang/birdie/cmd/birdie/object"
	"github.com/birdie-lang/birdie/cmd/birdie/parser"
	"github.com/birdie-lang/birdie/cmd/birdie/source"
)

func TestEvalIntegerExpression(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"0", "0"},
		{"5", "5"},
		{"10", "10"},
		{"0 - 5", "-5"},
		{"5 + 5 + 5 + 5 - 10", "10"},
		{"2 * 2 * 2 * 2 * 2", "32"},
		{"5 * 2 + 10", "20"},
		{"5 + 2 * 10", "25"},
		{"3 * 3 * 3 + 10", "37"},
		{"20 - 5 - 3", "12"},
		{"9223372036854775808", "9223372036854775808"},
		{"1234567890123456789012345678901234567890", "1234567890123456789012345678901234567890"},
		{"9223372036854775807 + 1", "9223372036854775808"},
		{"9007199254740992 + 1", "9007199254740993"},
		{"9223372036854775808 * 3", "27670116110564327424"},
		{"9223372036854775809 - 9223372036854775808", "1"},
		{"5\n10\n15\n", "15"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			value, d := testEval(t, tt.input)
			if d != nil {
				t.Fatalf("unexpected runtime diagnostic: %+v", *d)
			}
			testIntegerObject(t, value, tt.want)
		})
	}
}

func TestEvalFloorDivision(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"7 // 2", "3"},
		{"-7 // 2", "-4"},
		{"7 // -2", "-4"},
		{"-7 // -2", "3"},
		{"1 // 2", "0"},
		{"-1 // 2", "-1"},
		{"50 // 2 * 2 + 10", "60"},
		{"50 // 5 // 2", "5"},
		{"27670116110564327424 // 3", "9223372036854775808"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			value, d := testEval(t, tt.input)
			if d != nil {
				t.Fatalf("unexpected runtime diagnostic: %+v", *d)
			}
			testIntegerObject(t, value, tt.want)
		})
	}
}

func TestEvalBooleanExpression(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"true", true},
		{"false", false},
		{"1 < 2", true},
		{"1 > 2", false},
		{"1 < 1", false},
		{"1 > 1", false},
		{"1 == 1", true},
		{"1 != 1", false},
		{"1 == 2", false},
		{"1 != 2", true},
		{"true == true", true},
		{"false == false", true},
		{"true == false", false},
		{"true != false", true},
		{"false != true", true},
		{"2 + 3 * 4 == 14", true},
		{"10 * 2 != 20", false},
		{"9007199254740992 == 9007199254740993", false},
		{"9223372036854775808 < 9223372036854775809", true},
		{"9223372036854775808 > 9223372036854775807", true},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			value, d := testEval(t, tt.input)
			if d != nil {
				t.Fatalf("unexpected runtime diagnostic: %+v", *d)
			}
			testBooleanObject(t, value, tt.want)
		})
	}
}

func TestRuntimeDiagnostics(t *testing.T) {
	tests := []struct {
		name  string
		input string
		code  diagnostic.Code
		span  source.Span
	}{
		{"mixed-operand-types", "5 + true", diagnostic.UnknownOperator, source.Span{Start: 2, End: 3}},
		{"boolean-addition", "true + false", diagnostic.UnknownOperator, source.Span{Start: 5, End: 6}},
		{"boolean-comparison", "true < false", diagnostic.UnknownOperator, source.Span{Start: 5, End: 6}},
		{"nested-right-error", "1 + true * false", diagnostic.UnknownOperator, source.Span{Start: 9, End: 10}},
		{"nested-left-error", "true + false + 1", diagnostic.UnknownOperator, source.Span{Start: 5, End: 6}},
		{"first-operand-error", "true + false + true * false", diagnostic.UnknownOperator, source.Span{Start: 5, End: 6}},
		{"stop-before-next-statement", "true + false\n5 + true", diagnostic.UnknownOperator, source.Span{Start: 5, End: 6}},
		{"error-after-valid-statement", "1\n5 + true\n3", diagnostic.UnknownOperator, source.Span{Start: 4, End: 5}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testRuntimeDiagnostic(t, tt.input, tt.code, tt.span)
		})
	}
}

func TestDivisionByZeroDiagnostics(t *testing.T) {
	tests := []struct {
		name  string
		input string
		span  source.Span
	}{
		{"division-by-zero", "5 // 0", source.Span{Start: 2, End: 4}},
		{"zero-divided-by-zero", "0 // 0", source.Span{Start: 2, End: 4}},
		{"nested-right-error", "1 + 10 // 0", source.Span{Start: 7, End: 9}},
		{"nested-left-error", "1 // 0 + 10", source.Span{Start: 2, End: 4}},
		{"first-operand-error", "1 // 0 + 2 // 0", source.Span{Start: 2, End: 4}},
		{"stop-before-next-statement", "1 // 0\ntrue + false", source.Span{Start: 2, End: 4}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testRuntimeDiagnostic(t, tt.input, diagnostic.DivisionByZero, tt.span)
		})
	}
}

func TestRuntimeDiagnosticMessages(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"5 + true", "type mismatch: INTEGER + BOOLEAN"},
		{"true + false", "unknown operator: BOOLEAN + BOOLEAN"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			_, d := testEval(t, tt.input)
			if d == nil {
				t.Fatal("expected a runtime diagnostic")
			}
			if d.Message != tt.want {
				t.Errorf("got message %q, want %q", d.Message, tt.want)
			}
		})
	}
}

func testRuntimeDiagnostic(t *testing.T, input string, code diagnostic.Code, span source.Span) {
	t.Helper()
	value, d := testEval(t, input)
	if d == nil {
		t.Fatalf("expected runtime diagnostic %s, got value %+v", code, value)
	}
	if value != nil {
		t.Errorf("expected no value on failure, got %+v", value)
	}
	if d.Code != code || d.Span != span {
		t.Errorf("got diagnostic %+v, want code %s and span %+v", *d, code, span)
	}
	if d.Message == "" {
		t.Error("runtime diagnostic has no message")
	}
}

func testEval(t *testing.T, input string) (object.Object, *diagnostic.Diagnostic) {
	t.Helper()
	return testEvalWithEnvironment(t, input, object.NewEnvironment())
}

// Front-end diagnostics are setup failures here; runtime diagnostics are returned
// so individual cases can assert success or an expected evaluation failure.
func testEvalWithEnvironment(t *testing.T, input string, env *object.Environment) (object.Object, *diagnostic.Diagnostic) {
	t.Helper()
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if diagnostics := l.Diagnostics(); len(diagnostics) != 0 {
		t.Fatalf("unexpected lexer diagnostics: %+v", diagnostics)
	}
	if diagnostics := p.Diagnostics(); len(diagnostics) != 0 {
		t.Fatalf("unexpected parser diagnostics: %+v", diagnostics)
	}
	return Eval(program, env)
}

func testIntegerObject(t *testing.T, value object.Object, want string) {
	t.Helper()
	result, ok := value.(*object.Integer)
	if !ok {
		t.Fatalf("expected an integer, got %T (%+v)", value, value)
	}
	expected, ok := new(big.Int).SetString(want, 10)
	if !ok {
		t.Fatalf("invalid expected integer %q", want)
	}
	if result.Value.Cmp(expected) != 0 {
		t.Errorf("got %s, want %s", result.Value.String(), want)
	}
}

func testBooleanObject(t *testing.T, value object.Object, want bool) {
	t.Helper()
	result, ok := value.(*object.Boolean)
	if !ok {
		t.Fatalf("expected a boolean, got %T (%+v)", value, value)
	}
	if result.Value != want {
		t.Errorf("got %t, want %t", result.Value, want)
	}
}
