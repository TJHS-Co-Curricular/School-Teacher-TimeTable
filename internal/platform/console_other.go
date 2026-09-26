//go:build !windows

package platform

// SetupConsole 在非 Windows 系统不需要做什么。
func SetupConsole(title string) {}
