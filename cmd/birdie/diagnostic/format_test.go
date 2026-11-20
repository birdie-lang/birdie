package diagnostic

import (
	"testing"

	"github.com/birdie-lang/birdie/cmd/birdie/source"
)

func TestFormat(t *testing.T) {
	tests := []struct {
		name string
		text string
		span source.Span
		want string
	}{
		{
			name: "line-ending",
			text: "x =\n",
			span: source.Span{Start: 3, End: 4},
			want: "example.bd:1:4: error[expected-expression]: expected an expression\n1 | x =\n  |    ^",
		},
		{
			name: "empty-file",
			text: "",
			span: source.Span{Start: 0, End: 0},
			want: "example.bd:1:1: error[expected-expression]: expected an expression\n1 | \n  | ^",
		},
		{
			name: "unicode-column",
			text: "é @",
			span: source.Span{Start: 3, End: 4},
			want: "example.bd:1:3: error[expected-expression]: expected an expression\n1 | é @\n  |   ^",
		},
		{
			name: "tab-width",
			text: " \tvalue",
			span: source.Span{Start: 1, End: 2},
			want: "example.bd:1:2: error[expected-expression]: expected an expression\n1 |     value\n  |  ^^^",
		},
		{
			name: "multiline-span",
			text: "one\ntwo",
			span: source.Span{Start: 1, End: 6},
			want: "example.bd:1:2: error[expected-expression]: expected an expression\n1 | one\n  |  ^^",
		},
		{
			name: "eof-after-newline",
			text: "x =\n",
			span: source.Span{Start: 4, End: 4},
			want: "example.bd:2:1: error[expected-expression]: expected an expression\n2 | \n  | ^",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := Diagnostic{Code: ExpectedExpression, Message: "expected an expression", Span: tt.span}
			if got := Format(source.NewFile("example.bd", tt.text), d); got != tt.want {
				t.Fatalf("Format() =\n%s\nwant:\n%s", got, tt.want)
			}
		})
	}
}
