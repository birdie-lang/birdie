package diagnostic

import "github.com/birdie-lang/birdie/cmd/birdie/source"

// Code identifies a diagnostic independently of its human-readable message.
type Code string

const (
	InvalidIndentation      Code = "invalid-indentation"
	TabsInIndentation       Code = "tabs-in-indentation"
	UnexpectedCharacter     Code = "unexpected-character"
	ExpectedExpression      Code = "expected-expression"
	UnexpectedToken         Code = "unexpected-token"
	UnexpectedEOF           Code = "unexpected-eof"
	InvalidAssignmentTarget Code = "invalid-assignment-target"
	InvalidInteger          Code = "invalid-integer"
	DivisionByZero          Code = "division-by-zero"
	UnknownOperator         Code = "unknown-operator"
	IndexOutOfBounds        Code = "index-out-of-bounds"
	UnboundVariable         Code = "unbound-variable"
	NoFunction              Code = "no-function"
	ArgumentCountMismatch   Code = "argument-count-mismatch"
	ExpectedValue           Code = "expected-value"
	BuiltinError            Code = "builtin-error"
)

// Diagnostic describes a source error. Spans refer to the source being analyzed;
// presentation and error recovery belong to the caller and producing stage.
type Diagnostic struct {
	Code    Code
	Message string
	Span    source.Span
}
