package main

import (
	"encoding/json"
	"syscall/js"

	"github.com/birdie-lang/birdie/cmd/birdie/diagnostic"
	"github.com/birdie-lang/birdie/cmd/birdie/object"
	"github.com/birdie-lang/birdie/internal/runner"
)

func main() {
	run := js.FuncOf(func(_ js.Value, args []js.Value) any {
		if len(args) != 1 {
			return "Invalid no of arguments passed"
		}
		return runSource(args[0].String())
	})

	js.Global().Set("runBirdie", run)
	select {}
}

func runSource(input string) string {
	value, diagnostics := runner.Run(input, object.NewEnvironment())
	return encodeResult(value, diagnostics)
}

func encodeResult(value object.Object, diagnostics []diagnostic.Diagnostic) string {
	result := struct {
		Value       *string             `json:"value"`
		Diagnostics []browserDiagnostic `json:"diagnostics"`
	}{
		Diagnostics: make([]browserDiagnostic, 0, len(diagnostics)),
	}

	if value != nil {
		text := value.Inspect()
		result.Value = &text
	}

	for _, d := range diagnostics {
		result.Diagnostics = append(result.Diagnostics, browserDiagnostic{
			Code:      d.Code,
			Message:   d.Message,
			StartByte: d.Span.Start,
			EndByte:   d.Span.End,
		})
	}

	data, err := json.Marshal(result)
	if err != nil {
		return `{"error":"could not encode Birdie result"}`
	}
	return string(data)
}

type browserDiagnostic struct {
	Code      diagnostic.Code `json:"code"`
	Message   string          `json:"message"`
	StartByte int             `json:"startByte"`
	EndByte   int             `json:"endByte"`
}
