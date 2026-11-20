package source

import (
	"sort"
	"unicode/utf8"
)

// Span is a half-open byte range [Start, End) in one source file.
// A zero-width span identifies a location such as EOF or a synthetic token.
type Span struct {
	Start int
	End   int
}

// Position uses one-based lines and Unicode code point columns for display.
// Spans remain byte-based; editor protocol positions can be derived separately.
type Position struct {
	Line   int
	Column int
}

// File represents a source snapshot. Create a new File when its text changes.
type File struct {
	Name string
	Text string

	lineStarts []int
}

func NewFile(name, text string) *File {
	f := &File{Name: name, Text: text, lineStarts: []int{0}}
	for i := range text {
		if text[i] == '\n' {
			f.lineStarts = append(f.lineStarts, i+1)
		}
	}
	return f
}

func (f *File) Position(offset int) Position {
	offset = min(max(offset, 0), len(f.Text))
	line := sort.Search(len(f.lineStarts), func(i int) bool {
		return f.lineStarts[i] > offset
	}) - 1
	return Position{
		Line:   line + 1,
		Column: utf8.RuneCountInString(f.Text[f.lineStarts[line]:offset]) + 1,
	}
}

// Line returns a one-based line without its line ending.
func (f *File) Line(line int) string {
	if line < 1 || line > len(f.lineStarts) {
		return ""
	}
	start := f.lineStarts[line-1]
	end := len(f.Text)
	if line < len(f.lineStarts) {
		end = f.lineStarts[line] - 1
	}
	if end > start && f.Text[end-1] == '\r' {
		end--
	}
	return f.Text[start:end]
}
