package repl

import (
	"cmp"
	"errors"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"

	"github.com/reeflective/readline"

	"github.com/birdie-lang/birdie/cmd/birdie/ast"
	"github.com/birdie-lang/birdie/cmd/birdie/diagnostic"
	"github.com/birdie-lang/birdie/cmd/birdie/evaluator"
	"github.com/birdie-lang/birdie/cmd/birdie/lexer"
	"github.com/birdie-lang/birdie/cmd/birdie/object"
	"github.com/birdie-lang/birdie/cmd/birdie/parser"
	"github.com/birdie-lang/birdie/cmd/birdie/source"
	"github.com/birdie-lang/birdie/cmd/birdie/token"
)

const prompt = "> "

type mode string

const (
	modeTokens mode = "tokens"
	modeParse  mode = "parse"
	modeEval   mode = "eval"
)

func (m mode) prompt() string {
	if m == modeEval {
		return prompt
	}
	return string(m) + "> "
}

const welcome = "Birdie\n/help for commands. Ctrl+D to exit.\nEmpty line finishes a block.\n\n"

var commandHelp = []struct{ name, description string }{
	{"/eval", "Evaluate input (default)"},
	{"/tokens", "Show lexer tokens"},
	{"/parse", "Show the parsed AST"},
	{"/help", "Show commands and editing shortcuts"},
	{"/exit, /quit", "Exit the REPL"},
}

var editingHelp = []struct{ name, description string }{
	{"Enter", "Submit input or continue an unfinished entry"},
	{"Empty line", "Submit an indented block"},
	{"Tab", "Insert spaces to the next four-space boundary"},
	{"Shift+Tab", "Remove one indentation level"},
	{"Up / Down", "Navigate multiline input and session history"},
	{"Ctrl+C", "Cancel the current entry"},
	{"Ctrl+D", "Exit when input is empty"},
}

func Start(in io.Reader, out io.Writer) {
	readInput := streamInput(in, out)
	// Readline uses the process standard streams. Other streams retain the
	// plain reader, including redirected input/output and tests.
	if in == os.Stdin && out == os.Stdout && isTerminal(os.Stdin) && isTerminal(os.Stdout) {
		fmt.Fprint(out, welcome)
		if os.Getenv("TERM") != "dumb" {
			readInput = terminalInput()
		}
	}
	run(readInput, out)
}

func run(readInput func(string) (string, error), out io.Writer) {
	env := object.NewEnvironment()
	m := modeEval

	for {
		text, err := readInput(m.prompt())
		atEOF := errors.Is(err, io.EOF)
		if errors.Is(err, readline.ErrInterrupt) {
			continue
		}
		if atEOF {
			if text == "" {
				return
			}
			// EOF can finish a block, just like its final empty line. Open
			// delimiters still require a closing token and must not execute.
			if !inputComplete(text + "\n") {
				printDiagnostics(text, []diagnostic.Diagnostic{{
					Code: diagnostic.UnexpectedEOF, Message: "unexpected end of input in a multiline entry",
					Span: source.Span{Start: len(text), End: len(text)},
				}}, out)
				return
			}
		} else if err != nil {
			fmt.Fprintf(out, "input error: %s\n", err)
			return
		}

		command := strings.TrimSpace(text)
		if command == "/exit" || command == "/quit" {
			fmt.Fprintln(out, "Goodbye!")
			return
		}

		// Handle commands
		if handled, newMode := handleCommand(command, m, out); handled {
			m = newMode
			if atEOF {
				return
			}
			continue
		}

		// Process based on mode
		switch m {
		case modeTokens:
			printTokens(text, out)
		case modeParse:
			printAST(text, out)
		case modeEval:
			printEval(text, out, env)
		}
		if atEOF {
			return
		}
	}
}

func handleCommand(line string, current mode, out io.Writer) (bool, mode) {
	switch line {
	case "/tokens":
		return true, modeTokens
	case "/parse":
		return true, modeParse
	case "/eval":
		return true, modeEval
	case "/help":
		fmt.Fprintln(out, "Commands:")
		for _, cmd := range commandHelp {
			fmt.Fprintf(out, "  %-14s %s\n", cmd.name, cmd.description)
		}
		fmt.Fprintln(out, "\nEditing:")
		for _, key := range editingHelp {
			fmt.Fprintf(out, "  %-14s %s\n", key.name, key.description)
		}
		return true, current
	}
	if strings.HasPrefix(line, "/") {
		fmt.Fprintf(out, "Unknown command %q. Use /help for commands.\n", line)
		return true, current
	}
	return false, current
}

func parseLine(line string, out io.Writer) (*ast.Program, bool) {
	l := lexer.New(line)
	p := parser.New(l)
	program := p.ParseProgram()

	diagnostics := slices.Concat(l.Diagnostics(), p.Diagnostics())
	if len(diagnostics) != 0 {
		printDiagnostics(line, diagnostics, out)
		return nil, false
	}

	return program, true
}

func printTokens(line string, out io.Writer) {
	l := lexer.New(line)
	for tok := l.NextToken(); tok.Type != token.EOF; tok = l.NextToken() {
		fmt.Fprintf(out, "%+v\n", tok)
	}
	printDiagnostics(line, l.Diagnostics(), out)
}

func printDiagnostics(text string, diagnostics []diagnostic.Diagnostic, out io.Writer) {
	if len(diagnostics) == 0 {
		return
	}
	slices.SortStableFunc(diagnostics, func(a, b diagnostic.Diagnostic) int {
		return cmp.Compare(a.Span.Start, b.Span.Start)
	})
	file := source.NewFile("<repl>", text)
	color := diagnosticColorEnabled(out)
	for _, d := range diagnostics {
		fmt.Fprintln(out, diagnostic.FormatWithColor(file, d, color))
	}
}

func diagnosticColorEnabled(out io.Writer) bool {
	file, ok := out.(*os.File)
	return ok && os.Getenv("NO_COLOR") == "" && os.Getenv("TERM") != "dumb" && isTerminal(file)
}

func printAST(line string, out io.Writer) {
	program, ok := parseLine(line, out)
	if !ok {
		return
	}

	fmt.Fprintln(out, program.String())
}

func printEval(line string, out io.Writer, env *object.Environment) {
	program, ok := parseLine(line, out)
	if !ok {
		return
	}

	value, d := evaluator.Eval(program, env)
	if d != nil {
		printDiagnostics(line, []diagnostic.Diagnostic{*d}, out)
		return
	}
	if value != nil {
		fmt.Fprintln(out, value.Inspect())
	}
}
