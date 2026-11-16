package lexer

import (
	"testing"

	"github.com/birdie-lang/birdie/cmd/birdie/token"
)

func TestNextToken(t *testing.T) {
	input :=
		`1 + 2 + 3
x = 2
x = x + 3`

	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{token.INT, "1"},
		{token.PLUS, "+"},
		{token.INT, "2"},
		{token.PLUS, "+"},
		{token.INT, "3"},
		{token.EOL, "\n"},
		{token.IDENT, "x"},
		{token.ASSIGN, "="},
		{token.INT, "2"},
		{token.EOL, "\n"},
		{token.IDENT, "x"},
		{token.ASSIGN, "="},
		{token.IDENT, "x"},
		{token.PLUS, "+"},
		{token.INT, "3"},
	}

	l := New(input)
	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - TokenType wrong. expected %q, got %q", i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - Literal wrong. expected %q, got %q", i, tt.expectedLiteral, tok.Literal)
		}
	}
}
