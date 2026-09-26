// Package source 找出 eSchool 导出的课表文件。
//
// 找的顺序：运行时文件夹的 data/ → exe 旁边（v1.2 以前的放法）。
// 每个位置先找固定文件名，找不到再找内容符合的最新 .html。
package source

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Kind 是课表来源的种类。
type Kind struct {
	Label  string   // 显示用
	Route  string   // 网页读取的网址
	Names  []string // 优先使用的文件名
	Marker string   // 找不到固定文件名时，文件内容里要有的字（eSchool 列印页的网址）
}

var (
	// Class 是教务处系统的「班级课表」（必需）。
	Class = Kind{
		Label:  "班级课表.html",
		Route:  "/source.html",
		Names:  []string{"班级课表.html", "班级课表.htm", "班級課表.html"},
		Marker: "csche_prt.php",
	}
	// English 是英文系统的「场地课表_English」（可选）。
	English = Kind{
		Label:  "场地课表_English.html",
		Route:  "/source_en.html",
		Names:  []string{"场地课表_English.html", "场地课表_english.html", "場地課表_English.html"},
		Marker: "schedule_eng/psche_prt.php",
	}
)

// Result 是找到的文件。
type Result struct {
	Path   string
	Legacy bool // 在 exe 旁边（旧放法），建议搬到 data/
}

// Find 依序在 dataDir、legacyDir 找课表；都找不到时 Path 为空字串。
func Find(dataDir, legacyDir string, k Kind) Result {
	if p := findIn(dataDir, k); p != "" {
		return Result{Path: p}
	}
	if legacyDir != "" && legacyDir != dataDir {
		if p := findIn(legacyDir, k); p != "" {
			return Result{Path: p, Legacy: true}
		}
	}
	return Result{}
}

func findIn(dir string, k Kind) string {
	for _, name := range k.Names {
		p := filepath.Join(dir, name)
		if st, err := os.Stat(p); err == nil && !st.IsDir() {
			return p
		}
	}
	entries, _ := os.ReadDir(dir)
	type cand struct {
		p string
		t time.Time
	}
	var cs []cand
	for _, e := range entries {
		n := strings.ToLower(e.Name())
		if e.IsDir() || !(strings.HasSuffix(n, ".html") || strings.HasSuffix(n, ".htm")) {
			continue
		}
		p := filepath.Join(dir, e.Name())
		if !fileContains(p, k.Marker) {
			continue
		}
		info, _ := e.Info()
		cs = append(cs, cand{p, info.ModTime()})
	}
	sort.Slice(cs, func(i, j int) bool { return cs[i].t.After(cs[j].t) })
	if len(cs) > 0 {
		return cs[0].p
	}
	return ""
}

// 网址写在文件开头的「saved from url」注解里，只读前 4 KB 就够
func fileContains(path, marker string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()
	buf := make([]byte, 4096)
	n, _ := f.Read(buf)
	return strings.Contains(string(buf[:n]), marker)
}
