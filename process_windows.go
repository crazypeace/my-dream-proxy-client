//go:build windows

package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"
)

func platformStart(pm *ProcessManager, command string) error {
	// Split command into exe + args, launch directly (no cmd /c).
	// This makes cmd.Process.Kill() kill the actual core process.
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return fmt.Errorf("empty core-start command")
	}
	exe := convertPaths(parts[0])
	args := make([]string, len(parts)-1)
	for i, a := range parts[1:] {
		args[i] = convertPaths(a)
	}

	pm.cmd = exec.Command(exe, args...)
	pm.cmd.Stdout = os.Stdout
	pm.cmd.Stderr = os.Stderr

	if err := pm.cmd.Start(); err != nil {
		pm.cmd = nil
		pm.pid = 0
		return fmt.Errorf("start core: %w", err)
	}

	pm.pid = pm.cmd.Process.Pid
	return nil
}

func platformStop(pm *ProcessManager) {
	if pm.cmd != nil && pm.cmd.Process != nil {
		pm.cmd.Process.Kill()
	}

	// Wait for process to actually die
	pid := pm.pid
	for i := 0; i < 30; i++ {
		if !processAlive(pid) {
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func platformShellExec(ctx context.Context, command string) *exec.Cmd {
	command = convertPaths(command)
	return exec.CommandContext(ctx, "cmd", "/c", command)
}

func platformProcessAlive(pid int) bool {
	handle, err := syscall.OpenProcess(0x0400 /* PROCESS_QUERY_INFORMATION */, false, uint32(pid))
	if err != nil {
		return false
	}
	defer syscall.CloseHandle(handle)
	var exitCode uint32
	if syscall.GetExitCodeProcess(handle, &exitCode) != nil {
		return false
	}
	return exitCode == 259 // STILL_ACTIVE
}

func convertPaths(s string) string {
	return strings.ReplaceAll(s, "/", "\\")
}
