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
	InvalidAssignmentTarget Code = "invalid-assignment-target"
	InvalidInteger          Code = "invalid-integer"
	DivisionByZero          Code = "division-by-zero"
	UnknownOperator         Code = "unknown-operator"
	UnboundVariable         Code = "unbound-variable"
)

// Diagnostic describes a source error. Spans refer to the source being analyzed;
// presentation and error recovery belong to the caller and producing stage.
type Diagnostic struct {
	Code    Code
	Message string
	Span    source.Span
}
