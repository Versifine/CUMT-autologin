//go:build windows
// +build windows

package main

import (
	_ "embed"
	"encoding/base64"
)

//go:embed assets/icon.ico
var iconData []byte

func IconBase64() string {
	return "data:image/x-icon;base64," + base64.StdEncoding.EncodeToString(iconData)
}
