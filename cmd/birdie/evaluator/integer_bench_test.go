package evaluator

import (
	"testing"

	"github.com/birdie-lang/birdie/cmd/birdie/object"
	"github.com/birdie-lang/birdie/cmd/birdie/parser"
	"github.com/birdie-lang/birdie/internal/testutil"
)

func BenchmarkEvalIntegers(b *testing.B) {
	const factorial = `fn factorial(n: int) int:
    if n == 0:
        return 1
    return n * factorial(n - 1)
factorial(20)`
	tests := []struct{ name, input, want string }{
		{"SmallAdd", "12345 + 67890", "80235"},
		{"SmallMul", "12345 * 67890", "838102050"},
		{"FloorDivide", "-100 // 7", "-15"},
		{"Compare", "12345 < 67890", "true"},
		{"OverflowAdd", "9223372036854775807 + 1", "9223372036854775808"},
		{"MixedAdd", "1267650600228229401496703205376 + 4", "1267650600228229401496703205380"},
		{"Factorial20", factorial, "2432902008176640000"},
		{"ListLiteralAndIndex", "[[10, 20], [30]][0][1]", "20"},
	}
	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			program, diagnostics := parser.Parse(tt.input)
			testutil.RequireNoDiagnostics(b, "parser", tt.input, diagnostics...)
			env := object.NewEnvironment()
			var value object.Object
			for b.Loop() {
				result, d := Eval(program, env)
				if d != nil {
					b.Fatal(d.Message)
				}
				value = result
			}
			if value == nil || value.Inspect() != tt.want {
				b.Fatalf("got %v, want %s", value, tt.want)
			}
		})
	}
}
