// Package server 是本机网页服务：提供打包在 exe 里的网页，以及每次重新读取的课表文件。
//
// 路由：
//
//	/                 网页（index.html、css/、js/、favicon.ico）
//	/source.html      班级课表（data/班级课表.html）
//	/source_en.html   英文场地课表（data/场地课表_English.html）
//	/config.json      网页显示设置（config/app.json + 版本号）
//	/__ping           心跳；也用来辨认「正在运行的是不是本程序」
//	/__cfg            正在运行的设置（给新启动的程序比对）
//	/__quit           请本程序结束（只接受本机）
package server

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"github.com/TJHS-Co-Curricular/School-Teacher-TimeTable/internal/buildinfo"
	"github.com/TJHS-Co-Curricular/School-Teacher-TimeTable/internal/logx"
	"github.com/TJHS-Co-Curricular/School-Teacher-TimeTable/internal/netutil"
	"github.com/TJHS-Co-Curricular/School-Teacher-TimeTable/internal/paths"
	"github.com/TJHS-Co-Curricular/School-Teacher-TimeTable/internal/settings"
	"github.com/TJHS-Co-Curricular/School-Teacher-TimeTable/internal/source"
)

// Server 保存网页服务需要的一切。
type Server struct {
	Layout      paths.Layout
	Settings    settings.Settings
	Log         *logx.Logger
	Web         fs.FS  // 网页文件
	AppDefaults []byte // 默认 app.json
	OnQuit      func() // 收到 /__quit 时调用

	lastPing atomic.Int64 // unix 秒；0 = 还没有页面连上
	clients  sync.Map     // 看过课表的 IP（每次运行只记录一次）
}

// LastPing 回传最后一次心跳的时间（没有则为零值）。
func (s *Server) LastPing() time.Time {
	if v := s.lastPing.Load(); v != 0 {
		return time.Unix(v, 0)
	}
	return time.Time{}
}

// Handler 回传所有路由。
func (s *Server) Handler() http.Handler {
	static := http.FileServer(http.FS(s.Web))
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")
		switch r.URL.Path {
		case source.Class.Route:
			s.serveSource(w, r, source.Class)
		case source.English.Route:
			s.serveSource(w, r, source.English)
		case "/config.json":
			s.serveAppConfig(w)
		case "/__ping":
			s.lastPing.Store(time.Now().Unix())
			w.Write([]byte(buildinfo.AppID))
		case "/__cfg":
			w.Write([]byte(s.Settings.Key()))
		case "/__version":
			w.Write([]byte(buildinfo.Version))
		case "/__quit":
			if !netutil.IsLoopback(r.RemoteAddr) {
				http.NotFound(w, r)
				return
			}
			w.Write([]byte("bye"))
			if s.OnQuit != nil {
				go s.OnQuit()
			}
		default:
			static.ServeHTTP(w, r)
		}
	})
	return mux
}

// 每次请求都重新找、重新读，所以替换文件后按 F5 就会更新
func (s *Server) serveSource(w http.ResponseWriter, r *http.Request, k source.Kind) {
	ip := netutil.ClientIP(r.RemoteAddr)
	res := source.Find(s.Layout.Data, s.Layout.Root, k)
	if res.Path == "" {
		s.Log.Debugf("找不到 %s（%s 请求）", k.Label, ip)
		http.Error(w, "找不到 "+k.Label, http.StatusNotFound)
		return
	}
	b, err := os.ReadFile(res.Path)
	if err != nil {
		s.Log.Errorf("读取 %s 失败：%v", res.Path, err)
		http.Error(w, "读取失败："+err.Error(), http.StatusInternalServerError)
		return
	}
	who := "本机"
	if !netutil.IsLoopback(r.RemoteAddr) {
		who = ip
		if _, seen := s.clients.LoadOrStore(ip, true); !seen {
			s.Log.Infof("同事 %s 开始查看课表", ip)
		}
	}
	s.Log.Filef("已读取 %s（%.1f MB）← %s", filepath.Base(res.Path), float64(len(b))/1048576, who)
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(b)
}

// app.json 每次重新读，改了按 F5 就生效
func (s *Server) serveAppConfig(w http.ResponseWriter) {
	app, err := settings.LoadApp(s.Layout.AppFile(), s.AppDefaults)
	if err != nil {
		s.Log.Warnf("%v", err)
	}
	app["version"] = buildinfo.Version
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if err := json.NewEncoder(w).Encode(app); err != nil {
		s.Log.Errorf("回传 config.json 失败：%v", err)
	}
}

// ---------- 单一实例 ----------

var probe = http.Client{Timeout: 800 * time.Millisecond}

func get(url string) (string, bool) {
	r, err := probe.Get(url)
	if err != nil {
		return "", false
	}
	defer r.Body.Close()
	if r.StatusCode != http.StatusOK {
		return "", false
	}
	buf := make([]byte, 128)
	n, _ := r.Body.Read(buf)
	return string(buf[:n]), true
}

// IsRunning 判断 base（例如 http://127.0.0.1:17380/）上运行的是不是本程序。
func IsRunning(base string) bool {
	s, ok := get(base + "__ping")
	return ok && s == buildinfo.AppID
}

// RunningKey 回传正在运行的程序的设置（旧版本没有 /__cfg，会回传空字串）。
func RunningKey(base string) string {
	s, _ := get(base + "__cfg")
	return s
}

// StopRunning 请正在运行的程序结束，并等它真的结束（最多 5 秒）。
func StopRunning(base string) bool {
	get(base + "__quit")
	for i := 0; i < 25; i++ {
		time.Sleep(200 * time.Millisecond)
		if !IsRunning(base) {
			time.Sleep(300 * time.Millisecond) // 等端口释放
			return true
		}
	}
	return false
}

// URL 组合网址。
func URL(host string, port int) string { return fmt.Sprintf("http://%s:%d/", host, port) }
