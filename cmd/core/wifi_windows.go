//go:build windows

package main

import (
	"bytes"
	"os/exec"
	"regexp"
	"strings"
	"syscall"
)

func currentSSID() (string, error) {
	cmd := exec.Command("netsh", "wlan", "show", "interfaces")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow: true,
	}
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}

	normalized := bytes.ReplaceAll(out, []byte("\r\n"), []byte("\n"))
	re := regexp.MustCompile(`(?m)^\s*SSID\s*:\s*(.+)$`)
	matches := re.FindSubmatch(normalized)
	if len(matches) >= 2 {
		return strings.TrimSpace(string(matches[1])), nil
	}
	return "", nil
}
