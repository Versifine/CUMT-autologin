//go:build windows
// +build windows

package main

import (
	"fmt"

	"golang.org/x/sys/windows"
)

var guiMutex windows.Handle

// ensureSingleInstance guards GUI to a single process.
func ensureSingleInstance() bool {
	name, err := windows.UTF16PtrFromString("CUMT_AUTLOGIN_GUI_MUTEX_V1")
	if err != nil {
		return true
	}
	h, err := windows.CreateMutex(nil, false, name)
	if err != nil {
		if err == windows.ERROR_ALREADY_EXISTS {
			fmt.Println("[GUI] another instance is already running")
			_ = windows.CloseHandle(h)
			return false
		}
		fmt.Println("[GUI] CreateMutex error:", err)
		return true
	}
	guiMutex = h
	return true
}

func releaseSingleInstance() {
	if guiMutex != 0 {
		_ = windows.ReleaseMutex(guiMutex)
		_ = windows.CloseHandle(guiMutex)
		guiMutex = 0
	}
}
