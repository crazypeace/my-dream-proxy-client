//go:build !windows

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
	if command == "" {
		return fmt.Errorf("empty core-start command")
	}
	pm.cmd = exec.Command("sh", "-c", command)
	pm.cmd.Stdout = os.Stdout
	pm.cmd.Stderr = os.Stderr
	pm.cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true, // Create new process group for clean kill
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

	// SIGTERM the process group, wait, then SIGKILL
	syscall.Kill(-pid, syscall.SIGTERM)

	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if !processAlive(pid) {
			return
		}
		time.Sleep(200 * time.Millisecond)
	}

	if processAlive(pid) {
		syscall.Kill(-pid, syscall.SIGKILL)
		time.Sleep(200 * time.Millisecond)
	}
}

func platformShellExec(ctx context.Context, command string) *exec.Cmd {
	return exec.CommandContext(ctx, "sh", "-c", command)
}

func platformProcessAlive(pid int) bool {
	p, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	return p.Signal(syscall.Signal(0)) == nil
}
