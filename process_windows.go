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
	"unsafe"
)

var (
	modkernel32                 = syscall.NewLazyDLL("kernel32.dll")
	procCreateJobObjectW        = modkernel32.NewProc("CreateJobObjectW")
	procAssignProcessToJob      = modkernel32.NewProc("AssignProcessToJobObject")
	procTerminateJobObject      = modkernel32.NewProc("TerminateJobObject")
	procSetInformationJobObject = modkernel32.NewProc("SetInformationJobObject")
)

const (
	jobObjectLimitKillOnJobClose              = 0x2000
	jobObjectExtendedLimitInformationClass    = 9
	processSetQuota                           = 0x0100
	processTerminate                          = 0x0001
)

type ioCounters struct {
	ReadOperationCount  uint64
	WriteOperationCount uint64
	OtherOperationCount uint64
	ReadTransferCount   uint64
	WriteTransferCount  uint64
	OtherTransferCount  uint64
}

type jobObjectExtendedLimitInformation struct {
	BasicLimitInformation struct {
		PerProcessUserTimeLimit uint64
		PerJobUserTimeLimit     uint64
		LimitFlags              uint32
		MinimumWorkingSetSize   uintptr
		MaximumWorkingSetSize   uintptr
		ActiveProcessLimit      uint32
		Affinity                uintptr
		PriorityClass           uint32
		SchedulingClass         uint32
	}
	_                    [4]byte
	IoInfo               ioCounters
	ProcessMemoryLimit   uintptr
	JobMemoryLimit       uintptr
	PeakProcessMemoryUsed uintptr
	PeakJobMemoryUsed    uintptr
}

// createJobObject creates a Windows Job Object with kill-on-close semantics.
func createJobObject() (uintptr, error) {
	handle, _, err := procCreateJobObjectW.Call(0, 0)
	if handle == 0 {
		return 0, fmt.Errorf("CreateJobObject: %v", err)
	}

	info := jobObjectExtendedLimitInformation{}
	info.BasicLimitInformation.LimitFlags = jobObjectLimitKillOnJobClose
	infoSize := unsafe.Sizeof(info)

	ret, _, err := procSetInformationJobObject.Call(
		handle,
		uintptr(jobObjectExtendedLimitInformationClass),
		uintptr(unsafe.Pointer(&info)),
		uintptr(infoSize),
	)
	if ret == 0 {
		syscall.CloseHandle(syscall.Handle(handle))
		return 0, fmt.Errorf("SetInformationJobObject: %v", err)
	}

	return handle, nil
}

func assignProcessToJob(job uintptr, pid int) error {
	proc, err := syscall.OpenProcess(processSetQuota|processTerminate, false, uint32(pid))
	if err != nil {
		return err
	}
	defer syscall.CloseHandle(proc)

	ret, _, err := procAssignProcessToJob.Call(job, uintptr(proc))
	if ret == 0 {
		return fmt.Errorf("AssignProcessToJobObject: %v", err)
	}
	return nil
}

func terminateJob(job uintptr) {
	procTerminateJobObject.Call(job, 1)
}

func platformStart(pm *ProcessManager, command string) error {
	command = convertPaths(command)

	job, err := createJobObject()
	if err != nil {
		return fmt.Errorf("create job object: %w", err)
	}

	pm.cmd = exec.Command("cmd", "/c", command)
	pm.cmd.Stdout = os.Stdout
	pm.cmd.Stderr = os.Stderr

	if err := pm.cmd.Start(); err != nil {
		syscall.CloseHandle(syscall.Handle(job))
		pm.cmd = nil
		pm.pid = 0
		return fmt.Errorf("start core: %w", err)
	}

	pm.pid = pm.cmd.Process.Pid

	// Assign cmd.exe to the job. All child processes (xray etc.)
	// inherit job membership and will be killed together.
	if err := assignProcessToJob(job, pm.pid); err != nil {
		pm.cmd.Process.Kill()
		syscall.CloseHandle(syscall.Handle(job))
		pm.cmd = nil
		pm.pid = 0
		return fmt.Errorf("assign to job: %w", err)
	}

	pm.jobHandle = job
	return nil
}

func platformStop(pm *ProcessManager) {
	pid := pm.pid
	if pid <= 0 && pm.jobHandle == 0 {
		return
	}

	if pm.jobHandle != 0 {
		terminateJob(pm.jobHandle)
		deadline := time.Now().Add(3 * time.Second)
		for time.Now().Before(deadline) {
			if pid <= 0 || !processAlive(pid) {
				break
			}
			time.Sleep(200 * time.Millisecond)
		}
		syscall.CloseHandle(syscall.Handle(pm.jobHandle))
		pm.jobHandle = 0
	} else if pm.cmd != nil && pm.cmd.Process != nil {
		pm.cmd.Process.Kill()
	}

	pm.cmd = nil
	pm.pid = 0
}

func platformShellExec(ctx context.Context, command string) *exec.Cmd {
	command = convertPaths(command)
	return exec.CommandContext(ctx, "cmd", "/c", command)
}

func platformProcessAlive(pid int) bool {
	_, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	handle, err := syscall.OpenProcess(0x0400 /* PROCESS_QUERY_INFORMATION */, false, uint32(pid))
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

func convertPaths(s string) string {
	return strings.ReplaceAll(s, "/", "\\")
}
