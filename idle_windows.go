//go:build windows
// +build windows

package main

import (
	"syscall"
	"time"
	"unsafe"
)

var (
	user32               = syscall.NewLazyDLL("user32.dll")
	kernel32             = syscall.NewLazyDLL("kernel32.dll")
	procGetLastInputInfo = user32.NewProc("GetLastInputInfo")
	procGetTickCount     = kernel32.NewProc("GetTickCount")
)

type LASTINPUTINFO struct {
	CbSize uint32
	DwTime uint32
}

// getSystemIdleTimeWindows returns the idle time for Windows
func getSystemIdleTimeWindows() (time.Duration, error) {
	lastInput := LASTINPUTINFO{
		CbSize: uint32(unsafe.Sizeof(LASTINPUTINFO{})),
	}

	ret, _, _ := procGetLastInputInfo.Call(uintptr(unsafe.Pointer(&lastInput)))
	if ret == 0 {
		return 0, nil
	}

	currentTicks, _, _ := procGetTickCount.Call()

	idleTime := uint32(currentTicks) - lastInput.DwTime

	return time.Duration(idleTime) * time.Millisecond, nil
}
