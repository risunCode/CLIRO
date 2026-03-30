//go:build darwin || linux

package platform

import (
	"syscall"
)

// getOSVersion returns Unix-like OS version from uname
func getOSVersion() string {
	var utsname syscall.Utsname
	if err := syscall.Uname(&utsname); err != nil {
		return ""
	}

	// Convert [65]int8 to string
	release := make([]byte, 0, 65)
	for _, b := range utsname.Release {
		if b == 0 {
			break
		}
		release = append(release, byte(b))
	}

	return string(release)
}
