//go:build darwin || linux

package repl

import (
	"os"
	"syscall"
	"unsafe"
)

func isTerminal(file *os.File) bool {
	var termios syscall.Termios
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, file.Fd(), terminalRequest, uintptr(unsafe.Pointer(&termios)))
	return errno == 0
}
