package parser

import (
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

			program := parseAndCheckProgram(t, input)
			if len(program.Statements) != 1 {
				t.Fatalf("expected one statement, got=%d", len(program.Statements))
			}
			stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
			if !ok {
				t.Fatalf("program.Statements[0] is not ast.ExpressionStatement. got=%T",
					program.Statements[0])
			}

			testIntegerLiteral(t, stmt.Expression, expected)
		})
	}
}

func testIntegerLiteral(t *testing.T, il ast.Expression, value *big.Int) bool {
	t.Helper()
	integ, ok := il.(*ast.IntegerLiteral)
	if !ok {
		t.Errorf("il not *ast.IntegerLiteral. got=%T", il)
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
