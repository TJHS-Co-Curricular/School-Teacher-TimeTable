// Package buildinfo 保存程序名称与版本号。
//
// Version / Commit / BuildDate 由打包脚本在编译时注入：
//
//	go build -ldflags "-X github.com/TJHS-Co-Curricular/School-Teacher-TimeTable/internal/buildinfo.Version=1.3.0"
//
// 版本号放在项目根目录的 VERSION 文件，改版时只改那里（并更新 CHANGELOG.md）。
package buildinfo

const (
	AppName = "循人课表"
	ExeName = "Teacher-TimeTable"
	AppID   = "tjhs-timetable" // 用来辨认「正在运行的是不是本程序」
)

var (
	Version   = "dev"
	Commit    = ""
	BuildDate = ""
)

// String 回传显示用的版本字串，例如「v1.3.0 (a1b2c3d, 2026-09-26)」。
func String() string {
	s := "v" + Version
	if Version == "dev" {
		s = "开发版"
	}
	extra := ""
	if Commit != "" {
		extra = Commit
	}
	if BuildDate != "" {
		if extra != "" {
			extra += ", "
		}
		extra += BuildDate
	}
	if extra != "" {
		s += " (" + extra + ")"
	}
	return s
}
