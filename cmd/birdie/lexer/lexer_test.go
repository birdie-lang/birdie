package lexer

import (
	"testing"

	"github.com/birdie-lang/birdie/cmd/birdie/diagnostic"
	"github.com/birdie-lang/birdie/cmd/birdie/source"
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
			if diagnostics := l.Diagnostics(); len(diagnostics) != 0 {
				t.Fatalf("unexpected lexer diagnostics: %+v", diagnostics)
			}
		})
	}
}

func TestIndentationErrors(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantCode diagnostic.Code
		wantSpan source.Span
	}{
		{"partial-level", "if ready:\n  x = 1\nafter = 2\n", diagnostic.InvalidIndentation, source.Span{Start: 10, End: 12}},
		{"skipped-level", "if ready:\n        x = 1\nafter = 2\n", diagnostic.InvalidIndentation, source.Span{Start: 10, End: 18}},
		{"tab-indentation", "if ready:\n\tx = 1\nafter = 2\n", diagnostic.TabsInIndentation, source.Span{Start: 10, End: 11}},
		{"mixed-indentation", "if ready:\n \t  x = 1\nafter = 2\n", diagnostic.TabsInIndentation, source.Span{Start: 11, End: 12}},
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
					diagnostics := l.Diagnostics()
					if len(diagnostics) != 1 {
						t.Fatalf("expected one indentation diagnostic, got %+v", diagnostics)
					}
					if d := diagnostics[0]; d.Code != tt.wantCode || d.Span != tt.wantSpan || d.Message == "" {
						t.Fatalf("got %+v, want code %s and span %+v with a message", d, tt.wantCode, tt.wantSpan)
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

func TestTokenSpans(t *testing.T) {
	l := New("if x:\n    y = \"é\"\nz = 12")
	want := []token.Token{
		{Type: token.IF, Span: source.Span{Start: 0, End: 2}},
		{Type: token.IDENT, Literal: "x", Span: source.Span{Start: 3, End: 4}},
		{Type: token.COLON, Span: source.Span{Start: 4, End: 5}},
		{Type: token.EOL, Span: source.Span{Start: 5, End: 6}},
		{Type: token.INDENT, Span: source.Span{Start: 10, End: 10}},
		{Type: token.IDENT, Literal: "y", Span: source.Span{Start: 10, End: 11}},
		{Type: token.ASSIGN, Span: source.Span{Start: 12, End: 13}},
		{Type: token.STRING, Literal: "é", Span: source.Span{Start: 14, End: 18}},
		{Type: token.EOL, Span: source.Span{Start: 18, End: 19}},
		{Type: token.DEDENT, Span: source.Span{Start: 19, End: 19}},
		{Type: token.IDENT, Literal: "z", Span: source.Span{Start: 19, End: 20}},
		{Type: token.ASSIGN, Span: source.Span{Start: 21, End: 22}},
		{Type: token.INT, Literal: "12", Span: source.Span{Start: 23, End: 25}},
		{Type: token.EOF, Span: source.Span{Start: 25, End: 25}},
		{Type: token.EOF, Span: source.Span{Start: 25, End: 25}},
	}
	for i, expected := range want {
		if got := l.NextToken(); got != expected {
			t.Fatalf("token %d: got %+v, want %+v", i, got, expected)
		}
	}
}

func TestUnexpectedCharacters(t *testing.T) {
	l := New("@ é\nx = 1")
	wantTokens := []token.Token{
		tok(token.ILLEGAL, "@"), tok(token.ILLEGAL, "é"), tok(token.EOL),
		tok(token.IDENT, "x"), tok(token.ASSIGN), tok(token.INT, "1"), tok(token.EOF),
	}
	for i, expected := range wantTokens {
		if got := l.NextToken(); got.Type != expected.Type || got.Literal != expected.Literal {
			t.Fatalf("token %d: got %+v, want %+v", i, got, expected)
		}
	}
	wantSpans := []source.Span{{Start: 0, End: 1}, {Start: 2, End: 4}}
	diagnostics := l.Diagnostics()
	if len(diagnostics) != len(wantSpans) {
		t.Fatalf("expected two diagnostics, got %+v", diagnostics)
	}
	for i, d := range diagnostics {
		if d.Code != diagnostic.UnexpectedCharacter || d.Span != wantSpans[i] || d.Message == "" {
			t.Errorf("diagnostic %d: got %+v, want unexpected character at %+v", i, d, wantSpans[i])
		}
	}
}
