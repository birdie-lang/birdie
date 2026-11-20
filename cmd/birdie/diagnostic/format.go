package diagnostic

import (
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/birdie-lang/birdie/cmd/birdie/source"
)

// Format renders a diagnostic and underlines its first source line. Empty spans
// still receive a caret, including errors pointing at the end of a file.
func Format(file *source.File, d Diagnostic) string {
	position := file.Position(d.Span.Start)
	endPosition := file.Position(d.Span.End)
	line := []rune(file.Line(position.Line))
	start := min(position.Column-1, len(line))
	end := len(line)
	if endPosition.Line == position.Line {
		end = max(start, min(endPosition.Column-1, len(line)))
	}
	column := utf8.RuneCountInString(expandTabs(string(line[:start])))
	endColumn := utf8.RuneCountInString(expandTabs(string(line[:end])))
	number := strconv.Itoa(position.Line)
	return fmt.Sprintf("%s:%d:%d: error[%s]: %s\n%s | %s\n%s | %s%s",
		file.Name, position.Line, position.Column, d.Code, d.Message,
		number, expandTabs(string(line)), strings.Repeat(" ", len(number)),
		strings.Repeat(" ", column), strings.Repeat("^", max(1, endColumn-column)))
}

func expandTabs(text string) string {
	var out strings.Builder
	column := 0
	for _, ch := range text {
		if ch == '\t' {
			spaces := 4 - column%4
			out.WriteString(strings.Repeat(" ", spaces))
			column += spaces
		} else {
			out.WriteRune(ch)
			column++
		}
	}
	return out.String()
}
