package punchmonitor

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const pidFileName = "punch_monitor.pid"

// Dir returns ~/.connectcli (creates directory if missing).
func Dir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	d := filepath.Join(home, ".connectcli")
	if err := os.MkdirAll(d, 0755); err != nil {
		return "", err
	}
	return d, nil
}

func pidPath() (string, error) {
	d, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(d, pidFileName), nil
}

// WritePID writes the monitor process ID for punch-out / stale detection.
func WritePID(pid int) error {
	p, err := pidPath()
	if err != nil {
		return err
	}
	return os.WriteFile(p, []byte(fmt.Sprintf("%d\n", pid)), 0600)
}

// ReadPID returns the PID from the pid file, or an error if missing/invalid.
func ReadPID() (int, error) {
	p, err := pidPath()
	if err != nil {
		return 0, err
	}
	data, err := os.ReadFile(p)
	if err != nil {
		return 0, err
	}
	s := strings.TrimSpace(string(data))
	if s == "" {
		return 0, fmt.Errorf("empty pid file")
	}
	return strconv.Atoi(s)
}

// RemovePIDFile deletes the pid file unconditionally.
func RemovePIDFile() error {
	p, err := pidPath()
	if err != nil {
		return err
	}
	if err := os.Remove(p); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

// RemovePIDFileIfMatches removes the pid file only if it contains the given PID.
func RemovePIDFileIfMatches(expected int) error {
	pid, err := ReadPID()
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		_ = RemovePIDFile()
		return nil
	}
	if pid != expected {
		return nil
	}
	return RemovePIDFile()
}

// Stop sends SIGTERM to the stored monitor process and removes the pid file.
func Stop() error {
	pid, err := ReadPID()
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		_ = RemovePIDFile()
		return nil
	}

	proc, err := os.FindProcess(pid)
	if err != nil {
		_ = RemovePIDFile()
		return nil
	}

	_ = terminateProcess(proc)
	_ = RemovePIDFile()
	return nil
}
