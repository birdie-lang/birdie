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
	return FormatWithColor(file, d, false)
}

// FormatWithColor optionally highlights the error, location, and source marker.
func FormatWithColor(file *source.File, d Diagnostic, color bool) string {
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
	indent := strings.Repeat(" ", len(number))
	headline := fmt.Sprintf("error[%s]", d.Code)
	location := fmt.Sprintf("--> %s:%d:%d", file.Name, position.Line, position.Column)
	gutter := indent + " |"
	numberedGutter := number + " |"
	carets := strings.Repeat("^", max(1, endColumn-column))
	if color {
		headline = "\x1b[1;31m" + headline + "\x1b[0m"
		carets = "\x1b[1;31m" + carets + "\x1b[0m"
		location = "\x1b[2m" + location + "\x1b[0m"
		gutter = "\x1b[2m" + gutter + "\x1b[0m"
		numberedGutter = "\x1b[2m" + numberedGutter + "\x1b[0m"
	}
	return fmt.Sprintf("%s: %s\n%s%s\n%s\n%s %s\n%s %s%s",
		headline, d.Message, indent, location, gutter, numberedGutter,
		expandTabs(string(line)), gutter, strings.Repeat(" ", column), carets)
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
