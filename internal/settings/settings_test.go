package settings

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unicode/utf16"

	defaults "github.com/TJHS-Co-Curricular/School-Teacher-TimeTable/config"
)

// 在临时文件夹写一个设置文件，回传完整路径
func writeINI(t *testing.T, name string, b []byte) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(p, b, 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestDefaultINIHasNoIssues(t *testing.T) {
	p := writeINI(t, "settings.ini", defaults.SettingsINI)
	s, err := Load(p, defaults.SettingsINI)
	if err != nil {
		t.Fatal(err)
	}
	if len(s.Issues) > 0 {
		t.Fatalf("默认设置文件不该有错误：%v", s.Issues)
	}
	if !s.LAN || s.Port != 17380 || !s.OpenBrowser || s.LogKeepDays != 30 || s.LogLevel != "info" {
		t.Fatalf("默认值不对：%+v", s)
	}
}

func TestLoadCreatesMissingFile(t *testing.T) {
	p := filepath.Join(t.TempDir(), "settings.ini")
	s, err := Load(p, defaults.SettingsINI)
	if err != nil {
		t.Fatal(err)
	}
	if !s.Created {
		t.Fatal("文件不存在时应该建立，Created 要是 true")
	}
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Count(string(b), "\n") != strings.Count(string(b), "\r\n") {
		t.Fatal("写出的设置文件要用 CRLF 换行（记事本）")
	}
	if !strings.Contains(string(b), "port = 17380") {
		t.Fatal("写出的设置文件应该是默认内容")
	}
}

func TestLoadSections(t *testing.T) {
	ini := "[server]\r\nlan = false\r\nport = 5001 ; 行尾注解\r\nopen_browser = no\r\n\r\n[log]\r\nkeep_days = 7\r\nlevel = DEBUG\r\n"
	s, err := Load(writeINI(t, "settings.ini", []byte(ini)), defaults.SettingsINI)
	if err != nil {
		t.Fatal(err)
	}
	if s.LAN || s.Port != 5001 || s.OpenBrowser || s.LogKeepDays != 7 || s.LogLevel != "debug" {
		t.Fatalf("读到的值不对：%+v", s)
	}
	if len(s.Issues) > 0 {
		t.Fatalf("不该有错误：%v", s.Issues)
	}
	if s.Key() != "lan=false;port=5001" {
		t.Fatalf("Key() = %q", s.Key())
	}
}

func TestBoolValues(t *testing.T) {
	for v, want := range map[string]bool{
		"true": true, "TRUE": true, "1": true, "yes": true, "on": true, "是": true,
		"false": false, "0": false, "No": false, "off": false, "否": false,
	} {
		s, _ := Load(writeINI(t, "settings.ini", []byte("[server]\nlan = "+v+"\n")), defaults.SettingsINI)
		if s.LAN != want || len(s.Issues) > 0 {
			t.Errorf("lan = %s：得到 %v（错误 %v），应为 %v", v, s.LAN, s.Issues, want)
		}
	}
}

func TestBadValuesKeepDefaults(t *testing.T) {
	ini := "[server]\nlan = maybe\nport = abc\nport2 = 1\n\n[log]\nkeep_days = -1\nlevel = loud\n没有等号\n"
	s, _ := Load(writeINI(t, "settings.ini", []byte(ini)), defaults.SettingsINI)
	if !s.LAN || s.Port != 17380 || s.LogKeepDays != 30 || s.LogLevel != "info" {
		t.Fatalf("写错的值应该退回默认：%+v", s)
	}
	if len(s.Issues) != 6 {
		t.Fatalf("应该有 6 个错误，得到 %d：%v", len(s.Issues), s.Issues)
	}
	if !strings.HasPrefix(s.Issues[0], "第 2 行") || !strings.HasPrefix(s.Issues[1], "第 3 行") {
		t.Fatalf("错误要标出行号：%v", s.Issues)
	}
}

func TestPortRange(t *testing.T) {
	for _, v := range []string{"80", "1023", "65536"} {
		s, _ := Load(writeINI(t, "settings.ini", []byte("[server]\nport = "+v+"\n")), defaults.SettingsINI)
		if s.Port != 17380 || len(s.Issues) != 1 {
			t.Errorf("port = %s 应该退回 17380 并提示：%+v", v, s)
		}
	}
}

func TestNotepadEncodings(t *testing.T) {
	text := "[server]\r\nport = 5002\r\n"
	bom := append([]byte{0xEF, 0xBB, 0xBF}, text...)
	le := []byte{0xFF, 0xFE}
	be := []byte{0xFE, 0xFF}
	for _, u := range utf16.Encode([]rune(text)) {
		le = append(le, byte(u), byte(u>>8))
		be = append(be, byte(u>>8), byte(u))
	}
	for name, b := range map[string][]byte{"UTF-8 BOM": bom, "UTF-16 LE": le, "UTF-16 BE": be} {
		s, _ := Load(writeINI(t, "settings.ini", b), defaults.SettingsINI)
		if s.Port != 5002 || len(s.Issues) > 0 {
			t.Errorf("%s：port = %d，错误 %v", name, s.Port, s.Issues)
		}
	}
}

// v1.2 的 Teacher-TimeTable.ini 没有 [section]：键名直接对应到新位置
func TestLegacyBareKeys(t *testing.T) {
	s, _ := Load(writeINI(t, "Teacher-TimeTable.ini", []byte("lan=false\r\nport=5001\r\n")), defaults.SettingsINI)
	if s.LAN || s.Port != 5001 || len(s.Issues) > 0 {
		t.Fatalf("旧版 ini 读取不对：%+v", s)
	}
}

// 旧版 ini 搬家：读旧值 → 套进默认模板（Render）→ 新文件读回来值相同，注解保留
func TestLegacyMigrationRender(t *testing.T) {
	old, _ := Load(writeINI(t, "Teacher-TimeTable.ini", []byte("lan=false\nport=5001\n")), defaults.SettingsINI)
	out := Render(defaults.SettingsINI, old)

	if strings.Count(string(out), "\n") != strings.Count(string(out), "\r\n") {
		t.Fatal("Render 的结果要用 CRLF 换行")
	}
	for _, want := range []string{"[server]", "[log]", "lan = false", "port = 5001", "; 网页服务的端口"} {
		if !strings.Contains(string(out), want) {
			t.Errorf("搬家后的设置文件缺少「%s」", want)
		}
	}
	s, _ := Load(writeINI(t, "settings.ini", out), defaults.SettingsINI)
	if len(s.Issues) > 0 || s.LAN || s.Port != 5001 || !s.OpenBrowser || s.LogKeepDays != 30 || s.LogLevel != "info" {
		t.Fatalf("搬家后读回来不对：%+v", s)
	}
}

func TestLoadApp(t *testing.T) {
	dir := t.TempDir()

	// 不存在：建立并回传默认值
	p := filepath.Join(dir, "app.json")
	m, err := LoadApp(p, defaults.AppJSON)
	if err != nil || m["schoolName"] != "循人中学" {
		t.Fatalf("默认 app.json 不对：%v %v", m, err)
	}
	if _, err := os.Stat(p); err != nil {
		t.Fatal("app.json 不存在时应该建立")
	}

	// 只改一个键：其它保留默认
	if err := os.WriteFile(p, []byte("\xEF\xBB\xBF{\"schoolName\": \"测试学校\"}"), 0o644); err != nil {
		t.Fatal(err)
	}
	m, err = LoadApp(p, defaults.AppJSON)
	if err != nil || m["schoolName"] != "测试学校" || m["appName"] != "循人课表" {
		t.Fatalf("合并不对：%v %v", m, err)
	}

	// 格式错误：回传错误与默认值
	if err := os.WriteFile(p, []byte("{schoolName: 测试"), 0o644); err != nil {
		t.Fatal(err)
	}
	m, err = LoadApp(p, defaults.AppJSON)
	if err == nil || m["schoolName"] != "循人中学" {
		t.Fatalf("格式错误应回传错误与默认值：%v %v", m, err)
	}
}
