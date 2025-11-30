//go:build windows
// +build windows

package main

import (
	"fmt"

	"golang.org/x/sys/windows"
)

var appMutex windows.Handle

// ensureSingleInstance 尝试创建一个命名 Mutex。
// 如果已经有实例在运行，则返回 false。
func ensureSingleInstance() bool {
	name, err := windows.UTF16PtrFromString("CUMT_AUTLOGIN_MUTEX_V1")
	if err != nil {
		// 转换失败就不做单实例限制
		return true
	}

	h, err := windows.CreateMutex(nil, false, name)
	if err != nil {
		// 如果已经存在，说明有别的实例在跑
		if err == windows.ERROR_ALREADY_EXISTS {
			fmt.Println("[INFO] another instance is already running")
			// 这里可以 CloseHandle，也可以直接返回
			_ = windows.CloseHandle(h)
			return false
		}
		// 其他错误就当没限制
		fmt.Println("[WARN] CreateMutex error:", err)
		return true
	}

	appMutex = h
	return true
}

func releaseSingleInstance() {
	if appMutex != 0 {
		_ = windows.ReleaseMutex(appMutex)
		_ = windows.CloseHandle(appMutex)
		appMutex = 0
	}
}
