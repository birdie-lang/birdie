package lexer

import (
	"github.com/birdie-lang/birdie/cmd/birdie/token"
)

type Lexer struct {
	input        string
	position     int
	readPosition int
	ch           byte

	indentLevel    int
	atLineStart    bool
	pendingDedents int
	delimiterDepth int

	errors []string
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
	} else {
		l.ch = l.input[l.readPosition]
	}
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
		tok.Type = token.DIV
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
		l.delimiterDepth--
	case ',':
		tok.Type = token.COMMA
	case '+':
		tok.Type = token.PLUS
	case '{':
		tok.Type = token.LBRACE
		l.delimiterDepth++
	case '}':
		tok.Type = token.RBRACE
		l.delimiterDepth--
	case '"':
		tok.Type = token.STRING
		tok.Literal = l.readString()
	case '[':
		tok.Type = token.LBRACKET
		l.delimiterDepth++
	case ']':
		tok.Type = token.RBRACKET
		l.delimiterDepth--
	case '\n':
		l.atLineStart = true
		tok.Type = token.EOL
	case 0:
		if l.indentLevel > 0 {
			l.indentLevel--
			return token.Token{Type: token.DEDENT}
		}
		return token.Token{Type: token.EOF}
	default:
		if isIdentifierStart(l.ch) {
			literal := l.readIdentifier()
			tok.Type = token.LookupIdent(literal)
			if tok.Type == token.IDENT {
				tok.Literal = literal
			}
			return tok
		} else if isDigit(l.ch) {
			tok.Literal = l.readNumber()
			tok.Type = token.INT
			return tok
		} else {
			tok = token.Token{Type: token.ILLEGAL, Literal: string(l.ch)}
		}
	}

	l.readChar()
	return tok
}

func (l *Lexer) handleIndentation() (token.Token, bool) {
	if l.pendingDedents > 0 {
		l.pendingDedents--
		return token.Token{Type: token.DEDENT}, true
	}

	if !l.atLineStart {
		return token.Token{}, false
	}

	if l.delimiterDepth > 0 {
		for l.ch == ' ' || l.ch == '\n' {
			l.readChar()
		}
		return token.Token{}, false
	}

	spaces := 0
	for {
		spaces = 0
		for l.ch == ' ' {
			spaces++
			l.readChar()
		}

		if l.ch == '\t' {
			l.errors = append(l.errors, "tabs are not allowed in indentation")
			return token.Token{}, false
		}

		if l.ch != '\n' {
			break
		}

		// Blank line: consume its newline and inspect the next line
		l.readChar()
	}

	l.atLineStart = false

	if l.ch == 0 {
		return token.Token{}, false
	}

	if spaces%4 != 0 {
		l.errors = append(l.errors, "Wrong indentation")
		return token.Token{Type: token.ILLEGAL}, false
	}

	level := spaces / 4
	if level == l.indentLevel {
		return token.Token{}, false
	} else if level < l.indentLevel {
		l.pendingDedents = l.indentLevel - level - 1
		l.indentLevel = level
		return token.Token{Type: token.DEDENT}, true
	} else if level == l.indentLevel+1 {
		l.indentLevel = level
		return token.Token{Type: token.INDENT}, true
	} else {
		l.errors = append(l.errors, "Wrong indentation")
		return token.Token{Type: token.ILLEGAL}, false
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

func (l *Lexer) Errors() []string {
	return l.errors
}
