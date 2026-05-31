//go:build windows

package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"time"
)

func platformStart(pm *ProcessManager, command string) error {
	pm.cmd = exec.Command("cmd", "/c", command)
	pm.cmd.Stdout = os.Stdout
	pm.cmd.Stderr = os.Stderr
	pm.cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP,
	}

	if err := pm.cmd.Start(); err != nil {
		pm.cmd = nil
		pm.pid = 0
		return fmt.Errorf("start core: %w", err)
	}
	pm.pid = pm.cmd.Process.Pid
	return nil
}

func platformStop(pm *ProcessManager) {
	pid := pm.pid
	if pid <= 0 {
		return
	}

	// Try graceful terminate first
	if pm.cmd != nil && pm.cmd.Process != nil {
		pm.cmd.Process.Kill()
	}

	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if !processAlive(pid) {
			break
		}
		time.Sleep(200 * time.Millisecond)
	}

	// Force kill if still alive
	if processAlive(pid) && pm.cmd != nil && pm.cmd.Process != nil {
		pm.cmd.Process.Kill()
		time.Sleep(200 * time.Millisecond)
	}

	pm.cmd = nil
	pm.pid = 0
}

func platformShellExec(ctx context.Context, command string) *exec.Cmd {
	return exec.CommandContext(ctx, "cmd", "/c", command)
}

func platformProcessAlive(pid int) bool {
	_, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	// On Windows, FindProcess always succeeds.
	// Check if process is still running by getting exit code.
	handle, err := syscall.OpenProcess(syscall.PROCESS_QUERY_INFORMATION, false, uint32(pid))
	if err != nil {
		return false
	}
	defer syscall.CloseHandle(handle)
	var exitCode uint32
	err = syscall.GetExitCodeProcess(handle, &exitCode)
	if err != nil {
		return false
	}
	return exitCode == 259 // STILL_ACTIVE
}
