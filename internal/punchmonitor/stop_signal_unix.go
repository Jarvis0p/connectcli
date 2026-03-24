//go:build darwin || dragonfly || freebsd || linux || netbsd || openbsd || solaris || aix

package punchmonitor

import (
	"os"
	"syscall"
)

func terminateProcess(proc *os.Process) error {
	return proc.Signal(syscall.SIGTERM)
}
