// Teacher-TimeTable.exe —— 循人中学 教师与班级课表
//
// 每次启动：在本程序所在的文件夹里找「班级课表.html」（以及可选的英文系统
// 「场地课表_English.html」），开一个小型网页服务，并用默认浏览器打开课表页面。
// 页面每次打开 / 重新整理都会重新读取课表文件，所以只要替换文件，课表就会自动更新。
//
// 设置写在 exe 旁边的 Teacher-TimeTable.ini（第一次运行时自动建立，见 settings.go）：
//
//	lan  = true   同一网络的同事可以用 http://本机IP:端口/ 访问；程序持续运行，关闭窗口才结束
//	lan  = false  只有本机能访问；浏览器页面全部关闭约 3 分钟后自动结束
//	port = 17380
package main

import (
	"embed"
	"fmt"
	"io"
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
	if src := findSource(dir, kindClass); src != "" {
		fmt.Println("班级课表：", src)
	} else {
		fmt.Println("⚠ 在这个文件夹里找不到「班级课表.html」：")
		fmt.Println("  ", dir)
		fmt.Println("  页面打开后可以手动选择文件。")
	}
	if src := findSource(dir, kindEnglish); src != "" {
		fmt.Println("英文课表：", src)
	} else {
		fmt.Println("（没有「场地课表_English.html」，英文课会以 E1–E12 代号显示）")
	}

	cfg := loadSettings(dir)
	fmt.Println()
	fmt.Println("设置文件：", cfg.path)
	if cfg.note != "" {
		fmt.Println("  ", cfg.note)
	}
	fmt.Printf("  lan  = %v（%s）\n", cfg.lan, map[bool]string{true: "同事可访问", false: "只限本机"}[cfg.lan])
	fmt.Printf("  port = %d\n", cfg.port)
	fmt.Println()

	// 已经有一个课表程序在运行：
	//   设置相同 → 直接打开它的页面
	//   设置不同 → 请它结束，再用新设置启动（这样改了 ini 重新打开就会生效）
	existing := fmt.Sprintf("http://127.0.0.1:%d/", cfg.port)
	if isOurs(existing) {
		if runningCfg(existing) == cfg.key() {
			fmt.Println("课表程序已经在运行（设置相同），直接打开页面。")
			openBrowser(existing)
			time.Sleep(1500 * time.Millisecond)
			return
		}
		fmt.Println("发现旧的课表程序仍在运行，正在关闭它以套用新设置……")
		if !stopRunning(existing) {
			fail("旧的课表程序无法自动关闭（可能是旧版本）。\n请先关闭另一个「课表」黑色窗口，再重新打开。")
		}
	}

	host := "127.0.0.1"
	if cfg.lan {
		host = "0.0.0.0" // 所有网卡，同事才连得到
	}
	ln, err := net.Listen("tcp", fmt.Sprintf("%s:%d", host, cfg.port))
	if err != nil {
		if cfg.lan {
			fail(fmt.Sprintf("端口 %d 已被其他程序占用，同事需要固定的端口才能连进来。\n请在 %s 改用别的 port。", cfg.port, iniName))
		}
		ln, err = net.Listen("tcp", host+":0")
		if err != nil {
			fail("无法启动本机网页服务：" + err.Error())
		}
	}
	port := ln.Addr().(*net.TCPAddr).Port
	url := fmt.Sprintf("http://127.0.0.1:%d/", port)

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
			serveSource(w, dir, kindClass)
		case "/source_en.html":
			serveSource(w, dir, kindEnglish)
		case "/__ping":
			lastPing.Store(time.Now().Unix())
			w.Write([]byte(appID))
		case "/__cfg": // 给新启动的程序比对设置
			w.Write([]byte(cfg.key()))
		case "/__quit": // 只接受本机：新启动的程序请旧的结束
			if !isLoopback(r.RemoteAddr) {
				http.NotFound(w, r)
				return
			}
			w.Write([]byte("bye"))
			go func() {
				time.Sleep(200 * time.Millisecond)
				fmt.Println("已由新启动的课表程序接手，这个窗口结束。")
				os.Exit(0)
			}()
		default:
			static.ServeHTTP(w, r) // index.html、css/、js/、favicon.ico
		}
	})

	go func() {
		if err := http.Serve(ln, mux); err != nil {
			fail("网页服务停止：" + err.Error())
		}
	}()

	fmt.Println("本机网址：", url)
	if cfg.lan {
		ips := lanIPs()
		if len(ips) == 0 {
			fmt.Println("同事网址：（这台电脑目前没有连上网络）")
		}
		for _, ip := range ips {
			fmt.Printf("同事网址： http://%s:%d/   （%s）\n", ip.addr, port, ip.iface)
		}
		fmt.Println()
		fmt.Println("※ 同事连不上时：")
		fmt.Println("  1. Windows 防火墙询问时要按「允许」；已按过「取消」的话，")
		fmt.Println("     以系统管理员身份运行 firewall-allow.bat。")
		fmt.Println("  2. 同事要和你在同一个网络（同一个 Wi-Fi / 同一个网段）。")
	} else {
		fmt.Printf("（只限本机。要让同事访问，把 %s 里的 lan 改成 true，再重新打开）\n", iniName)
	}
	fmt.Println()
	fmt.Println("更新课表：把新的「班级课表.html」或「场地课表_English.html」")
	fmt.Println("放进这个文件夹覆盖旧的，然后在浏览器按 F5 重新整理。")
	if cfg.lan {
		fmt.Println("关闭这个窗口即结束程序（同事也会连不上）。")
	} else {
		fmt.Println("关闭这个窗口即可结束程序；浏览器页面全部关闭约 3 分钟后也会自动结束。")
	}
	openBrowser(url)

	if cfg.lan {
		select {} // 分享模式：一直运行，直到关闭窗口
	}

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

