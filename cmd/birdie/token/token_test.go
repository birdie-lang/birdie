package token

import "testing"

func TestTokenText(t *testing.T) {
	tests := []struct {
		name string
		tok  Token
		want string
	}{
		{"identifier", Token{Type: IDENT, Literal: "count"}, "count"},
		{"integer", Token{Type: INT, Literal: "123456789012345678901234567890"}, "123456789012345678901234567890"},
		{"string", Token{Type: STRING, Literal: "hello world"}, "hello world"},
		{"empty-string", Token{Type: STRING}, ""},
		{"illegal-character", Token{Type: ILLEGAL, Literal: "@"}, "@"},
		{"operator", Token{Type: PLUS}, "+"},
		{"dot", Token{Type: DOT}, "."},
		{"comparison", Token{Type: EQ}, "=="},
		{"true", Token{Type: TRUE}, "true"},
		{"false", Token{Type: FALSE}, "false"},
		{"return", Token{Type: RETURN}, "return"},
		{"line-ending", Token{Type: EOL}, "\n"},
		{"indent", Token{Type: INDENT}, ""},
		{"dedent", Token{Type: DEDENT}, ""},
		{"eof", Token{Type: EOF}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.tok.Text(); got != tt.want {
				t.Fatalf("Text() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestTokenSpelling(t *testing.T) {
	if PLUS.Spelling() != "+" || LBRACKET.Spelling() != "[" || RBRACKET.Spelling() != "]" {
		t.Fatal("operator and bracket spellings must match their source text")
	}
	for spelling, kind := range keywords {
		t.Run(spelling, func(t *testing.T) {
			if got := kind.Spelling(); got != spelling {
				t.Fatalf("Spelling() = %q, want %q", got, spelling)
			}
		})
	}
}
