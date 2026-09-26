package platform

import (
	"syscall"
	"unsafe"
)

// SetupConsole 让 Windows 命令窗口正确显示中文，并设定窗口标题。
func SetupConsole(title string) {
	k := syscall.NewLazyDLL("kernel32.dll")
	k.NewProc("SetConsoleOutputCP").Call(65001)
	k.NewProc("SetConsoleCP").Call(65001)
	if p, err := syscall.UTF16PtrFromString(title); err == nil {
		k.NewProc("SetConsoleTitleW").Call(uintptr(unsafe.Pointer(p)))
	}
}
