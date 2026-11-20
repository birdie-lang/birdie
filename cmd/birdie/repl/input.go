package repl

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/birdie-lang/birdie/cmd/birdie/lexer"
	"github.com/birdie-lang/birdie/cmd/birdie/token"
)

const continuationPrompt = "| "

// inputComplete decides when to submit an entry, not whether it is valid.
// Definite lexical errors and mismatched delimiters are submitted immediately
// so diagnostics are shown instead of asking for more input indefinitely.
func inputComplete(text string) bool {
	// Slash commands are handled by the REPL, even when they contain brackets
	// or colons. They should not enter language continuation mode.
	if strings.HasPrefix(strings.TrimSpace(text), "/") {
		return true
	}
	l := lexer.New(text)
	var closing []token.TokenType
	var last token.TokenType
	block := false
	for t := l.NextToken(); t.Type != token.EOF; t = l.NextToken() {
		switch t.Type {
		case token.ILLEGAL:
			return true
		case token.INDENT, token.DEDENT:
			continue
		case token.EOL:
			if last == token.COLON {
				block = true
			}
			last = ""
			continue
		case token.LPAREN:
			closing = append(closing, token.RPAREN)
		case token.LBRACKET:
			closing = append(closing, token.RBRACKET)
		case token.LBRACE:
			closing = append(closing, token.RBRACE)
		case token.RPAREN, token.RBRACKET, token.RBRACE:
			if len(closing) == 0 || closing[len(closing)-1] != t.Type {
				return true
			}
			closing = closing[:len(closing)-1]
		}
		last = t.Type
	}
	if len(l.Diagnostics()) != 0 {
		return true
	}
	if len(closing) != 0 {
		return false
	}
	if !block && last != token.COLON {
		return true
	}
	lastLine := text[strings.LastIndex(text, "\n")+1:]
	return strings.TrimSpace(lastLine) == ""
}

// streamInput uses the same continuation rules as the terminal editor. It keeps
// supplied streams usable for pipes and deterministic REPL tests.
func streamInput(in io.Reader, out io.Writer) func(string) (string, error) {
	reader := bufio.NewReader(in)
	return func(primaryPrompt string) (string, error) {
		var input strings.Builder
		nextPrompt := primaryPrompt
		for {
			fmt.Fprint(out, nextPrompt)
			line, err := reader.ReadString('\n')
			if line == "" && err != nil {
				return input.String(), err
			}
			if input.Len() != 0 {
				input.WriteByte('\n')
			}
			line = strings.TrimSuffix(strings.TrimSuffix(line, "\n"), "\r")
			input.WriteString(line)
			text := input.String()
			if err != nil && !errors.Is(err, io.EOF) {
				return text, err
			}
			if inputComplete(text) {
				return text, nil
			}
			if err != nil {
				return text, err
			}
			nextPrompt = continuationPrompt
		}
	}
}
