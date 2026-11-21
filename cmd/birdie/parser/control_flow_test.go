package parser

import (
	"testing"

	"github.com/birdie-lang/birdie/cmd/birdie/lexer"
	"github.com/birdie-lang/birdie/cmd/birdie/source"
	"github.com/birdie-lang/birdie/internal/testutil"
)

func TestParseControlFlowBlock(t *testing.T) {
	const function = `fn choose(ready: bool) int:
    if ready:
        return 1
    else if false:
        return 0
    else:
        return 2`
	program := parseAndCheckProgram(t, function+"\nafter = 99")
	if len(program.Statements) != 2 {
		t.Fatalf("expected a function and a following assignment, got %d statements", len(program.Statements))
	}
	if got := program.Statements[0].String(); got != function {
		t.Errorf("function AST =\n%s\nwant:\n%s", got, function)
	}
	if got := program.Statements[1].String(); got != "after = 99" {
		t.Errorf("statement after the function = %q, want after = 99", got)
	}
}

func TestControlFlowSyntaxDiagnostics(t *testing.T) {
	tests := []struct {
		name, input string
		span        source.Span
	}{
		{"missing-colon", "if true\n    value = 1", source.Span{Start: 7, End: 8}},
		{"missing-body", "fn answer() int:\nreturn 42", source.Span{Start: 17, End: 23}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			p.ParseProgram()
			testutil.RequireNoDiagnostics(t, "lexer", tt.input, l.Diagnostics()...)
			diagnostics := p.Diagnostics()
			if len(diagnostics) == 0 {
				t.Fatal("expected a syntax diagnostic")
			}
			d := diagnostics[0]
			if d.Code == "" || d.Message == "" || d.Span != tt.span {
				t.Errorf("got first diagnostic %+v, want span %+v and a code/message", d, tt.span)
			}
		})
	}
}
