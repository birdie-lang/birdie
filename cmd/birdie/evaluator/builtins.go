package evaluator

import (
	"fmt"
	"io"

	"github.com/birdie-lang/birdie/cmd/birdie/object"
)

// NewEnvironment creates a session with built-ins that write to out.
func NewEnvironment(out io.Writer) *object.Environment {
	env := object.NewEnvironment()
	env.Set("print", &object.Builtin{
		Name:  "print",
		Arity: 1,
		Fn: func(args ...object.Object) (object.Object, error) {
			_, err := fmt.Fprintln(out, args[0].Inspect())
			return nil, err
		},
	})
	return env
}
