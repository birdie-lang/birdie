package repl

import (
	"strings"
	"testing"
	"unicode/utf8"
)

func TestTerminalEditing(t *testing.T) {
	tests := []struct {
		name, key, input, want string
	}{
		{"block-header", "\r", "fn add(a: int):|", "fn add(a: int):\n    |"},
		{"nested-block", "\r", "if true:\n    if false:|", "if true:\n    if false:\n        |"},
		{"preserve-indent", "\r", "if true:\n    value = 1|", "if true:\n    value = 1\n    |"},
		{"multiline-header", "\r", "fn add(\n    a: int\n):|", "fn add(\n    a: int\n):\n    |"},
		{"quoted-colon", "\r", "if true:\n    text = \"value:\"|", "if true:\n    text = \"value:\"\n    |"},
		{"map-colon", "\r", "values = {1:|", "values = {1:\n|"},
		{"parameter-colon", "\r", "fn add(a:|", "fn add(a:\n|"},
		{"delimiter-indent", "\r", "values = [\n    1,|", "values = [\n    1,\n    |"},
		{"split-line", "\r", "if true:\n    |value = 1", "if true:\n    \n    |value = 1"},
		{"line-feed", "\n", "if true:|", "if true:\n    |"},
		{"tab-empty", "\t", "|", "    |"},
		{"tab-partial-indent", "\t", "  |", "    |"},
		{"tab-full-indent", "\t", "    |", "        |"},
		{"tab-current-line", "\t", "if true:\n    |value", "if true:\n        |value"},
		{"tab-unicode", "\t", "\"é\"|", "\"é\" |"},
		{"unindent", "\x1b[Z", "        |", "    |"},
		{"unindent-current-line", "\x1b[Z", "if true:\n    value|", "if true:\nvalue|"},
		{"unindent-partial", "\x1b[Z", "  |", "|"},
		{"unindent-at-start", "\x1b[Z", "|    value", "|value"},
		{"unindent-inside-spaces", "\x1b[Z", "  |      value", "|    value"},
		{"unindent-no-spaces", "\x1b[Z", "value|", "value|"},
	}
	for _, keymap := range []string{"emacs", "vi-insert"} {
		t.Run(keymap, func(t *testing.T) {
			shell := newTerminalShell()
			for _, tt := range tests {
				t.Run(tt.name, func(t *testing.T) {
					// A pipe marks the cursor in each editor buffer.
					before, after, ok := strings.Cut(tt.input, "|")
					if !ok {
						t.Fatal("test input needs a cursor marker")
					}
					shell.Line().Set([]rune(before + after)...)
					shell.Cursor().Set(utf8.RuneCountInString(before))
					bind := shell.Config.Binds[keymap][tt.key]
					command := shell.Keymap.Commands()[bind.Action]
					if command == nil {
						t.Fatalf("no command bound to %q", tt.key)
					}
					command()
					line, pos := *shell.Line(), shell.Cursor().Pos()
					got := string(line[:pos]) + "|" + string(line[pos:])
					if got != tt.want {
						t.Fatalf("edited buffer = %q, want %q", got, tt.want)
					}
				})
			}
		})
	}
}

func TestAutoIndentedBlankLineCompletesInput(t *testing.T) {
	shell := newTerminalShell()
	shell.Cursor().InsertAt([]rune("fn example():")...)
	newline := shell.Keymap.Commands()["birdie-newline"]
	newline()
	shell.Cursor().InsertAt([]rune("return 1")...)
	newline()
	if got, want := string(*shell.Line()), "fn example():\n    return 1\n    "; got != want {
		t.Fatalf("auto-indented buffer = %q, want %q", got, want)
	}
	if !shell.AcceptMultiline(*shell.Line()) {
		t.Fatal("auto-inserted spaces prevented blank-line submission")
	}
}
