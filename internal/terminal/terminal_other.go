//go:build !darwin && !linux

package terminal

import "os"

func IsTerminal(_ *os.File) bool {
	return false
}
