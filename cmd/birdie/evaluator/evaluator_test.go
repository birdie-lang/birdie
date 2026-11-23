package evaluator

import (
	"testing"

	"github.com/birdie-lang/birdie/cmd/birdie/diagnostic"
	"github.com/birdie-lang/birdie/cmd/birdie/lexer"
	"github.com/birdie-lang/birdie/cmd/birdie/object"
	"github.com/birdie-lang/birdie/cmd/birdie/parser"
	"github.com/birdie-lang/birdie/cmd/birdie/source"
	"github.com/birdie-lang/birdie/internal/testutil"
)

func TestEvalIntegerExpression(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"0", "0"},
		{"-5", "-5"},
		{"5 * 2 + 10", "20"},
		{"5 + 2 * 10", "25"},
		{"20 - 5 - 3", "12"},
		{"9223372036854775807 + 1", "9223372036854775808"},
		{"-9223372036854775808 + -1", "-9223372036854775809"},
		{"-9223372036854775808 - 1", "-9223372036854775809"},
		{"4611686018427387904 * -2", "-9223372036854775808"},
		{"4611686018427387904 * 4", "18446744073709551616"},
		{"-9223372036854775808 * -1", "9223372036854775808"},
		{"9223372036854775808 * 3", "27670116110564327424"},
		{"n = -9223372036854775808\n-n", "9223372036854775808"},
		{"-7 / 2", "-3"},
		{"5\n10\n15\n", "15"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			value, d := testEval(t, tt.input)
			requireNoRuntimeDiagnostic(t, tt.input, d)
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
		{"-9223372036854775808 // -1", "9223372036854775808"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			value, d := testEval(t, tt.input)
			requireNoRuntimeDiagnostic(t, tt.input, d)
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
		{"!true", false},
		{"!false", true},
		{"1 < 2", true},
		{"1 > 2", false},
		{"1 < 1", false},
		{"1 > 1", false},
		{"1 == 1", true},
		{"1 != 1", false},
		{"1 == 2", false},
		{"1 != 2", true},
		{"true == true", true},
		{"true == false", false},
		{"true != false", true},
		{"true != true", false},
		{"2 + 3 * 4 == 14", true},
		{"9223372036854775808 > 9223372036854775807", true},
		{"-9223372036854775808 > -9223372036854775809", true},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			value, d := testEval(t, tt.input)
			requireNoRuntimeDiagnostic(t, tt.input, d)
			testBooleanObject(t, value, tt.want)
		})
	}
}

func TestRuntimeDiagnostics(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		code    diagnostic.Code
		span    source.Span
		message string
	}{
		{"mixed-operand-types", "5 + true", diagnostic.UnknownOperator, source.Span{Start: 2, End: 3}, "type mismatch: INTEGER + BOOLEAN"},
		{"boolean-addition", "true + false", diagnostic.UnknownOperator, source.Span{Start: 5, End: 6}, "unknown operator: BOOLEAN + BOOLEAN"},
		{"nested-right-error", "1 + true * false", diagnostic.UnknownOperator, source.Span{Start: 9, End: 10}, ""},
		{"first-operand-error", "true + false + true * false", diagnostic.UnknownOperator, source.Span{Start: 5, End: 6}, ""},
		{"error-after-valid-statement", "1\n5 + true\n3", diagnostic.UnknownOperator, source.Span{Start: 4, End: 5}, ""},
		{"division-by-zero", "5 // 0", diagnostic.DivisionByZero, source.Span{Start: 2, End: 4}, ""},
		{"first-list-element-error", "[1, 2 // 0, missing]", diagnostic.DivisionByZero, source.Span{Start: 6, End: 8}, ""},
		{"index-receiver-error", "missing[1 // 0]", diagnostic.UnboundVariable, source.Span{Start: 0, End: 7}, ""},
		{"index-expression-error", "[10][1 // 0]", diagnostic.DivisionByZero, source.Span{Start: 7, End: 9}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := testRuntimeDiagnostic(t, tt.input, tt.code, tt.span)
			if tt.message != "" && d.Message != tt.message {
				t.Errorf("got message %q, want %q", d.Message, tt.message)
			}
		})
	}
}

func testRuntimeDiagnostic(t *testing.T, input string, code diagnostic.Code, span source.Span) *diagnostic.Diagnostic {
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
	return d
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
	testutil.RequireNoDiagnostics(t, "lexer", input, l.Diagnostics()...)
	testutil.RequireNoDiagnostics(t, "parser", input, p.Diagnostics()...)
	return Eval(program, env)
}

func requireNoRuntimeDiagnostic(t *testing.T, input string, d *diagnostic.Diagnostic) {
	t.Helper()
	if d != nil {
		testutil.RequireNoDiagnostics(t, "runtime", input, *d)
	}
}

func testIntegerObject(t *testing.T, value object.Object, want string) {
	t.Helper()
	result, ok := value.(*object.Integer)
	if !ok {
		t.Fatalf("expected an integer, got %T (%+v)", value, value)
	}
	if got := result.Inspect(); got != want {
		t.Errorf("got %s, want %s", got, want)
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
