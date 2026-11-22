package cli

import (
	"cmp"
	"errors"
	"fmt"
	"io"
	"os"
	"runtime/debug"
	"slices"
	"strings"

	"github.com/birdie-lang/birdie/cmd/birdie/diagnostic"
	"github.com/birdie-lang/birdie/cmd/birdie/evaluator"
	"github.com/birdie-lang/birdie/cmd/birdie/repl"
	"github.com/birdie-lang/birdie/cmd/birdie/source"
	"github.com/birdie-lang/birdie/internal/runner"
	"github.com/birdie-lang/birdie/internal/terminal"
	"github.com/spf13/cobra"
)

// Execute runs the CLI with the supplied streams and returns its exit status.
func Execute(args []string, in io.Reader, out, errOut io.Writer) int {
	startREPL := func(cmd *cobra.Command, _ []string) {
		repl.Start(cmd.InOrStdin(), cmd.OutOrStdout())
	}
	root := &cobra.Command{
		Use:           "birdie",
		Short:         "Run Birdie programs or start an interactive session",
		Version:       version(),
		Args:          cobra.NoArgs,
		Run:           startREPL,
		SilenceErrors: true,
		SilenceUsage:  true,
	}
	root.AddCommand(&cobra.Command{
		Use:   "repl",
		Short: "Start an interactive Birdie session",
		Args:  cobra.NoArgs,
		Run:   startREPL,
	}, &cobra.Command{
		Use:   "run <file>",
		Short: "Run a Birdie file (use - to read from stdin)",
		Example: "  birdie run factorial.bd\n" +
			"  birdie run - < factorial.bd",
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.ExactArgs(1)(cmd, args); err != nil {
				return fmt.Errorf("%w\nUsage: %s", err, cmd.UseLine())
			}
			return nil
		},
		ValidArgsFunction: cobra.FixedCompletions([]string{"bd"}, cobra.ShellCompDirectiveFilterFileExt),
		RunE:              run,
	})
	root.SetArgs(append([]string{}, args...))
	root.SetIn(in)
	root.SetOut(out)
	root.SetErr(errOut)

	if err := root.Execute(); err != nil {
		root.PrintErrln(err)
		return 1
	}
	return 0
}

func run(cmd *cobra.Command, args []string) error {
	name := args[0]
	var input []byte
	var err error
	if name == "-" {
		name = "<stdin>"
		input, err = io.ReadAll(cmd.InOrStdin())
	} else {
		input, err = os.ReadFile(name)
	}
	if err != nil {
		return fmt.Errorf("read source: %w", err)
	}

	text := string(input)
	_, diagnostics := runner.Run(text, evaluator.NewEnvironment(cmd.OutOrStdout()))
	if len(diagnostics) == 0 {
		return nil
	}

	slices.SortStableFunc(diagnostics, func(a, b diagnostic.Diagnostic) int {
		return cmp.Compare(a.Span.Start, b.Span.Start)
	})
	file := source.NewFile(name, text)
	color := terminal.ColorEnabled(cmd.ErrOrStderr())
	messages := make([]string, len(diagnostics))
	for i, d := range diagnostics {
		messages[i] = diagnostic.FormatWithColor(file, d, color)
	}
	return errors.New(strings.Join(messages, "\n"))
}

func version() string {
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
		return info.Main.Version
	}
	return "dev"
}
