package source

import "testing"

func TestPosition(t *testing.T) {
	file := NewFile("example.bd", "aé\n\nb")
	tests := []struct {
		offset int
		want   Position
	}{
		{-1, Position{Line: 1, Column: 1}},
		{0, Position{Line: 1, Column: 1}},
		{1, Position{Line: 1, Column: 2}},
		{3, Position{Line: 1, Column: 3}},
		{4, Position{Line: 2, Column: 1}},
		{5, Position{Line: 3, Column: 1}},
		{6, Position{Line: 3, Column: 2}},
		{100, Position{Line: 3, Column: 2}},
	}
	for _, tt := range tests {
		if got := file.Position(tt.offset); got != tt.want {
			t.Errorf("Position(%d) = %+v, want %+v", tt.offset, got, tt.want)
		}
	}
	if got := NewFile("empty.bd", "").Position(0); got != (Position{Line: 1, Column: 1}) {
		t.Errorf("empty file position = %+v", got)
	}
}

func TestLine(t *testing.T) {
	file := NewFile("example.bd", "a\r\n\nb\n")
	for line, want := range []string{"a", "", "b", ""} {
		if got := file.Line(line + 1); got != want {
			t.Errorf("Line(%d) = %q, want %q", line+1, got, want)
		}
	}
	for _, line := range []int{0, 5} {
		if got := file.Line(line); got != "" {
			t.Errorf("out-of-range Line(%d) = %q, want empty string", line, got)
		}
	}
}
