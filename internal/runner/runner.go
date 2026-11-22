package runner

import (
	"github.com/birdie-lang/birdie/cmd/birdie/diagnostic"
	"github.com/birdie-lang/birdie/cmd/birdie/evaluator"
	"github.com/birdie-lang/birdie/cmd/birdie/object"
	"github.com/birdie-lang/birdie/cmd/birdie/parser"
)

// Run parses source and evaluates it in the supplied environment.
func Run(input string, env *object.Environment) (object.Object, []diagnostic.Diagnostic) {
	program, diagnostics := parser.Parse(input)
	if len(diagnostics) != 0 {
		return nil, diagnostics
	}
	value, d := evaluator.Eval(program, env)
	if d != nil {
		return nil, []diagnostic.Diagnostic{*d}
	}
	return value, nil
}
