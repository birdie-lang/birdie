//go:build !darwin && !linux

package repl

import "os"

func isTerminal(_ *os.File) bool {
	return false
}
