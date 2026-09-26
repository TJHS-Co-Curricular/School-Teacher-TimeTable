// Package logx 同时把讯息印在黑色窗口、并写进 logs/YYYY-MM-DD.log。
//
// 窗口只显示讯息本身；日志文件每行加上时间与等级，例如：
//
//	2026-09-26 08:15:02 [INFO ] 已读取 班级课表.html（2.3 MB）← 192.168.0.31
package logx

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// Level 是日志等级。
type Level int

const (
	Debug Level = iota
	Info
	Warn
	Error
)

var levelNames = [...]string{"DEBUG", "INFO ", "WARN ", "ERROR"}

// ParseLevel 把设置文件里的文字转成等级，不认识时用 Info。
func ParseLevel(s string) Level {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "debug":
		return Debug
	case "warn", "warning":
		return Warn
	case "error":
		return Error
	}
	return Info
}

// Logger 写日志；可同时在多个 goroutine 使用。
type Logger struct {
	mu      sync.Mutex
	dir     string
	level   Level
	console io.Writer
	file    *os.File
	day     string
}

// New 建立 Logger；dir 为空字串时只印在窗口。
func New(dir string, level Level) *Logger {
	return &Logger{dir: dir, level: level, console: os.Stdout}
}

// SetLevel 修改等级（读完设置文件后调用）。
func (l *Logger) SetLevel(level Level) {
	l.mu.Lock()
	l.level = level
	l.mu.Unlock()
}

// Debugf 只写进日志文件（等级为 debug 时）。
func (l *Logger) Debugf(format string, a ...any) { l.write(Debug, false, format, a...) }

// Infof 印在窗口并写进日志。
func (l *Logger) Infof(format string, a ...any) { l.write(Info, true, format, a...) }

// Warnf 印在窗口（前面加 ⚠）并写进日志。
func (l *Logger) Warnf(format string, a ...any) { l.write(Warn, true, format, a...) }

// Errorf 印在窗口（前面加 ✖）并写进日志。
func (l *Logger) Errorf(format string, a ...any) { l.write(Error, true, format, a...) }

// Filef 只写进日志文件，不印在窗口。
func (l *Logger) Filef(format string, a ...any) { l.write(Info, false, format, a...) }

// Say 只印在窗口（排版用的说明文字，不写日志）。
func (l *Logger) Say(format string, a ...any) {
	l.mu.Lock()
	defer l.mu.Unlock()
	fmt.Fprintf(l.console, format+"\n", a...)
}

func (l *Logger) write(lv Level, toConsole bool, format string, a ...any) {
	l.mu.Lock()
	defer l.mu.Unlock()
	msg := fmt.Sprintf(format, a...)
	if toConsole && lv >= l.level {
		prefix := ""
		switch lv {
		case Warn:
			prefix = "⚠ "
		case Error:
			prefix = "✖ "
		}
		fmt.Fprintln(l.console, prefix+msg)
	}
	if lv < l.level || l.dir == "" {
		return
	}
	now := time.Now()
	if f := l.fileFor(now); f != nil {
		for _, line := range strings.Split(msg, "\n") {
			fmt.Fprintf(f, "%s [%s] %s\r\n", now.Format("2006-01-02 15:04:05"), levelNames[lv], line)
		}
	}
}

// 每天换一个文件
func (l *Logger) fileFor(now time.Time) *os.File {
	day := now.Format("2006-01-02")
	if l.file != nil && l.day == day {
		return l.file
	}
	if l.file != nil {
		l.file.Close()
		l.file = nil
	}
	f, err := os.OpenFile(filepath.Join(l.dir, day+".log"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		fmt.Fprintln(l.console, "⚠ 无法写入日志：", err)
		l.dir = "" // 不再尝试
		return nil
	}
	l.file, l.day = f, day
	return f
}

// Cleanup 删除超过 keepDays 天的 *.log；回传删除的数量。
func (l *Logger) Cleanup(keepDays int) int {
	if l.dir == "" || keepDays <= 0 {
		return 0
	}
	cutoff := time.Now().AddDate(0, 0, -keepDays)
	entries, _ := os.ReadDir(l.dir)
	n := 0
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".log") {
			continue
		}
		d, err := time.ParseInLocation("2006-01-02", strings.TrimSuffix(name, ".log"), time.Local)
		if err == nil && d.Before(cutoff) && os.Remove(filepath.Join(l.dir, name)) == nil {
			n++
		}
	}
	return n
}

// Close 关闭日志文件。
func (l *Logger) Close() {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.file != nil {
		l.file.Close()
		l.file = nil
	}
}
