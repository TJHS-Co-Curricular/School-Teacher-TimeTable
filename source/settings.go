package main

import (
	"bytes"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"unicode/utf16"
)

// 设置文件固定叫这个名字，放在 exe 旁边
const iniName = "Teacher-TimeTable.ini"

// 设置（来自 Teacher-TimeTable.ini）
type settings struct {
	lan  bool
	port int

	path string // 设置文件的完整路径
	note string // 读取时的提示（新建立 / 有错误的行 …）
}

// key 用来比对「正在运行的程序」和「这次读到的设置」是否相同
func (s settings) key() string {
	return fmt.Sprintf("lan=%v;port=%d", s.lan, s.port)
}

const defaultINI = `; 循人课表 设置
; 修改后，重新打开 Teacher-TimeTable.exe 就会生效（会自动关闭旧的窗口）。
;
; lan = true   同一网络（例如学校 Wi-Fi）的同事可以用 http://你的IP:端口/ 打开课表。
;              程序会一直运行，关闭黑色窗口才结束。
;              第一次运行时 Windows 防火墙会询问，请按「允许访问」。
; lan = false  只有这台电脑能打开；浏览器页面全部关闭约 3 分钟后程序自动结束。
lan = true

; 网页服务的端口（1024–65535）
port = 17380
`

func loadSettings(dir string) settings {
	st := settings{lan: true, port: defaultPort, path: filepath.Join(dir, iniName)}

	b, err := os.ReadFile(st.path)
	if os.IsNotExist(err) {
		if werr := os.WriteFile(st.path, []byte(strings.ReplaceAll(defaultINI, "\n", "\r\n")), 0o644); werr != nil {
			st.note = "（无法建立设置文件：" + werr.Error() + "，使用默认设置）"
		} else {
			st.note = "（第一次运行，已建立默认设置文件）"
		}
		return st
	}
	if err != nil {
		st.note = "（无法读取设置文件：" + err.Error() + "，使用默认设置）"
		return st
	}

	var bad []string
	for i, raw := range strings.Split(decodeText(b), "\n") {
		line := stripComment(raw)
		if line == "" || strings.HasPrefix(line, "[") { // 空行、[section] 都略过
			continue
		}
		k, v, ok := strings.Cut(line, "=")
		if !ok {
			k, v, ok = strings.Cut(line, ":")
		}
		k = strings.ToLower(strings.TrimSpace(k))
		v = strings.ToLower(strings.Trim(strings.TrimSpace(v), `"'`))
		if !ok {
			bad = append(bad, fmt.Sprintf("第 %d 行", i+1))
			continue
		}
		switch k {
		case "lan":
			switch v {
			case "true", "1", "yes", "on", "是":
				st.lan = true
			case "false", "0", "no", "off", "否":
				st.lan = false
			default:
				bad = append(bad, fmt.Sprintf("第 %d 行 lan 只能是 true 或 false", i+1))
			}
		case "port":
			n, err := strconv.Atoi(v)
			if err != nil || n < 1024 || n > 65535 {
				bad = append(bad, fmt.Sprintf("第 %d 行 port 要是 1024–65535 的数字", i+1))
				continue
			}
			st.port = n
		default:
			bad = append(bad, fmt.Sprintf("第 %d 行 不认识的设置「%s」", i+1, k))
		}
	}
	if len(bad) > 0 {
		st.note = "⚠ 设置文件有问题，已略过：" + strings.Join(bad, "；")
	}
	return st
}

// 去掉 ; 或 # 之后的注解
func stripComment(s string) string {
	if i := strings.IndexAny(s, ";#"); i >= 0 {
		s = s[:i]
	}
	return strings.TrimSpace(s)
}

// 记事本可能存成 UTF-8（含 BOM）或 UTF-16（「Unicode」），都转成字串
func decodeText(b []byte) string {
	switch {
	case bytes.HasPrefix(b, []byte{0xEF, 0xBB, 0xBF}):
		b = b[3:]
	case len(b) >= 2 && (b[0] == 0xFF && b[1] == 0xFE || b[0] == 0xFE && b[1] == 0xFF):
		le := b[0] == 0xFF
		b = b[2:]
		u := make([]uint16, len(b)/2)
		for i := range u {
			if le {
				u[i] = uint16(b[2*i]) | uint16(b[2*i+1])<<8
			} else {
				u[i] = uint16(b[2*i])<<8 | uint16(b[2*i+1])
			}
		}
		return strings.ReplaceAll(string(utf16.Decode(u)), "\r", "")
	}
	return strings.ReplaceAll(string(b), "\r", "")
}

type lanIP struct{ addr, iface string }

// 本机在局域网的 IPv4 地址（给同事用）
func lanIPs() []lanIP {
	var ips []lanIP
	ifaces, _ := net.Interfaces()
	for _, ifc := range ifaces {
		if ifc.Flags&net.FlagUp == 0 || ifc.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, _ := ifc.Addrs()
		for _, a := range addrs {
			ipn, ok := a.(*net.IPNet)
			if !ok {
				continue
			}
			ip := ipn.IP.To4()
			if ip == nil || ip.IsLinkLocalUnicast() {
				continue
			}
			ips = append(ips, lanIP{ip.String(), ifc.Name})
		}
	}
	return ips
}
