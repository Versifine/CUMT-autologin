//go:build windows
// +build windows

package main

import (
	_ "embed"
)

//go:embed assets/icon.ico
var iconData []byte
