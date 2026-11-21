package testutil

import (
	"strings"
	"testing"

	"github.com/birdie-lang/birdie/cmd/birdie/diagnostic"
	"github.com/birdie-lang/birdie/cmd/birdie/source"
)

// RequireNoDiagnostics reports unexpected diagnostics with their source context.
func RequireNoDiagnostics(t *testing.T, stage, input string, diagnostics ...diagnostic.Diagnostic) {
	t.Helper()
	if len(diagnostics) == 0 {
		return
	}

	file := source.NewFile("<test>", input)
	messages := make([]string, 0, len(diagnostics))
	for _, d := range diagnostics {
		messages = append(messages, diagnostic.Format(file, d))
	}
	t.Fatalf("unexpected %s diagnostics:\n%s", stage, strings.Join(messages, "\n\n"))
}
