// Package settings 读取运行时文件夹里的 config/settings.ini 与 config/app.json。
//
// 文件不存在时，用打包在 exe 里的默认内容建立（见项目根目录的 config/）。
package settings

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"unicode/utf16"
)

// Settings 是 settings.ini 的内容。
type Settings struct {
	LAN         bool   // [server] lan
	Port        int    // [server] port
	OpenBrowser bool   // [server] open_browser
	LogKeepDays int    // [log] keep_days
	LogLevel    string // [log] level

	Path    string   // 设置文件完整路径
	Created bool     // 这次运行才建立的
	Issues  []string // 写错、被略过的行
}

// Key 用来比对「正在运行的程序」和「这次读到的设置」是否相同。
func (s Settings) Key() string {
	return fmt.Sprintf("lan=%v;port=%d", s.LAN, s.Port)
}

// Load 读取 path；不存在时用 defaults 建立。defaults 本身也会先被解析，作为默认值。
func Load(path string, defaults []byte) (Settings, error) {
	s := Settings{LAN: true, Port: 17380, OpenBrowser: true, LogKeepDays: 30, LogLevel: "info", Path: path}
	s.apply(defaults, false)

	b, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		if werr := os.WriteFile(path, toCRLF(defaults), 0o644); werr != nil {
			return s, fmt.Errorf("无法建立设置文件：%w", werr)
		}
		s.Created = true
		return s, nil
	}
	if err != nil {
		return s, fmt.Errorf("无法读取设置文件：%w", err)
	}
	s.apply(b, true)
	return s, nil
}

// 旧版（v1.2）的 ini 没有 [section]，这些键直接对应到新位置
var bareKeys = map[string]string{
	"lan":          "server.lan",
	"port":         "server.port",
	"open_browser": "server.open_browser",
	"keep_days":    "log.keep_days",
	"level":        "log.level",
}

func (s *Settings) apply(b []byte, report bool) {
	section := ""
	bad := func(line int, msg string) {
		if report {
			s.Issues = append(s.Issues, fmt.Sprintf("第 %d 行：%s", line, msg))
		}
	}
	for i, raw := range strings.Split(decodeText(b), "\n") {
		n := i + 1
		line := stripComment(raw)
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			section = strings.ToLower(strings.TrimSpace(line[1 : len(line)-1]))
			continue
		}
		k, v, ok := strings.Cut(line, "=")
		if !ok {
			bad(n, "缺少「=」")
			continue
		}
		k = strings.ToLower(strings.TrimSpace(k))
		v = strings.Trim(strings.TrimSpace(v), `"'`)
		key := k
		if section != "" {
			key = section + "." + k
		} else if full, ok := bareKeys[k]; ok {
			key = full
		}
		switch key {
		case "server.lan":
			if b, ok := parseBool(v); ok {
				s.LAN = b
			} else {
				bad(n, "lan 只能是 true 或 false")
			}
		case "server.open_browser":
			if b, ok := parseBool(v); ok {
				s.OpenBrowser = b
			} else {
				bad(n, "open_browser 只能是 true 或 false")
			}
		case "server.port":
			if p, err := strconv.Atoi(v); err == nil && p >= 1024 && p <= 65535 {
				s.Port = p
			} else {
				bad(n, "port 要是 1024–65535 的数字")
			}
		case "log.keep_days":
			if d, err := strconv.Atoi(v); err == nil && d >= 0 {
				s.LogKeepDays = d
			} else {
				bad(n, "keep_days 要是 0 以上的数字")
			}
		case "log.level":
			switch strings.ToLower(v) {
			case "debug", "info", "warn":
				s.LogLevel = strings.ToLower(v)
			default:
				bad(n, "level 只能是 debug、info 或 warn")
			}
		default:
			bad(n, "不认识的设置「"+key+"」")
		}
	}
}

// Render 把 s 的值填进 template（默认设置文件），保留 template 的注解与排版。
// 用于把旧版设置转成新格式。
func Render(template []byte, s Settings) []byte {
	vals := map[string]string{
		"server.lan":          strconv.FormatBool(s.LAN),
		"server.port":         strconv.Itoa(s.Port),
		"server.open_browser": strconv.FormatBool(s.OpenBrowser),
		"log.keep_days":       strconv.Itoa(s.LogKeepDays),
		"log.level":           s.LogLevel,
	}
	section := ""
	lines := strings.Split(decodeText(template), "\n")
	for i, raw := range lines {
		line := stripComment(raw)
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			section = strings.ToLower(line[1 : len(line)-1])
			continue
		}
		k, _, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		if v, ok := vals[section+"."+strings.ToLower(strings.TrimSpace(k))]; ok {
			lines[i] = strings.TrimSpace(k) + " = " + v
		}
	}
	return toCRLF([]byte(strings.Join(lines, "\n")))
}

func parseBool(v string) (bool, bool) {
	switch strings.ToLower(v) {
	case "true", "1", "yes", "on", "是":
		return true, true
	case "false", "0", "no", "off", "否":
		return false, true
	}
	return false, false
}

// LoadApp 读取 app.json（网页显示设置），回传合并了默认值的内容。
// 文件不存在时用 defaults 建立；格式错误时回传默认值与错误。
func LoadApp(path string, defaults []byte) (map[string]any, error) {
	out := map[string]any{}
	_ = json.Unmarshal(defaults, &out)

	b, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		if werr := os.WriteFile(path, toCRLF(defaults), 0o644); werr != nil {
			return out, fmt.Errorf("无法建立 app.json：%w", werr)
		}
		return out, nil
	}
	if err != nil {
		return out, fmt.Errorf("无法读取 app.json：%w", err)
	}
	user := map[string]any{}
	if err := json.Unmarshal([]byte(decodeText(b)), &user); err != nil {
		return out, fmt.Errorf("app.json 格式错误（%v），改用默认值", err)
	}
	for k, v := range user {
		out[k] = v
	}
	return out, nil
}

// 去掉 ; 或 # 之后的注解
func stripComment(s string) string {
	if i := strings.IndexAny(s, ";#"); i >= 0 {
		s = s[:i]
	}
	return strings.TrimSpace(s)
}

// 记事本可能存成 UTF-8（含 BOM）或 UTF-16（「Unicode」），都转成一般字串
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

// Windows 记事本友好的换行
func toCRLF(b []byte) []byte {
	s := strings.ReplaceAll(string(b), "\r\n", "\n")
	return []byte(strings.ReplaceAll(s, "\n", "\r\n"))
}
