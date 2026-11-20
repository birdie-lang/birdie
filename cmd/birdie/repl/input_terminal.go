package repl

import (
	"errors"
	"io"
	"strings"

	"github.com/birdie-lang/birdie/cmd/birdie/lexer"
	"github.com/birdie-lang/birdie/cmd/birdie/token"

	"github.com/reeflective/readline"
	"github.com/reeflective/readline/inputrc"
)

func terminalInput() func(string) (string, error) {
	shell := newTerminalShell()
	return func(primaryPrompt string) (string, error) {
		shell.Prompt.Primary(func() string { return primaryPrompt })
		text, err := shell.Readline()
		if errors.Is(err, io.EOF) && text == "" {
			text = string(*shell.Line())
		}
		return text, err
	}
}

const indentWidth = 4

func newTerminalShell() *readline.Shell {
	shell := readline.NewShell(inputrc.WithApp("birdie"))
	shell.Config.Vars["enable-bracketed-paste"] = true
	shell.Config.Vars["cursor-emacs"] = "beam"
	shell.Config.Vars["multiline-column"] = true
	shell.Config.Vars["multiline-column-custom"] = continuationPrompt
	shell.Prompt.Primary(func() string { return prompt })
	shell.Prompt.Secondary(func() string { return continuationPrompt })

	// The built-in accept-line inserts the newline after this callback returns.
	// Add the indentation afterward so the cursor lands after the spaces.
	var indent string
	shell.AcceptMultiline = func(line []rune) bool {
		if inputComplete(string(line)) {
			return true
		}
		indent = continuationIndent(line[:shell.Cursor().Pos()])
		return false
	}
	acceptLine := shell.Keymap.Commands()["accept-line"]
	shell.Keymap.Register(map[string]func(){
		"birdie-newline": func() {
			indent = ""
			acceptLine()
			shell.Cursor().InsertAt([]rune(indent)...)
		},
		"birdie-indent": func() {
			pos := shell.Cursor().Pos()
			start := lineStart(*shell.Line(), pos)
			spaces := indentWidth - (pos-start)%indentWidth
			shell.Cursor().InsertAt([]rune(strings.Repeat(" ", spaces))...)
		},
		"birdie-unindent": func() {
			line, cursor := shell.Line(), shell.Cursor()
			pos := cursor.Pos()
			start := lineStart(*line, pos)
			end := start
			for end < line.Len() && end-start < indentWidth && (*line)[end] == ' ' {
				end++
			}
			line.Cut(start, end)
			cursor.Set(max(start, pos-(end-start)))
		},
	})
	for _, keymap := range []string{"emacs", "vi-insert"} {
		for key, command := range map[string]string{
			"\r": "birdie-newline", "\n": "birdie-newline",
			"\t": "birdie-indent", "\x1b[Z": "birdie-unindent",
		} {
			shell.Config.Binds[keymap][key] = inputrc.Bind{Action: command}
		}
	}
	return shell
}

func lineStart(line []rune, pos int) int {
	for pos > 0 && line[pos-1] != '\n' {
		pos--
	}
	return pos
}

// continuationIndent preserves the current line's spaces and adds a level
// after a block colon, ignoring quoted colons and those inside delimiters.
func continuationIndent(prefix []rune) string {
	start := lineStart(prefix, len(prefix))
	spaces := 0
	for start+spaces < len(prefix) && prefix[start+spaces] == ' ' {
		spaces++
	}
	l := lexer.New(string(prefix))
	var last token.TokenType
	depth := 0
	for t := l.NextToken(); t.Type != token.EOF; t = l.NextToken() {
		switch t.Type {
		case token.INDENT, token.DEDENT:
			continue
		case token.LPAREN, token.LBRACKET, token.LBRACE:
			depth++
		case token.RPAREN, token.RBRACKET, token.RBRACE:
			depth--
		}
		last = t.Type
	}
	if last == token.COLON && depth == 0 {
		spaces += indentWidth
	}
	return strings.Repeat(" ", spaces)
}
