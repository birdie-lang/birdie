package parser

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/birdie-lang/birdie/cmd/birdie/ast"
	"github.com/birdie-lang/birdie/cmd/birdie/lexer"
)

func TestIntegerLiteralExpression(t *testing.T) {
	for _, input := range []string{
		"0",
		"5",
		"9223372036854775808",
		"1234567890123456789012345678901234567890",
	} {
		t.Run(input, func(t *testing.T) {
			expected, ok := new(big.Int).SetString(input, 10)
			if !ok {
				t.Fatalf("invalid integer test input %q", input)
			}

			testLiteralExpression(t, parseAndCheckExpression(t, input), expected)
		})
	}
}

func TestBooleanExpression(t *testing.T) {
	tests := []struct {
		input           string
		expectedBoolean bool
	}{
		{"true", true},
		{"false", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			testLiteralExpression(t, parseAndCheckExpression(t, tt.input), tt.expectedBoolean)
		})
	}
}

func TestAssignStatements(t *testing.T) {
	largeValue, ok := new(big.Int).SetString("1234567890123456789012345678901234567890", 10)
	if !ok {
		t.Fatal("invalid large integer test value")
	}

	tests := []struct {
		input              string
		expectedIdentifier string
		expectedValue      any
	}{
		{"x = 5", "x", 5},
		{"large = 9223372036854775807", "large", int64(9223372036854775807)},
		{"huge = 1234567890123456789012345678901234567890", "huge", largeValue},
		{"y = true", "y", true},
		{"ready = false", "ready", false},
		{"foobar = y", "foobar", "y"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			program := parseAndCheckProgram(t, tt.input)
			if len(program.Statements) != 1 {
				t.Fatalf("expected one statement, got=%d", len(program.Statements))
			}

			stmt := program.Statements[0]
			if !testAssignStatement(t, stmt, tt.expectedIdentifier) {
				return
			}
			testLiteralExpression(t, stmt.(*ast.AssignStatement).Value, tt.expectedValue)
			if got := stmt.String(); got != tt.input {
				t.Errorf("statement.String() = %q, want %q", got, tt.input)
			}
		})
	}
}

func testAssignStatement(t *testing.T, s ast.Statement, name string) bool {
	t.Helper()

	assignStmt, ok := s.(*ast.AssignStatement)
	if !ok {
		t.Errorf("s not *ast.AssignStatement. got=%T", s)
		return false
	}

	if assignStmt.Name.Value != name {
		t.Errorf("assignStmt.Name.Value not %q. got=%q", name, assignStmt.Name.Value)
		return false
	}

	if assignStmt.Name.TokenLiteral() != name {
		t.Errorf("assignStmt.Name.TokenLiteral() not %q. got=%q",
			name, assignStmt.Name.TokenLiteral())
		return false
	}

	return true
}

func testLiteralExpression(
	t *testing.T,
	exp ast.Expression,
	expected any,
) bool {
	t.Helper()
	switch v := expected.(type) {
	case int:
		return testIntegerLiteral(t, exp, big.NewInt(int64(v)))
	case int64:
		return testIntegerLiteral(t, exp, big.NewInt(v))
	case *big.Int:
		return testIntegerLiteral(t, exp, v)
	case string:
		return testIdentifier(t, exp, v)
	case bool:
		return testBooleanLiteral(t, exp, v)
	}
	t.Errorf("unsupported expected literal type: %T", expected)
	return false
}

func testIdentifier(t *testing.T, exp ast.Expression, value string) bool {
	t.Helper()
	ident, ok := exp.(*ast.Identifier)
	if !ok {
		t.Errorf("exp not *ast.Identifier. got=%T", exp)
		return false
	}

	if ident.Value != value {
		t.Errorf("ident.Value not %s. got=%s", value, ident.Value)
		return false
	}

	if ident.TokenLiteral() != value {
		t.Errorf("ident.TokenLiteral not %s. got=%s", value,
			ident.TokenLiteral())
		return false
	}

	return true
}

func testIntegerLiteral(t *testing.T, exp ast.Expression, value *big.Int) bool {
	t.Helper()
	integ, ok := exp.(*ast.IntegerLiteral)
	if !ok {
		t.Errorf("exp not *ast.IntegerLiteral. got=%T", exp)
		return false
	}

	if integ.Value.Cmp(value) != 0 {
		t.Errorf("integ.Value not %s. got=%s", value.String(), integ.Value.String())
		return false
	}

	if integ.TokenLiteral() != value.String() {
		t.Errorf("integ.TokenLiteral not %s. got=%s", value.String(),
			integ.TokenLiteral())
		return false
	}

	return true
}

func testBooleanLiteral(t *testing.T, exp ast.Expression, value bool) bool {
	t.Helper()
	boolean, ok := exp.(*ast.Boolean)
	if !ok {
		t.Errorf("exp not *ast.Boolean. got=%T", exp)
		return false
	}

	if boolean.Value != value {
		t.Errorf("boolean.Value not %t. got=%t", value, boolean.Value)
		return false
	}

	if boolean.TokenLiteral() != fmt.Sprintf("%t", value) {
		t.Errorf("boolean.TokenLiteral not %t. got=%s",
			value, boolean.TokenLiteral())
		return false
	}

	return true
}

func parseAndCheckExpression(t *testing.T, input string) ast.Expression {
	t.Helper()
	program := parseAndCheckProgram(t, input)
	if len(program.Statements) != 1 {
		t.Fatalf("expected one statement, got=%d", len(program.Statements))
	}
	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not *ast.ExpressionStatement. got=%T", program.Statements[0])
	}
	return stmt.Expression
}

func parseAndCheckProgram(t *testing.T, input string) *ast.Program {
	t.Helper()
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)
	return program
}

func checkParserErrors(t *testing.T, p *Parser) {
	t.Helper()
	errors := p.Errors()
	if len(errors) == 0 {
		return
	}

	t.Errorf("parser has %d errors", len(errors))
	for _, msg := range errors {
		t.Errorf("parser error: %q", msg)
	}
	t.FailNow()
}
