// Teacher-TimeTable.exe —— 循人中学 教师与班级课表
//
// 启动流程：
//  1. 准备运行时文件夹（config/、data/、logs/），旧版设置自动搬到 config/
//  2. 读取 config/settings.ini，开始写日志
//  3. 检查课表文件（data/班级课表.html、data/场地课表_English.html）
//  4. 已有本程序在运行：设置相同就直接打开页面；不同就请它结束再接手
//  5. 启动网页服务、打开浏览器
//  6. lan=true 一直运行；lan=false 页面全部关闭约 3 分钟后自动结束
package main

import (
	"errors"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	defaults "github.com/TJHS-Co-Curricular/School-Teacher-TimeTable/config"
	"github.com/TJHS-Co-Curricular/School-Teacher-TimeTable/internal/buildinfo"
	"github.com/TJHS-Co-Curricular/School-Teacher-TimeTable/internal/logx"
	"github.com/TJHS-Co-Curricular/School-Teacher-TimeTable/internal/netutil"
	"github.com/TJHS-Co-Curricular/School-Teacher-TimeTable/internal/paths"
	"github.com/TJHS-Co-Curricular/School-Teacher-TimeTable/internal/platform"
	"github.com/TJHS-Co-Curricular/School-Teacher-TimeTable/internal/server"
	"github.com/TJHS-Co-Curricular/School-Teacher-TimeTable/internal/settings"
	"github.com/TJHS-Co-Curricular/School-Teacher-TimeTable/internal/source"
	"github.com/TJHS-Co-Curricular/School-Teacher-TimeTable/web"
)

const idleTimeout = 3 * time.Minute

var log *logx.Logger

func main() {
	title := buildinfo.AppName + " " + buildinfo.String()
	platform.SetupConsole(title)

	// 1. 运行时文件夹
	layout := paths.Detect()
	log = logx.New("", logx.Info)
	if err := layout.Ensure(); err != nil {
		fail("无法建立运行时文件夹：%v", err)
	}
	log = logx.New(layout.Logs, logx.Info)
	defer log.Close()

	log.Say("==============================================")
	log.Say("  %s  %s", buildinfo.AppName, buildinfo.String())
	log.Say("==============================================")
	log.Filef("===== 启动 %s %s =====", buildinfo.AppName, buildinfo.String())
	log.Filef("运行时文件夹：%s", layout.Root)

	migrateLegacySettings(layout)

	// 2. 设置
	st, err := settings.Load(layout.SettingsFile(), defaults.SettingsINI)
	if err != nil {
		log.Warnf("%v（使用默认设置）", err)
	}
	log.SetLevel(logx.ParseLevel(st.LogLevel))
	if n := log.Cleanup(st.LogKeepDays); n > 0 {
		log.Filef("已删除 %d 个超过 %d 天的旧日志", n, st.LogKeepDays)
	}
	log.Say("")
	log.Infof("设置文件：%s%s", st.Path, map[bool]string{true: "（第一次运行，已建立）", false: ""}[st.Created])
	for _, is := range st.Issues {
		log.Warnf("设置有误，已略过 %s", is)
	}
	log.Infof("  lan = %v（%s）  port = %d  open_browser = %v",
		st.LAN, map[bool]string{true: "同事可访问", false: "只限本机"}[st.LAN], st.Port, st.OpenBrowser)
	if _, err := settings.LoadApp(layout.AppFile(), defaults.AppJSON); err != nil {
		log.Warnf("%v", err)
	}

	// 3. 课表文件
	log.Say("")
	checkSource(layout, source.Class, "页面打开后可以手动选择文件")
	checkSource(layout, source.English, "英文课会以 E1–E12 代号显示")

	// 4. 单一实例
	local := server.URL("127.0.0.1", st.Port)
	if server.IsRunning(local) {
		if server.RunningKey(local) == st.Key() {
			log.Infof("课表程序已经在运行（设置相同），直接打开页面。")
			openBrowser(local)
			time.Sleep(1500 * time.Millisecond)
			return
		}
		log.Infof("发现旧的课表程序仍在运行，正在关闭它以套用新设置……")
		if !server.StopRunning(local) {
			fail("旧的课表程序无法自动关闭（可能是旧版本）。\n请先关闭另一个课表的黑色窗口，再重新打开。")
		}
	}

	// 5. 网页服务
	host := "127.0.0.1"
	if st.LAN {
		host = "0.0.0.0" // 所有网卡，同事才连得到
	}
	ln, err := net.Listen("tcp", fmt.Sprintf("%s:%d", host, st.Port))
	if err != nil {
		if st.LAN {
			fail("端口 %d 已被其他程序占用，同事需要固定的端口才能连进来。\n请在 %s 改用别的 port。", st.Port, st.Path)
		}
		log.Warnf("端口 %d 已被占用，改用随机端口", st.Port)
		if ln, err = net.Listen("tcp", "127.0.0.1:0"); err != nil {
			fail("无法启动网页服务：%v", err)
		}
	}
	port := ln.Addr().(*net.TCPAddr).Port
	local = server.URL("127.0.0.1", port)

	webRoot, _ := fs.Sub(web.Files, ".")
	srv := &server.Server{
		Layout:      layout,
		Settings:    st,
		Log:         log,
		Web:         webRoot,
		AppDefaults: defaults.AppJSON,
		OnQuit: func() {
			time.Sleep(200 * time.Millisecond)
			exit(0, "已由新启动的课表程序接手，这个窗口结束。")
		},
	}
	go func() {
		if err := http.Serve(ln, srv.Handler()); err != nil && !errors.Is(err, net.ErrClosed) {
			fail("网页服务停止：%v", err)
		}
	}()

	log.Say("")
	log.Infof("本机网址：%s", local)
	if st.LAN {
		addrs := netutil.LANAddrs()
		if len(addrs) == 0 {
			log.Warnf("同事网址：这台电脑目前没有连上网络")
		}
		for _, a := range addrs {
			log.Infof("同事网址：%s   （%s）", server.URL(a.IP, port), a.Interface)
		}
		log.Say("")
		log.Say("※ 同事连不上时：")
		log.Say("  1. Windows 防火墙询问时要按「允许」；按过「取消」的话，")
		log.Say("     以系统管理员身份运行 firewall-allow.bat。")
		log.Say("  2. 同事要和你在同一个网络（同一个 Wi-Fi / 同一个网段）。")
	} else {
		log.Say("（只限本机。要让同事访问，把 config\\settings.ini 里的 lan 改成 true，再重新打开）")
	}
	log.Say("")
	log.Say("更新课表：把新的课表文件放进 data 文件夹覆盖旧的，然后在浏览器按 F5。")
	if st.LAN {
		log.Say("关闭这个窗口即结束程序（同事也会连不上）。")
	} else {
		log.Say("关闭这个窗口即可结束程序；浏览器页面全部关闭约 3 分钟后也会自动结束。")
	}
	log.Say("日志：%s", layout.Logs)

	if st.OpenBrowser {
		openBrowser(local)
	}

	// 关闭窗口 / Ctrl+C 时写一行日志
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sig
		exit(0, "窗口已关闭，程序结束。")
	}()

	// 6. 结束条件
	if st.LAN {
		select {} // 分享模式：一直运行，直到关闭窗口
	}
	start := time.Now()
	for range time.Tick(15 * time.Second) {
		last := srv.LastPing()
		switch {
		case last.IsZero() && time.Since(start) > 10*time.Minute:
			exit(0, "没有页面连上，程序结束。")
		case !last.IsZero() && time.Since(last) > idleTimeout:
			exit(0, "浏览器页面已关闭，程序结束。")
		}
	}
}

