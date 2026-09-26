// Package platform 放与操作系统相关的功能：命令窗口、打开浏览器。
package platform

import (
	"os/exec"
	"runtime"
)

// OpenBrowser 用默认浏览器打开网址。
func OpenBrowser(url string) error {
	switch runtime.GOOS {
	case "windows":
		return exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		return exec.Command("open", url).Start()
	default:
		return exec.Command("xdg-open", url).Start()
	}
}
