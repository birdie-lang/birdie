package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"math/big"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"testing"
	"time"
)

type fixture struct {
	Entry         string        `json:"entry"`
	Check         string        `json:"check,omitempty"`
	ArgAdapters   []string      `json:"argAdapters,omitempty"`
	ResultAdapter string        `json:"resultAdapter,omitempty"`
	Cases         []exampleCase `json:"cases"`
}

type exampleCase struct {
	Name     string            `json:"name"`
	Args     []json.RawMessage `json:"args"`
	Expected json.RawMessage   `json:"expected"`
}

type example struct {
	Directory string
	fixture
}

var codecDecoders = map[string]string{
	"encode":    "decode",
	"serialize": "deserialize",
}

// Fixture checks do not claim Birdie parsing or execution coverage.
func TestExampleFixtures(t *testing.T) {
	examples := loadExamples(t)
	count := 0
	for _, example := range examples {
		count += len(example.Cases)
	}
	t.Logf("validated %d cases for %d examples; Birdie execution is not implemented", count, len(examples))
}

func TestPythonExamples(t *testing.T) {
	python := os.Getenv("BIRDIE_TEST_PYTHON")
	if python == "" {
		t.Skip("set BIRDIE_TEST_PYTHON to a Python 3.10+ executable to check references")
	}
	python, err := exec.LookPath(python)
	if err != nil {
		t.Fatal(err)
	}
	for _, example := range loadExamples(t) {
		t.Run(filepath.Base(example.Directory), func(t *testing.T) {
			for _, test := range example.Cases {
				t.Run(test.Name, func(t *testing.T) {
					suite := example.fixture
					if suite.Check == "codec-round-trip" {
						// The result adapter belongs to the decoded value, not its wire string.
						suite.ResultAdapter = ""
					}
					actual := runPython(t, python, example.Directory, suite, test.Args)
					if example.Check == "codec-round-trip" {
						var encoded *string
						if err := json.Unmarshal(actual, &encoded); err != nil || encoded == nil {
							t.Fatalf("%s must return a string, got %s", example.Entry, actual)
						}
						decoder := fixture{Entry: codecDecoders[example.Entry], ResultAdapter: example.ResultAdapter}
						actual = runPython(t, python, example.Directory, decoder, []json.RawMessage{actual})
					}
					if resultKey(t, actual, example.Check) != resultKey(t, test.Expected, example.Check) {
						t.Fatalf("got %s; want %s", actual, test.Expected)
					}
				})
			}
		})
	}
}

func loadExamples(t *testing.T) []example {
	t.Helper()
	root := filepath.Join("..", "examples", "leetcode")
	entries, err := os.ReadDir(root)
	if err != nil {
		t.Fatal(err)
	}
	var examples []example
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		directory := filepath.Join(root, entry.Name())
		birdie, bdErr := os.ReadFile(filepath.Join(directory, "solution.bd"))
		python, pyErr := os.ReadFile(filepath.Join(directory, "solution.py"))
		if os.IsNotExist(bdErr) && os.IsNotExist(pyErr) {
			continue
		}
		if bdErr != nil || pyErr != nil || len(bytes.TrimSpace(birdie)) == 0 || len(bytes.TrimSpace(python)) == 0 {
			t.Fatalf("%s requires nonempty solution.bd and solution.py: %v; %v", directory, bdErr, pyErr)
		}
		data, err := os.ReadFile(filepath.Join(directory, "tests.json"))
		if err != nil {
			t.Fatal(err)
		}
		var suite fixture
		decoder := json.NewDecoder(bytes.NewReader(data))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&suite); err != nil || !json.Valid(data) {
			t.Fatalf("%s/tests.json: invalid fixture JSON: %v", directory, err)
		}
		if suite.Entry == "" || len(suite.Cases) == 0 {
			t.Fatalf("%s requires an entry and at least one case", directory)
		}
		if suite.Check == "codec-round-trip" && codecDecoders[suite.Entry] == "" {
			t.Fatalf("%s: codec entry must be encode or serialize", directory)
		}
		for _, adapter := range append(slices.Clone(suite.ArgAdapters), suite.ResultAdapter) {
			if adapter != "" && adapter != "json" && adapter != "tree" && adapter != "node" {
				t.Fatalf("%s: unknown adapter %q", directory, adapter)
			}
		}
		names := make(map[string]bool)
		for _, test := range suite.Cases {
			if test.Name == "" || names[test.Name] || test.Args == nil || test.Expected == nil {
				t.Fatalf("%s: cases need unique names, args, and expected results", directory)
			}
			names[test.Name] = true
			if len(test.Args) != len(suite.Cases[0].Args) || (len(suite.ArgAdapters) != 0 && len(suite.ArgAdapters) != len(test.Args)) {
				t.Fatalf("%s/%s: inconsistent argument or adapter count", directory, test.Name)
			}
			resultKey(t, test.Expected, suite.Check)
			if suite.Check == "codec-round-trip" && (len(test.Args) != 1 || resultKey(t, test.Args[0], "exact") != resultKey(t, test.Expected, "exact")) {
				t.Fatalf("%s/%s: codec expectation must match its single input", directory, test.Name)
			}
		}
		examples = append(examples, example{directory, suite})
	}
	if len(examples) == 0 {
		t.Fatal("no translated examples found")
	}
	return examples
}

