package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"time"
)

type ProcessManager struct {
	mu   sync.Mutex
	cmd  *exec.Cmd
	pid  int
}

type Status struct {
	Running bool   `json:"running"`
	PID     int    `json:"pid"`
}

func NewProcessManager() *ProcessManager {
	return &ProcessManager{}
}

func (pm *ProcessManager) Status() Status {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if pm.pid > 0 && !processAlive(pm.pid) {
		pm.cmd = nil
		pm.pid = 0
		return Status{Running: false, PID: 0}
	}
	if pm.pid > 0 {
		return Status{Running: true, PID: pm.pid}
	}
	return Status{Running: false, PID: 0}
}

// Start starts the core process. If already running, acts as restart (stop then start).
func (pm *ProcessManager) Start(command string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	// If already running, stop it first (restart behavior)
	if pm.pid > 0 && processAlive(pm.pid) {
		pm.stopLocked()
		time.Sleep(500 * time.Millisecond)
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

	// Wait in background, clean up state when done
	go func() {
		pm.cmd.Wait()
		pm.mu.Lock()
		if pm.cmd != nil && pm.cmd.Process != nil && pm.cmd.Process.Pid == pm.pid {
			pm.cmd = nil
			pm.pid = 0
		}
		pm.mu.Unlock()
	}()

	return nil
}

func (pm *ProcessManager) Stop() error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if pm.pid <= 0 {
		return fmt.Errorf("core is not running")
	}

	pm.stopLocked()
	return nil
}

// stopLocked does the actual stop. Caller must hold pm.mu.
func (pm *ProcessManager) stopLocked() {
	pid := pm.pid
	if pid <= 0 {
		return
	}

	// SIGTERM the process group, wait, then SIGKILL
	syscall.Kill(-pid, syscall.SIGTERM)

	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if !processAlive(pid) {
			break
		}
		time.Sleep(200 * time.Millisecond)
	}

	if processAlive(pid) {
		syscall.Kill(-pid, syscall.SIGKILL)
		time.Sleep(200 * time.Millisecond)
	}

	pm.cmd = nil
	pm.pid = 0
}

func (pm *ProcessManager) TestConfig(command string) (valid bool, errMsg string) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "sh", "-c", command)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return false, string(output)
	}
	return true, ""
}

// Cleanup kills the core process if running. Called on app exit.
func (pm *ProcessManager) Cleanup() {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.stopLocked()
}

// processAlive checks if a process with the given PID exists.
func processAlive(pid int) bool {
	p, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	return p.Signal(syscall.Signal(0)) == nil
}
