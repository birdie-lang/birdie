package lexer

import (
	"fmt"
	"unicode/utf8"

	"github.com/birdie-lang/birdie/cmd/birdie/diagnostic"
	"github.com/birdie-lang/birdie/cmd/birdie/source"
	"github.com/birdie-lang/birdie/cmd/birdie/token"
)

const indentWidth = 4

type Lexer struct {
	input        string
	position     int
	readPosition int
	ch           byte

	indentLevel    int
	atLineStart    bool
	pendingDedents int
	delimiterDepth int

	diagnostics []diagnostic.Diagnostic
}

// Creates a new lexer with the given input
func New(input string) *Lexer {
	l := &Lexer{input: input, atLineStart: true}
	l.readChar()
	return l
}

// Reads and consumes the next character from the input
func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0
		l.position = len(l.input)
		return
	}
	l.ch = l.input[l.readPosition]
	l.position = l.readPosition
	l.readPosition += 1
}

// Returns the next token in the input
func (l *Lexer) NextToken() token.Token {
	var tok token.Token

	if tok, ok := l.handleIndentation(); ok {
		return tok
	}
	l.skipWhitespace()
	start := l.position

	switch l.ch {
	case '=':
		if l.peekChar() == '=' {
			l.readChar()
			tok.Type = token.EQ
		} else {
			tok.Type = token.ASSIGN
		}
	case '-':
		tok.Type = token.MINUS
	case '*':
		tok.Type = token.MUL
	case '/':
		if l.peekChar() == '/' {
			l.readChar()
			tok.Type = token.DDIV
		} else {
			tok.Type = token.DIV
		}
	case '!':
		if l.peekChar() == '=' {
			l.readChar()
			tok.Type = token.NEQ
		} else {
			tok.Type = token.NOT
		}
	case '<':
		tok.Type = token.LT
	case '>':
		tok.Type = token.GT
	case ':':
		tok.Type = token.COLON
	case '.':
		tok.Type = token.DOT
	case '(':
		tok.Type = token.LPAREN
		l.delimiterDepth++
	case ')':
		tok.Type = token.RPAREN
		l.delimiterDepth = max(0, l.delimiterDepth-1)
	case ',':
		tok.Type = token.COMMA
	case '+':
		tok.Type = token.PLUS
	case '{':
		tok.Type = token.LBRACE
		l.delimiterDepth++
	case '}':
		tok.Type = token.RBRACE
		l.delimiterDepth = max(0, l.delimiterDepth-1)
	case '"':
		tok.Type = token.STRING
		tok.Literal = l.readString()
	case '[':
		tok.Type = token.LBRACKET
		l.delimiterDepth++
	case ']':
		tok.Type = token.RBRACKET
		l.delimiterDepth = max(0, l.delimiterDepth-1)
	case '\n':
		l.atLineStart = true
		tok.Type = token.EOL
	case 0:
		tok.Span = source.Span{Start: start, End: start}
		if l.indentLevel > 0 {
			l.indentLevel--
			tok.Type = token.DEDENT
		} else {
			tok.Type = token.EOF
		}
		return tok
	default:
		if isIdentifierStart(l.ch) {
			literal := l.readIdentifier()
			tok.Type = token.LookupIdent(literal)
			if tok.Type == token.IDENT {
				tok.Literal = literal
			}
			tok.Span = source.Span{Start: start, End: l.position}
			return tok
		} else if isDigit(l.ch) {
			tok.Literal = l.readNumber()
			tok.Type = token.INT
			tok.Span = source.Span{Start: start, End: l.position}
			return tok
		} else {
			_, width := utf8.DecodeRuneInString(l.input[start:])
			literal := l.input[start : start+width]
			for range width {
				l.readChar()
			}
			tok = token.Token{Type: token.ILLEGAL, Literal: literal, Span: source.Span{Start: start, End: l.position}}
			l.diagnostics = append(l.diagnostics, diagnostic.Diagnostic{
				Code: diagnostic.UnexpectedCharacter, Message: fmt.Sprintf("unexpected character %q", literal), Span: tok.Span,
			})
			return tok
		}
	}

	l.readChar()
	tok.Span = source.Span{Start: start, End: l.position}
	return tok
}

