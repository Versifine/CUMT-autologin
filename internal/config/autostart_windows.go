//go:build windows
// +build windows

package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"syscall"

	"golang.org/x/sys/windows/registry"
)

const runKeyPath = `Software\Microsoft\Windows\CurrentVersion\Run`
const runValueName = "CUMTAutologin"

func exePath() (string, error) {
	p, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.Abs(p)
}

func SetAutoStart(enabled bool) error {
	exe, err := exePath()
	if err != nil {
		return err
	}

	if enabled {
		k, _, err := registry.CreateKey(registry.CURRENT_USER, runKeyPath, registry.SET_VALUE)
		if err != nil {
			return err
		}
		defer k.Close()
		return k.SetStringValue(runValueName, fmt.Sprintf(`"%s"`, exe))
	}

	k, err := registry.OpenKey(registry.CURRENT_USER, runKeyPath, registry.SET_VALUE)
	if err != nil {
		if errors.Is(err, syscall.ERROR_FILE_NOT_FOUND) {
			return nil
		}
		return err
	}
	defer k.Close()
	err = k.DeleteValue(runValueName)
	if err != nil && !errors.Is(err, syscall.ERROR_FILE_NOT_FOUND) {
		return err
	}
	return nil
}

func IsAutoStartEnabled() bool {
	k, err := registry.OpenKey(registry.CURRENT_USER, runKeyPath, registry.QUERY_VALUE)
	if err != nil {
		return false
	}
	defer k.Close()
	_, _, err = k.GetStringValue(runValueName)
	return err == nil
}