// Unordered results are multisets: duplicates and group boundaries still matter.
func resultKey(t *testing.T, raw json.RawMessage, check string) string {
	t.Helper()
	var value any
	switch check {
	case "unordered", "unordered-groups":
		var items []json.RawMessage
		if err := json.Unmarshal(raw, &items); err != nil || items == nil {
			t.Fatalf("%s requires an array, got %s", check, raw)
		}
		itemCheck := "exact"
		if check == "unordered-groups" {
			itemCheck = "unordered"
		}
		keys := make([]string, len(items))
		for i, item := range items {
			keys[i] = resultKey(t, item, itemCheck)
		}
		slices.Sort(keys)
		value = keys
	case "", "exact", "codec-round-trip":
		decoder := json.NewDecoder(bytes.NewReader(raw))
		decoder.UseNumber()
		if err := decoder.Decode(&value); err != nil || !json.Valid(raw) {
			t.Fatalf("invalid JSON result %s: %v", raw, err)
		}
		value = normalizeNumbers(t, value)
	default:
		t.Fatalf("unknown check %q", check)
	}
	encoded, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return string(encoded)
}

func normalizeNumbers(t *testing.T, value any) any {
	t.Helper()
	switch value := value.(type) {
	case json.Number:
		number, ok := new(big.Int).SetString(string(value), 10)
		if !ok {
			t.Fatalf("expected an integer, got %s", value)
		}
		return json.Number(number.String())
	case []any:
		for i, item := range value {
			value[i] = normalizeNumbers(t, item)
		}
	case map[string]any:
		for key, item := range value {
			value[key] = normalizeNumbers(t, item)
		}
	}
	return value
}

func runPython(t *testing.T, python, directory string, suite fixture, args []json.RawMessage) json.RawMessage {
	t.Helper()
	request := struct {
		Entry         string            `json:"entry"`
		Args          []json.RawMessage `json:"args"`
		ArgAdapters   []string          `json:"argAdapters,omitempty"`
		ResultAdapter string            `json:"resultAdapter,omitempty"`
	}{suite.Entry, args, suite.ArgAdapters, suite.ResultAdapter}
	input, err := json.Marshal(request)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, python, filepath.Join("testdata", "python_runner.py"), filepath.Join(directory, "solution.py"))
	cmd.Stdin = bytes.NewReader(input)
	cmd.Env = append(os.Environ(), "PYTHONDONTWRITEBYTECODE=1", "PYTHONUTF8=1")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	output, err := cmd.Output()
	if err != nil {
		if ctx.Err() != nil {
			t.Fatal(ctx.Err())
		}
		t.Fatalf("%s: %v\n%s", suite.Entry, err, stderr.String())
	}
	return json.RawMessage(output)
}