// 课表来源的种类
type sourceKind struct {
	label  string   // 用于讯息
	names  []string // 优先使用的文件名
	marker string   // 找不到时，内容里要有的字（eSchool 列印页的网址）
}

var (
	kindClass   = sourceKind{"班级课表.html", []string{"班级课表.html", "班级课表.htm", "班級課表.html"}, "csche_prt.php"}
	kindEnglish = sourceKind{"场地课表_English.html", []string{"场地课表_English.html", "场地课表_english.html", "場地課表_English.html"}, "schedule_eng/psche_prt.php"}
)

// 每次请求都重新找、重新读，所以替换文件后按 F5 就会更新
func serveSource(w http.ResponseWriter, dir string, k sourceKind) {
	src := findSource(dir, k)
	if src == "" {
		http.Error(w, "找不到 "+k.label, http.StatusNotFound)
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
}

// 找课表来源：优先用固定文件名，否则找文件夹里最新的、内容符合的 .html
func findSource(dir string, k sourceKind) string {
	for _, name := range k.names {
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
		if err != nil || !strings.Contains(string(b), k.marker) {
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

func exePath() string {
	exe, err := os.Executable()
	if err != nil {
		return "Teacher-TimeTable.exe"
	}
	if r, err := filepath.EvalSymlinks(exe); err == nil {
		exe = r
	}
	return exe
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

// 旧程序的设置（旧版本没有 /__cfg，会回传空字串）
func runningCfg(url string) string {
	c := http.Client{Timeout: 800 * time.Millisecond}
	r, err := c.Get(url + "__cfg")
	if err != nil {
		return ""
	}
	defer r.Body.Close()
	if r.StatusCode != http.StatusOK {
		return ""
	}
	b, _ := io.ReadAll(io.LimitReader(r.Body, 128))
	return string(b)
}

// 请旧程序结束，并等到它真的结束（最多 5 秒）
func stopRunning(url string) bool {
	c := http.Client{Timeout: 800 * time.Millisecond}
	if r, err := c.Get(url + "__quit"); err == nil {
		r.Body.Close()
	}
	for i := 0; i < 25; i++ {
		time.Sleep(200 * time.Millisecond)
		if !isOurs(url) {
			time.Sleep(300 * time.Millisecond) // 等端口释放
			return true
		}
	}
	return false
}

func isLoopback(remoteAddr string) bool {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		return false
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
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
