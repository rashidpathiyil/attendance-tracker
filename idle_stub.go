//go:build !windows
// +build !windows

package main

import "time"

// getSystemIdleTimeWindows is a stub implementation for non-Windows platforms
// This function is only used when cross-compiling for Windows from non-Windows systems
func getSystemIdleTimeWindows() (time.Duration, error) {
	return time.Duration(0), nil
}
