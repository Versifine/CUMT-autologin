//go:build !windows

package main

func currentSSID() (string, error) {
	return "", nil
}
