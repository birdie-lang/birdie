package tests

import (
	"encoding/json"
	"testing"
)

func TestResultComparison(t *testing.T) {
	cases := []struct {
		name, check, actual, expected string
		equal                         bool
	}{
		{"ordered-list", "exact", `[2,1]`, `[1,2]`, false},
		{"unordered-list", "unordered", `[2,1]`, `[1,2]`, true},
		{"duplicates-matter", "unordered", `[1,1,2]`, `[1,2,2]`, false},
		{"unordered-groups", "unordered-groups", `[["b"],["tea","eat"]]`, `[["eat","tea"],["b"]]`, true},
		{"group-duplicates-matter", "unordered-groups", `[["a","a"],["b"]]`, `[["a"],["b"]]`, false},
		{"group-boundaries-matter", "unordered-groups", `[["a","b"],["c"]]`, `[["a"],["b","c"]]`, false},
		{"large-integers", "exact", `9007199254740993`, `9007199254740992`, false},
		{"arbitrary-precision", "exact", `123456789012345678901234567890`, `123456789012345678901234567890`, true},
		{"integer-vs-string", "exact", `123`, `"123"`, false},
		{"negative-zero", "exact", `-0`, `0`, true},
		{"object-key-order", "exact", `{"a":1,"b":2}`, `{"b":2,"a":1}`, true},
		{"unicode-is-not-normalized", "exact", `"é"`, `"e\u0301"`, false},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			actual := resultKey(t, json.RawMessage(test.actual), test.check)
			expected := resultKey(t, json.RawMessage(test.expected), test.check)
			if (actual == expected) != test.equal {
				t.Fatalf("equal=%v; got %s; want %s", test.equal, test.actual, test.expected)
			}
		})
	}
}
