package paths

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectUsesEnv(t *testing.T) {
	root := t.TempDir()
	t.Setenv(envHome, root)
	l := Detect()
	if l.Root != root {
		t.Fatalf("Root = %s，应为 %s", l.Root, root)
	}
	want := map[string]string{
		l.Config:         filepath.Join(root, "config"),
		l.Data:           filepath.Join(root, "data"),
		l.Logs:           filepath.Join(root, "logs"),
		l.SettingsFile(): filepath.Join(root, "config", "settings.ini"),
		l.AppFile():      filepath.Join(root, "config", "app.json"),
	}
	for got, w := range want {
		if got != w {
			t.Errorf("得到 %s，应为 %s", got, w)
		}
	}
}

func TestEnsure(t *testing.T) {
	t.Setenv(envHome, t.TempDir())
	l := Detect()
	if err := l.Ensure(); err != nil {
		t.Fatal(err)
	}
	for _, d := range []string{l.Config, l.Data, l.Logs} {
		if fi, err := os.Stat(d); err != nil || !fi.IsDir() {
			t.Errorf("没有建立文件夹 %s", d)
		}
	}
	readme := filepath.Join(l.Data, "请把课表放在这里.txt")
	if _, err := os.Stat(readme); err != nil {
		t.Fatal("data/ 里应该有说明文件")
	}

	// 已经存在的说明文件不覆盖
	if err := os.WriteFile(readme, []byte("老师自己改过"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := l.Ensure(); err != nil {
		t.Fatal(err)
	}
	if b, _ := os.ReadFile(readme); string(b) != "老师自己改过" {
		t.Fatal("Ensure 不可以覆盖已有的说明文件")
	}
}

// v1.2 的 Teacher-TimeTable.ini 只在 config/settings.ini 还没建立时才需要搬家
func TestLegacySettings(t *testing.T) {
	t.Setenv(envHome, t.TempDir())
	l := Detect()
	if err := l.Ensure(); err != nil {
		t.Fatal(err)
	}
	if p := l.LegacySettings(); p != "" {
		t.Fatalf("没有旧文件时应回传空字串，得到 %s", p)
	}

	old := filepath.Join(l.Root, "Teacher-TimeTable.ini")
	if err := os.WriteFile(old, []byte("lan=false\r\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if p := l.LegacySettings(); p != old {
		t.Fatalf("有旧文件、没有新文件时应回传 %s，得到 %q", old, p)
	}

	if err := os.WriteFile(l.SettingsFile(), []byte("[server]\r\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if p := l.LegacySettings(); p != "" {
		t.Fatalf("新文件已存在时不该再搬家，得到 %s", p)
	}
}
