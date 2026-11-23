package parser

import (
	"testing"

	"github.com/birdie-lang/birdie/cmd/birdie/diagnostic"
	"github.com/birdie-lang/birdie/cmd/birdie/lexer"
	"github.com/birdie-lang/birdie/cmd/birdie/source"
	"github.com/birdie-lang/birdie/cmd/birdie/token"
)

func TestDiagnostics(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantLexer   []diagnostic.Diagnostic
		wantParser  []diagnostic.Diagnostic
		wantProgram string
	}{
		{
			name:       "missing-value-at-eof",
			input:      "x =",
			wantParser: []diagnostic.Diagnostic{{Code: diagnostic.ExpectedExpression, Span: source.Span{Start: 3, End: 3}}},
		},
		{
			name:       "missing-list-bracket",
			input:      "[1, 2",
			wantParser: []diagnostic.Diagnostic{{Code: diagnostic.UnexpectedToken, Span: source.Span{Start: 5, End: 5}}},
		},
		{
			name:       "missing-index-expression",
			input:      "nums[]",
			wantParser: []diagnostic.Diagnostic{{Code: diagnostic.ExpectedExpression, Span: source.Span{Start: 5, End: 6}}},
		},
		{
			name:        "missing-operand-before-newline",
			input:       "x = 1 +\ny = 2",
			wantParser:  []diagnostic.Diagnostic{{Code: diagnostic.ExpectedExpression, Span: source.Span{Start: 7, End: 8}}},
			wantProgram: "y = 2",
		},
		{
			name:        "unexpected-operator",
			input:       "x = * 2\ny = 3",
			wantParser:  []diagnostic.Diagnostic{{Code: diagnostic.ExpectedExpression, Span: source.Span{Start: 4, End: 5}}},
			wantProgram: "y = 3",
		},
		{
			name:        "invalid-assignment-target",
			input:       "1 + 2 = 3\nx = 4",
			wantParser:  []diagnostic.Diagnostic{{Code: diagnostic.InvalidAssignmentTarget, Span: source.Span{Start: 0, End: 5}}},
			wantProgram: "x = 4",
		},
		{
			name:        "invalid-integer",
			input:       "08\nx = 1",
			wantParser:  []diagnostic.Diagnostic{{Code: diagnostic.InvalidInteger, Span: source.Span{Start: 0, End: 2}}},
			wantProgram: "x = 1",
		},
		{
			name:        "lexer-error-without-parser-duplicate",
			input:       "x = @\ny = 2",
			wantLexer:   []diagnostic.Diagnostic{{Code: diagnostic.UnexpectedCharacter, Span: source.Span{Start: 4, End: 5}}},
			wantProgram: "y = 2",
		},
		{
			name:        "independent-lexer-and-parser-errors",
			input:       "@\nx =\ny = 3",
			wantLexer:   []diagnostic.Diagnostic{{Code: diagnostic.UnexpectedCharacter, Span: source.Span{Start: 0, End: 1}}},
			wantParser:  []diagnostic.Diagnostic{{Code: diagnostic.ExpectedExpression, Span: source.Span{Start: 5, End: 6}}},
			wantProgram: "y = 3",
		},
		{
			name:  "multiple-parser-errors",
			input: "x =\ny =\nz = 3",
			wantParser: []diagnostic.Diagnostic{
				{Code: diagnostic.ExpectedExpression, Span: source.Span{Start: 3, End: 4}},
				{Code: diagnostic.ExpectedExpression, Span: source.Span{Start: 7, End: 8}},
			},
			wantProgram: "z = 3",
		},
		{
			name:        "valid-multiple-statements",
			input:       "x = 1\ny = 2\n",
			wantProgram: "x = 1y = 2",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program := p.ParseProgram()
			checkDiagnostics(t, l.Diagnostics(), tt.wantLexer)
			checkDiagnostics(t, p.Diagnostics(), tt.wantParser)
			if got := program.String(); got != tt.wantProgram {
				t.Errorf("program.String() = %q, want %q", got, tt.wantProgram)
			}
		})
	}
}

func TestExpectedTokenDiagnostic(t *testing.T) {
	p := New(lexer.New("x + 1"))
	if p.expectPeek(token.COLON) {
		t.Fatal("expected a token mismatch")
	}
	checkDiagnostics(t, p.Diagnostics(), []diagnostic.Diagnostic{
		{Code: diagnostic.UnexpectedToken, Span: source.Span{Start: 2, End: 3}},
	})
	if got := p.Diagnostics()[0].Message; got != "expected \":\", found \"+\"" {
		t.Errorf("message = %q", got)
	}
}

func checkDiagnostics(t *testing.T, got, want []diagnostic.Diagnostic) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("got diagnostics %+v, want %+v", got, want)
	}
	for i, d := range got {
		if d.Code != want[i].Code || d.Span != want[i].Span || d.Message == "" {
			t.Errorf("diagnostic %d: got %+v, want code %s and span %+v with a message", i, d, want[i].Code, want[i].Span)
		}
	}
}
