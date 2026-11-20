package token

type TokenType string

type Token struct {
	Type    TokenType
	Literal string
}

// Text returns a payload or the token kind's fixed spelling. Empty string
// payloads are valid, so the choice depends on the kind, not the literal's length.
func (t Token) Text() string {
	switch t.Type {
	case IDENT, INT, STRING, ILLEGAL:
		return t.Literal
	default:
		return t.Type.Spelling()
	}
}

var keywords = map[string]TokenType{
	"fn":     FUN,
	"true":   TRUE,
	"false":  FALSE,
	"if":     IF,
	"else":   ELSE,
	"return": RETURN,
	"for":    FOR,
	"in":     IN,
}

var keywordSpellings = func() map[TokenType]string {
	spellings := make(map[TokenType]string, len(keywords))
	for spelling, kind := range keywords {
		spellings[kind] = spelling
	}
	return spellings
}()

// Spelling is defined once per fixed token kind; payload and structural kinds
// without source text return an empty string.
func (kind TokenType) Spelling() string {
	switch kind {
	case ASSIGN, PLUS, MINUS, NOT, MUL, DIV, GT, LT, EQ, NEQ,
		COMMA, COLON, DOT, LPAREN, RPAREN, LBRACE, RBRACE, LBRACKET, RBRACKET:
		return string(kind)
	case EOL:
		return "\n"
	default:
		return keywordSpellings[kind]
	}
}

func LookupIdent(ident string) TokenType {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return IDENT
}

const (
	ILLEGAL TokenType = "ILLEGAL"
	EOL     TokenType = "EOL"
	EOF     TokenType = "EOF"

	// Identifiers + Literal
	IDENT  TokenType = "IDENT"
	INT    TokenType = "INT"
	STRING TokenType = "STRING"

	// Operators
	ASSIGN TokenType = "="
	PLUS   TokenType = "+"
	MINUS  TokenType = "-"
	NOT    TokenType = "!"
	MUL    TokenType = "*"
	DIV    TokenType = "/"
	GT     TokenType = ">"
	LT     TokenType = "<"
	EQ     TokenType = "=="
	NEQ    TokenType = "!="

	// Delimiters
	COMMA TokenType = ","
	COLON TokenType = ":"
	DOT   TokenType = "."

	LPAREN   TokenType = "("
	RPAREN   TokenType = ")"
	LBRACE   TokenType = "{"
	RBRACE   TokenType = "}"
	LBRACKET TokenType = "["
	RBRACKET TokenType = "]"

	// Keywords
	FUN    TokenType = "FUN"
	TRUE   TokenType = "TRUE"
	FALSE  TokenType = "FALSE"
	IF     TokenType = "IF"
	ELSE   TokenType = "ELSE"
	RETURN TokenType = "RETURN"

	// Indentation
	INDENT TokenType = "INDENT"
	DEDENT TokenType = "DEDENT"

	// Loops
	FOR TokenType = "FOR"
	IN  TokenType = "IN"
)
