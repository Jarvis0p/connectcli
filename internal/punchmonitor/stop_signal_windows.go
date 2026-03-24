//go:build windows

package punchmonitor

import "os"

func terminateProcess(proc *os.Process) error {
	return proc.Kill()
}
