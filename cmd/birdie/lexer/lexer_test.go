package lexer

import (
	"testing"

	"github.com/birdie-lang/birdie/cmd/birdie/token"
)

func tok(kind token.TokenType, literal ...string) token.Token {
	if len(literal) > 1 {
		panic("tok accepts at most one literal")
	}

	result := token.Token{Type: kind}
	if len(literal) == 1 {
		result.Literal = literal[0]
	}
	return result
}

func TestNextToken(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []token.Token
	}{
		{
			name: "mixed-statements",
			input: `1 + 2 * 3
x = 2
x = x + 3
ready = true
done = false`,
			want: []token.Token{
				tok(token.INT, "1"), tok(token.PLUS), tok(token.INT, "2"),
				tok(token.MUL), tok(token.INT, "3"), tok(token.EOL),
				tok(token.IDENT, "x"), tok(token.ASSIGN), tok(token.INT, "2"), tok(token.EOL),
				tok(token.IDENT, "x"), tok(token.ASSIGN), tok(token.IDENT, "x"),
				tok(token.PLUS), tok(token.INT, "3"), tok(token.EOL),
				tok(token.IDENT, "ready"), tok(token.ASSIGN), tok(token.TRUE), tok(token.EOL),
				tok(token.IDENT, "done"), tok(token.ASSIGN), tok(token.FALSE),
				tok(token.EOF),
			},
		},
		{
			name:  "operators",
			input: "(1 + 2) * 3 - 4 / 2 < 5 > 0 == true != false",
			want: []token.Token{
				tok(token.LPAREN), tok(token.INT, "1"), tok(token.PLUS), tok(token.INT, "2"), tok(token.RPAREN),
				tok(token.MUL), tok(token.INT, "3"), tok(token.MINUS), tok(token.INT, "4"),
				tok(token.DIV), tok(token.INT, "2"), tok(token.LT), tok(token.INT, "5"),
				tok(token.GT), tok(token.INT, "0"), tok(token.EQ), tok(token.TRUE),
				tok(token.NEQ), tok(token.FALSE), tok(token.EOF),
			},
		},
		{
			name:  "large-integer",
			input: "1234567890123456789012345678901234567890",
			want: []token.Token{
				tok(token.INT, "1234567890123456789012345678901234567890"), tok(token.EOF),
			},
		},
		{
			name:  "strings",
			input: "label = \"hello world\"\nempty = \"\"",
			want: []token.Token{
				tok(token.IDENT, "label"), tok(token.ASSIGN), tok(token.STRING, "hello world"), tok(token.EOL),
				tok(token.IDENT, "empty"), tok(token.ASSIGN), tok(token.STRING), tok(token.EOF),
			},
		},
		{
			name:  "empty-input",
			input: "",
			want:  []token.Token{tok(token.EOF)},
		},
		{
			name:  "identifiers-and-type-names",
			input: "c1 row2 total_count int bool list set trueValue",
			want: []token.Token{
				tok(token.IDENT, "c1"), tok(token.IDENT, "row2"), tok(token.IDENT, "total_count"),
				tok(token.IDENT, "int"), tok(token.IDENT, "bool"), tok(token.IDENT, "list"),
				tok(token.IDENT, "set"), tok(token.IDENT, "trueValue"), tok(token.EOF),
			},
		},
		{
			name: "function",
			input: `fn add(a: int, b: int) int:
    return a + b
`,
			want: []token.Token{
				tok(token.FUN), tok(token.IDENT, "add"), tok(token.LPAREN),
				tok(token.IDENT, "a"), tok(token.COLON), tok(token.IDENT, "int"), tok(token.COMMA),
				tok(token.IDENT, "b"), tok(token.COLON), tok(token.IDENT, "int"), tok(token.RPAREN),
				tok(token.IDENT, "int"), tok(token.COLON), tok(token.EOL), tok(token.INDENT),
				tok(token.RETURN), tok(token.IDENT, "a"), tok(token.PLUS), tok(token.IDENT, "b"),
				tok(token.EOL), tok(token.DEDENT), tok(token.EOF),
			},
		},
		{
			name: "for-loop",
			input: `for num in nums:
    total = total + num
`,
			want: []token.Token{
				tok(token.FOR), tok(token.IDENT, "num"), tok(token.IN), tok(token.IDENT, "nums"),
				tok(token.COLON), tok(token.EOL), tok(token.INDENT),
				tok(token.IDENT, "total"), tok(token.ASSIGN), tok(token.IDENT, "total"),
				tok(token.PLUS), tok(token.IDENT, "num"), tok(token.EOL),
				tok(token.DEDENT), tok(token.EOF),
			},
		},
		{
			name: "nested-indentation",
			input: `if ready:
    if found:
        return true
    return false
x = 1
`,
			want: []token.Token{
				tok(token.IF), tok(token.IDENT, "ready"), tok(token.COLON), tok(token.EOL),
				tok(token.INDENT), tok(token.IF), tok(token.IDENT, "found"), tok(token.COLON), tok(token.EOL),
				tok(token.INDENT), tok(token.RETURN), tok(token.TRUE), tok(token.EOL),
				tok(token.DEDENT), tok(token.RETURN), tok(token.FALSE), tok(token.EOL),
				tok(token.DEDENT), tok(token.IDENT, "x"), tok(token.ASSIGN), tok(token.INT, "1"),
				tok(token.EOL), tok(token.EOF),
			},
		},
		{
			name: "multiple-dedents-before-statement",
			input: `if ready:
    if found:
        return true
x = 1
`,
			want: []token.Token{
				tok(token.IF), tok(token.IDENT, "ready"), tok(token.COLON), tok(token.EOL),
				tok(token.INDENT), tok(token.IF), tok(token.IDENT, "found"), tok(token.COLON), tok(token.EOL),
				tok(token.INDENT), tok(token.RETURN), tok(token.TRUE), tok(token.EOL),
				tok(token.DEDENT), tok(token.DEDENT), tok(token.IDENT, "x"), tok(token.ASSIGN), tok(token.INT, "1"),
				tok(token.EOL), tok(token.EOF),
			},
		},
		{
			name:  "eof-without-final-newline",
			input: "if ready:\n    if found:\n        return true",
			want: []token.Token{
				tok(token.IF), tok(token.IDENT, "ready"), tok(token.COLON), tok(token.EOL),
				tok(token.INDENT), tok(token.IF), tok(token.IDENT, "found"), tok(token.COLON), tok(token.EOL),
				tok(token.INDENT), tok(token.RETURN), tok(token.TRUE),
				tok(token.DEDENT), tok(token.DEDENT), tok(token.EOF),
			},
		},
		{
			name: "dedent-before-else-and-eof",
			input: `if ready:
    return true
else:
    return false
`,
			want: []token.Token{
				tok(token.IF), tok(token.IDENT, "ready"), tok(token.COLON), tok(token.EOL),
				tok(token.INDENT), tok(token.RETURN), tok(token.TRUE), tok(token.EOL),
				tok(token.DEDENT), tok(token.ELSE), tok(token.COLON), tok(token.EOL),
				tok(token.INDENT), tok(token.RETURN), tok(token.FALSE), tok(token.EOL),
				tok(token.DEDENT), tok(token.EOF),
			},
		},
		{
			name: "calls-and-collections",
			input: `nums = [1, 2]
seen = set()
seen.add(num)
`,
			want: []token.Token{
				tok(token.IDENT, "nums"), tok(token.ASSIGN), tok(token.LBRACKET),
				tok(token.INT, "1"), tok(token.COMMA), tok(token.INT, "2"), tok(token.RBRACKET), tok(token.EOL),
				tok(token.IDENT, "seen"), tok(token.ASSIGN), tok(token.IDENT, "set"),
				tok(token.LPAREN), tok(token.RPAREN), tok(token.EOL),
				tok(token.IDENT, "seen"), tok(token.DOT), tok(token.IDENT, "add"),
				tok(token.LPAREN), tok(token.IDENT, "num"), tok(token.RPAREN),
				tok(token.EOL), tok(token.EOF),
			},
		},
		{
			name:  "blank-lines",
			input: "\n\nx = 1\n\n",
			want: []token.Token{
				tok(token.IDENT, "x"), tok(token.ASSIGN), tok(token.INT, "1"), tok(token.EOL), tok(token.EOF),
			},
		},
		{
			name:  "blank-lines-within-block",
			input: "if ready:\n    x = 1\n\n  \n    y = 2\nz = 3\n",
			want: []token.Token{
				tok(token.IF), tok(token.IDENT, "ready"), tok(token.COLON), tok(token.EOL),
				tok(token.INDENT), tok(token.IDENT, "x"), tok(token.ASSIGN), tok(token.INT, "1"), tok(token.EOL),
				tok(token.IDENT, "y"), tok(token.ASSIGN), tok(token.INT, "2"), tok(token.EOL),
				tok(token.DEDENT), tok(token.IDENT, "z"), tok(token.ASSIGN), tok(token.INT, "3"), tok(token.EOL),
				tok(token.EOF),
			},
		},
		{
			name:  "trailing-spaces-at-eof",
			input: "if ready:\n    x = 1\n  ",
			want: []token.Token{
				tok(token.IF), tok(token.IDENT, "ready"), tok(token.COLON), tok(token.EOL),
				tok(token.INDENT), tok(token.IDENT, "x"), tok(token.ASSIGN), tok(token.INT, "1"), tok(token.EOL),
				tok(token.DEDENT), tok(token.EOF),
			},
		},
		{
			name: "multiline-call",
			input: `add(
    1,
    2
)
`,
			want: []token.Token{
				tok(token.IDENT, "add"), tok(token.LPAREN), tok(token.INT, "1"), tok(token.COMMA),
				tok(token.INT, "2"), tok(token.RPAREN), tok(token.EOL), tok(token.EOF),
			},
		},
		{
			name: "multiline-list",
			input: `nums = [
    1,
    2
]
`,
			want: []token.Token{
				tok(token.IDENT, "nums"), tok(token.ASSIGN), tok(token.LBRACKET),
				tok(token.INT, "1"), tok(token.COMMA), tok(token.INT, "2"), tok(token.RBRACKET),
				tok(token.EOL), tok(token.EOF),
			},
		},
		{
			name: "multiline-braces",
			input: `values = {
    1: 2,
}
next = 3
`,
			want: []token.Token{
				tok(token.IDENT, "values"), tok(token.ASSIGN), tok(token.LBRACE),
				tok(token.INT, "1"), tok(token.COLON), tok(token.INT, "2"), tok(token.COMMA),
				tok(token.RBRACE), tok(token.EOL),
				tok(token.IDENT, "next"), tok(token.ASSIGN), tok(token.INT, "3"), tok(token.EOL), tok(token.EOF),
			},
		},
		{
			name: "nested-multiline-delimiters",
			input: `result = add(
    [
        1,
    ],
    2
)
next = 3
`,
			want: []token.Token{
				tok(token.IDENT, "result"), tok(token.ASSIGN), tok(token.IDENT, "add"), tok(token.LPAREN),
				tok(token.LBRACKET), tok(token.INT, "1"), tok(token.COMMA), tok(token.RBRACKET), tok(token.COMMA),
				tok(token.INT, "2"), tok(token.RPAREN), tok(token.EOL),
				tok(token.IDENT, "next"), tok(token.ASSIGN), tok(token.INT, "3"), tok(token.EOL), tok(token.EOF),
			},
		},
		{
			name: "multiline-expression-within-block",
			input: `if ready:
    result = add(
        1,
        2
    )
    done = true
next = 3
`,
			want: []token.Token{
				tok(token.IF), tok(token.IDENT, "ready"), tok(token.COLON), tok(token.EOL),
				tok(token.INDENT), tok(token.IDENT, "result"), tok(token.ASSIGN), tok(token.IDENT, "add"), tok(token.LPAREN),
				tok(token.INT, "1"), tok(token.COMMA), tok(token.INT, "2"), tok(token.RPAREN), tok(token.EOL),
				tok(token.IDENT, "done"), tok(token.ASSIGN), tok(token.TRUE), tok(token.EOL),
				tok(token.DEDENT), tok(token.IDENT, "next"), tok(token.ASSIGN), tok(token.INT, "3"), tok(token.EOL),
				tok(token.EOF),
			},
		},
		{
			name:  "delimiters-in-string",
			input: "text = \"([{\"\nx = 1\n",
			want: []token.Token{
				tok(token.IDENT, "text"), tok(token.ASSIGN), tok(token.STRING, "([{"), tok(token.EOL),
				tok(token.IDENT, "x"), tok(token.ASSIGN), tok(token.INT, "1"), tok(token.EOL), tok(token.EOF),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New(tt.input)
			for i, expected := range tt.want {
				got := l.NextToken()
				if got.Type != expected.Type || got.Literal != expected.Literal {
					t.Fatalf("token %d: got %s(%q), want %s(%q)", i, got.Type, got.Literal, expected.Type, expected.Literal)
				}
			}
			if got := l.NextToken(); got.Type != token.EOF {
				t.Fatalf("after EOF: got %s(%q), want EOF", got.Type, got.Literal)
			}
			if errors := l.Errors(); len(errors) != 0 {
				t.Fatalf("unexpected lexer errors: %v", errors)
			}
		})
	}
}

func TestIndentationErrors(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"partial-level", "if ready:\n  x = 1\nafter = 2\n"},
		{"skipped-level", "if ready:\n        x = 1\nafter = 2\n"},
		{"tab-indentation", "if ready:\n\tx = 1\nafter = 2\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New(tt.input)
			sawAfter := false
			for i := 0; i <= len(tt.input); i++ {
				got := l.NextToken()
				if got.Type == token.IDENT && got.Literal == "after" {
					sawAfter = true
				}
				if got.Type == token.EOF {
					if len(l.Errors()) == 0 {
						t.Fatal("expected an indentation error")
					}
					if !sawAfter {
						t.Fatal("lexer did not continue to the statement after the invalid indentation")
					}
					return
				}
			}
			t.Fatal("lexer did not reach EOF")
		})
	}
}
