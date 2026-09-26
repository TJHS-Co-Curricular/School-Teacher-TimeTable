// Package paths 定义运行时文件夹（Teacher-TimeTable.exe 所在的文件夹）的结构：
//
//	Teacher-TimeTable/
//	├── Teacher-TimeTable.exe
//	├── config/   settings.ini、app.json（第一次运行自动建立）
//	├── data/     班级课表.html、场地课表_English.html
//	└── logs/     每天一个日志文件
//
// 开发时可以用环境变量 TEACHER_TIMETABLE_HOME 指定运行时文件夹。
package paths

import (
	"errors"
	"os"
	"path/filepath"
)

const envHome = "TEACHER_TIMETABLE_HOME"

// Layout 是运行时文件夹里各子文件夹的完整路径。
type Layout struct {
	Root   string
	Config string
	Data   string
	Logs   string
}

// Detect 找出运行时文件夹：环境变量优先，否则是 exe 所在的文件夹。
func Detect() Layout {
	root := os.Getenv(envHome)
	if root == "" {
		root = exeDir()
	}
	root, _ = filepath.Abs(root)
	return Layout{
		Root:   root,
		Config: filepath.Join(root, "config"),
		Data:   filepath.Join(root, "data"),
		Logs:   filepath.Join(root, "logs"),
	}
}

// SettingsFile 是程序设置文件。
func (l Layout) SettingsFile() string { return filepath.Join(l.Config, "settings.ini") }

// AppFile 是网页显示设置文件。
func (l Layout) AppFile() string { return filepath.Join(l.Config, "app.json") }

const dataReadme = "把从 eSchool「另存为」的课表文件放在这个文件夹：\r\n\r\n" +
	"  班级课表.html            （必需，教务处系统的班级课表）\r\n" +
	"  场地课表_English.html    （可选，英文系统的场地课表）\r\n\r\n" +
	"同名的 _files 文件夹可以一起放，也可以不放。\r\n" +
	"替换文件后，在浏览器按 F5 就会看到新课表。\r\n"

// Ensure 建立 config/、data/、logs/，并在 data/ 放一份说明。
func (l Layout) Ensure() error {
	for _, d := range []string{l.Config, l.Data, l.Logs} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			return err
		}
	}
	readme := filepath.Join(l.Data, "请把课表放在这里.txt")
	if _, err := os.Stat(readme); errors.Is(err, os.ErrNotExist) {
		_ = os.WriteFile(readme, []byte(dataReadme), 0o644)
	}
	return nil
}

// LegacySettings 回传 v1.2 放在 exe 旁边的 Teacher-TimeTable.ini；
// 只有在它存在、且 config/settings.ini 还没建立时才回传路径，否则为空字串。
func (l Layout) LegacySettings() string {
	old := filepath.Join(l.Root, "Teacher-TimeTable.ini")
	if _, err := os.Stat(old); err != nil {
		return ""
	}
	if _, err := os.Stat(l.SettingsFile()); err == nil {
		return ""
	}
	return old
}

func exeDir() string {
	exe, err := os.Executable()
	if err != nil {
		d, _ := os.Getwd()
		return d
	}
	if r, err := filepath.EvalSymlinks(exe); err == nil {
		exe = r
	}
	return filepath.Dir(exe)
}