// v1.2 的设置放在 exe 旁边（Teacher-TimeTable.ini）：
// 把值套进新格式写到 config/settings.ini，旧文件改名为 .bak
func migrateLegacySettings(layout paths.Layout) {
	old := layout.LegacySettings()
	if old == "" {
		return
	}
	prev, err := settings.Load(old, defaults.SettingsINI)
	if err == nil {
		err = os.WriteFile(layout.SettingsFile(), settings.Render(defaults.SettingsINI, prev), 0o644)
	}
	if err == nil {
		err = os.Rename(old, old+".bak")
	}
	if err != nil {
		log.Warnf("旧设置文件无法搬到 config/：%v", err)
		return
	}
	log.Infof("已把旧设置（lan=%v, port=%d）搬到 config/settings.ini，旧文件改名为 %s.bak", prev.LAN, prev.Port, filepath.Base(old))
}

func checkSource(layout paths.Layout, k source.Kind, missingHint string) {
	res := source.Find(layout.Data, layout.Root, k)
	switch {
	case res.Path == "":
		log.Warnf("找不到 %s（请放进 %s）——%s", k.Label, layout.Data, missingHint)
	case res.Legacy:
		log.Infof("课表：%s", res.Path)
		log.Warnf("建议把 %s 移到 data 文件夹", k.Label)
	default:
		log.Infof("课表：%s", res.Path)
	}
}

func openBrowser(url string) {
	if err := platform.OpenBrowser(url); err != nil {
		log.Warnf("无法自动打开浏览器，请手动打开：%s", url)
	}
}

func exit(code int, msg string) {
	log.Infof("%s", msg)
	log.Filef("===== 结束 =====")
	log.Close()
	os.Exit(code)
}

func fail(format string, a ...any) {
	log.Errorf(format, a...)
	log.Filef("===== 异常结束 =====")
	log.Close()
	fmt.Print("按 Enter 关闭……")
	fmt.Scanln()
	os.Exit(1)
}
