// Teacher-TimeTable.exe —— 循人中学 教师与班级课表
//
// 每次启动：在本程序所在的文件夹里找「班级课表.html」，开一个只供本机使用的
// 小型网页服务，并用默认浏览器打开课表页面。页面每次打开 / 重新整理都会重新
// 读取「班级课表.html」，所以只要替换那个文件，课表就会自动更新。
// 浏览器页面全部关闭约 3 分钟后，程序会自动结束。
package main

import (
	"embed"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync/atomic"
	"time"
)

// 网页文件（web/ 目录）在编译时打包进 exe
//
//go:embed web
var webFiles embed.FS

const (
	appID       = "tjhs-timetable"
	defaultPort = 17380
	idleTimeout = 3 * time.Minute
)

var lastPing atomic.Int64 // unix 秒；0 = 还没有页面连上

func main() {
	setConsoleUTF8()
	fmt.Println("==============================================")
	fmt.Println("  循人中学 · 教师与班级课表")
	fmt.Println("==============================================")

	dir := exeDir()
	if src := findSource(dir); src != "" {
		fmt.Println("课表来源：", src)
	} else {
		fmt.Println("⚠ 在这个文件夹里找不到「班级课表.html」：")
		fmt.Println("  ", dir)
		fmt.Println("  页面打开后可以手动选择文件。")
	}

	// 已经有一个在运行：直接打开它
	existing := fmt.Sprintf("http://127.0.0.1:%d/", defaultPort)
	if isOurs(existing) {
		fmt.Println("课表程序已经在运行，直接打开页面。")
		openBrowser(existing)
		time.Sleep(1500 * time.Millisecond)
		return
	}

	ln, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", defaultPort))
	if err != nil {
		ln, err = net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			fail("无法启动本机网页服务：" + err.Error())
		}
	}
	url := fmt.Sprintf("http://%s/", ln.Addr().String())

	webRoot, err := fs.Sub(webFiles, "web")
	if err != nil {
		fail("网页文件损坏：" + err.Error())
	}
	static := http.FileServer(http.FS(webRoot))

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		noCache(w)
		switch r.URL.Path {
		case "/source.html":
			src := findSource(dir) // 每次都重新找、重新读
			if src == "" {
				http.Error(w, "找不到 班级课表.html", http.StatusNotFound)
				return
			}
			b, err := os.ReadFile(src)
			if err != nil {
				http.Error(w, "读取失败："+err.Error(), http.StatusInternalServerError)
				return
			}
			fmt.Printf("[%s] 已读取 %s\n", time.Now().Format("15:04:05"), filepath.Base(src))
			w.Header().Set("Content-Type", "application/octet-stream")
			w.Write(b)
		case "/__ping":
			lastPing.Store(time.Now().Unix())
			w.Write([]byte(appID))
		default:
			static.ServeHTTP(w, r) // index.html、css/、js/、favicon.ico
		}
	})

	go func() {
		if err := http.Serve(ln, mux); err != nil {
			fail("网页服务停止：" + err.Error())
		}
	}()

	fmt.Println("课表网址：", url)
	fmt.Println()
	fmt.Println("浏览器会自动打开。更新课表：把新的「班级课表.html」")
	fmt.Println("放进这个文件夹覆盖旧的，然后在浏览器按 F5 重新整理。")
	fmt.Println("关闭这个窗口即可结束程序；浏览器页面全部关闭约 3 分钟后也会自动结束。")
	openBrowser(url)

	// 自动结束：页面连上过、之后超过 idleTimeout 没有心跳
	start := time.Now()
	for range time.Tick(15 * time.Second) {
		lp := lastPing.Load()
		if lp == 0 {
			if time.Since(start) > 10*time.Minute {
				fmt.Println("没有页面连上，程序结束。")
				return
			}
			continue
		}
		if time.Since(time.Unix(lp, 0)) > idleTimeout {
			fmt.Println("浏览器页面已关闭，程序结束。")
			return
		}
	}
}

// 找课表来源：优先「班级课表.html」，否则找文件夹里最新的、内容是 eSchool 班级课表的 .html
func findSource(dir string) string {
	for _, name := range []string{"班级课表.html", "班级课表.htm", "班級課表.html"} {
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
		b, err := os.ReadFile(p)
		if err != nil || !strings.Contains(string(b), "printarea--ref-csche") {
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

func isOurs(url string) bool {
	c := http.Client{Timeout: 800 * time.Millisecond}
	r, err := c.Get(url + "__ping")
	if err != nil {
		return false
	}
	defer r.Body.Close()
	buf := make([]byte, 64)
	n, _ := r.Body.Read(buf)
	return string(buf[:n]) == appID
}

func noCache(w http.ResponseWriter) {
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")
}

func openBrowser(url string) {
	var err error
	switch runtime.GOOS {
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = exec.Command("xdg-open", url).Start()
	}
	if err != nil {
		fmt.Println("无法自动打开浏览器，请手动打开：", url)
	}
}

func fail(msg string) {
	fmt.Println("错误：", msg)
	fmt.Println("按 Enter 关闭……")
	fmt.Scanln()
	os.Exit(1)
}
