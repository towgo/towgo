//go:build !windows

package processmanager

import "syscall"

func isProcessRunning(pid int) bool {
	if pid <= 0 {
		return false
	}
	return syscall.Kill(pid, 0) == nil
}

func killProcess(pid int) error {
	return syscall.Kill(pid, syscall.SIGKILL)
}