func (l *Lexer) handleIndentation() (token.Token, bool) {
	if l.pendingDedents > 0 {
		l.pendingDedents--
		return token.Token{Type: token.DEDENT, Span: source.Span{Start: l.position, End: l.position}}, true
	}

	if !l.atLineStart || l.delimiterDepth > 0 {
		return token.Token{}, false
	}

	var start, spaces, tabPosition int
	for {
		start, spaces, tabPosition = l.position, 0, -1
		for l.ch == ' ' || l.ch == '\t' {
			if l.ch == '\t' {
				if tabPosition == -1 {
					tabPosition = l.position
				}
			} else {
				spaces++
			}
			l.readChar()
		}

		if tabPosition != -1 {
			l.diagnostics = append(l.diagnostics, diagnostic.Diagnostic{
				Code: diagnostic.TabsInIndentation, Message: "tabs are not allowed in indentation",
				Span: source.Span{Start: tabPosition, End: tabPosition + 1},
			})
		}

		if l.ch != '\n' {
			break
		}

		// Blank line: consume its newline and inspect the next line
		l.readChar()
	}

	l.atLineStart = false

	if l.ch == 0 || tabPosition != -1 {
		return token.Token{}, false
	}

	if spaces%indentWidth != 0 {
		l.diagnostics = append(l.diagnostics, diagnostic.Diagnostic{
			Code: diagnostic.InvalidIndentation, Message: "indentation must use multiples of four spaces",
			Span: source.Span{Start: start, End: l.position},
		})
		return token.Token{}, false
	}

	level := spaces / indentWidth
	span := source.Span{Start: l.position, End: l.position}
	if level == l.indentLevel {
		return token.Token{}, false
	} else if level < l.indentLevel {
		l.pendingDedents = l.indentLevel - level - 1
		l.indentLevel = level
		return token.Token{Type: token.DEDENT, Span: span}, true
	} else if level == l.indentLevel+1 {
		l.indentLevel = level
		return token.Token{Type: token.INDENT, Span: span}, true
	} else {
		l.diagnostics = append(l.diagnostics, diagnostic.Diagnostic{
			Code: diagnostic.InvalidIndentation, Message: "indentation can only increase by four spaces at a time",
			Span: source.Span{Start: start, End: l.position},
		})
		return token.Token{}, false
	}
}

func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || (l.ch == '\n' && l.delimiterDepth > 0) {
		l.readChar()
	}
}

func (l *Lexer) peekChar() byte {
	if l.readPosition < len(l.input) {
		return l.input[l.readPosition]
	}
	return 0
}

// Checks if the current character is a valid identifier start
func isIdentifierStart(ch byte) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_'
}

// Checks if the current character is a valid identifier continuation
func isIdentifierContinue(ch byte) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || '0' <= ch && ch <= '9' || ch == '_'
}

// Reads the next character from the input
func (l *Lexer) readIdentifier() string {
	position := l.position
	for isIdentifierContinue(l.ch) {
		l.readChar()
	}
	return l.input[position:l.position]
}

// Reads string from the input
func (l *Lexer) readString() string {
	position := l.position + 1
	for {
		l.readChar()
		if l.ch == '"' || l.ch == 0 {
			break
		}
	}
	return l.input[position:l.position]
}

// Checks if the current character is a digit
func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}

func (l *Lexer) readNumber() string {
	position := l.position
	for isDigit(l.ch) {
		l.readChar()
	}
	return l.input[position:l.position]
}

func (l *Lexer) Diagnostics() []diagnostic.Diagnostic {
	return l.diagnostics
}
