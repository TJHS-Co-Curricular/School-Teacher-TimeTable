# 循人课表 · TJHS School Teacher TimeTable

把 eSchool 导出的「班级课表」整合成 **按老师**、**按班级**、**老师节数总表** 三种视图，
并合并英文系统的「场地课表\_English」（英文课跨班分组上课）。
打包成单一 `Teacher-TimeTable.exe`，不需要安装 Python 或其他软件；每次打开都重新读取课表。

当前版本见 [`VERSION`](VERSION)，更新记录见 [`CHANGELOG.md`](CHANGELOG.md)。

## 使用

运行时文件夹（打包后在 `dist/Teacher-TimeTable/`，整个文件夹可以复制到任何电脑）：

```
Teacher-TimeTable/
├── Teacher-TimeTable.exe
├── firewall-allow.bat      同事连不上时，以系统管理员身份运行
├── config/                 设定档（第一次运行自动建立）
│   ├── settings.ini        程序设置：局域网分享、端口、日志
│   └── app.json            网页显示：学校名称、GitHub 网址
├── data/                   课表文件放这里
│   ├── 班级课表.html        （必需）
│   └── 场地课表_English.html（可选）
└── logs/                   日志，每天一个文件：2026-09-26.log
```

1. 在 eSchool 打开「班级课表」列印页面，浏览器「另存为」→ `data/班级课表.html`。
2. （可选）在英文系统打开「场地课表」列印页面，「另存为」→ `data/场地课表_English.html`。
3. 双击 `Teacher-TimeTable.exe`，浏览器会自动打开课表。
4. 更新课表：覆盖 `data/` 里的文件，在浏览器按 F5。

不用 exe 也可以：直接打开 `web/index.html`，再选择课表文件（可一次选两个）。

## 设定档

### `config/settings.ini`（程序）

改完重新双击 exe 就生效（会自动关闭旧的窗口）。启动时窗口会显示读到的设置，写错的行会列出来并略过。

```ini
[server]
lan = true           ; true = 同事可访问（一直运行）；false = 只限本机（页面关闭 3 分钟后自动结束）
port = 17380
open_browser = true  ; 启动时自动打开浏览器

[log]
keep_days = 30       ; 日志保留天数
level = info         ; debug / info / warn
```

### `config/app.json`（网页）

改完在浏览器按 F5 就生效，不用重新打包。

```json
{
  "appName": "循人课表",
  "headerTitle": "循人中学 · 课表",
  "schoolName": "循人中学",
  "schoolNameEn": "Tsun Jin High School",
  "githubUrl": "https://github.com/TJHS-Co-Curricular/School-Teacher-TimeTable"
}
```

`githubUrl` 留空（`""`）会隐藏页脚的 GitHub 按钮。

## 让同事访问（局域网分享）

`lan = true` 时，窗口会显示「同事网址」，例如 `http://192.168.0.25:17380/`，同一个网络的同事用浏览器打开即可。

- 第一次运行时 Windows 防火墙会询问，请按 **允许访问**。
- 按过「取消」、或同事还是连不上：以系统管理员身份运行 `firewall-allow.bat`（会清掉旧的封锁规则再允许）。
- 同事必须和你在同一个网络；有些学校 Wi-Fi 会隔离设备（AP isolation），那种情况只能改用有线网络或请网管开放。
- 同事第一次连进来时，日志会记录「同事 192.168.0.31 开始查看课表」。

## 日志

`logs/YYYY-MM-DD.log`，每行有时间与等级：

```
2026-09-26 08:15:02 [INFO ] ===== 启动 循人课表 v1.3.0 (a1b2c3d, 2026-09-26) =====
2026-09-26 08:15:02 [INFO ] 课表：D:\Teacher-TimeTable\data\班级课表.html
2026-09-26 08:15:40 [INFO ] 同事 192.168.0.31 开始查看课表
2026-09-26 08:15:40 [INFO ] 已读取 班级课表.html（2.3 MB）← 192.168.0.31
```

超过 `keep_days` 天的日志会在启动时自动删除。有问题时请附上当天的日志。

## 英文课如何合并

教务处的班级课表里，英文课只写代号 E1–E12。英文系统的场地课表则按英文组（如 `B11A`、`P26C`）记录老师、课室和课型（GR / WR / SP/RD / WB …）。
程序把每个英文组的上课时段，与各班「英文」的时段完全比对，找出这个组的学生来自哪几班：

- **按老师**：英文老师以真实名字出现，格子显示英文组、课型和课室；E1–E12 代号不再出现。
- **按班级**：英文格子显示「N 组分班上课」，页面下方列出所有英文组、老师和课室。
- **节数总表**：英文老师的节数一并统计；同时在两个系统都有课的老师会自动合并。

没有 `场地课表_English.html` 时，一切照旧，英文课以 E1–E12 代号显示。

## 项目结构

```
.
├── cmd/teacher-timetable/     程序入口（main.go）与 exe 资源（.syso）
├── internal/                  程序内部代码
│   ├── buildinfo/             程序名称、版本号（打包时注入）
│   ├── logx/                  日志：窗口 + logs/YYYY-MM-DD.log
│   ├── netutil/               本机 IP、来源判断
│   ├── paths/                 运行时文件夹结构（config/ data/ logs/）
│   ├── platform/              Windows 命令窗口、打开浏览器
│   ├── server/                网页服务与路由、单一实例
│   ├── settings/              读取 settings.ini / app.json
│   └── source/                寻找课表文件
├── config/                    默认设定档（打包进 exe，第一次运行时复制到运行时的 config/）
├── web/                       网页（打包进 exe）
│   ├── index.html  favicon.ico
│   ├── css/style.css
│   └── js/  config.js · utils.js · parser.js · app.js · main.js
├── build/windows/             exe 图标、manifest、版本资讯模板
├── scripts/                   build.bat · build.sh · firewall-allow.bat · windres-pp.sh
├── build.bat                  快捷方式 → scripts/build.bat
├── VERSION                    版本号（唯一来源）
└── CHANGELOG.md
```

## 开发

### 打包

需要 [Go](https://go.dev/dl/) 1.22 以上（Windows 也可用 `winget install GoLang.Go`）。

- Windows：双击根目录的 `build.bat`（没装 Go 会提示自动安装；exe 正在运行时会先关闭它）
- Linux / macOS / WSL：`sh scripts/build.sh`

输出为 `dist/Teacher-TimeTable/`。重新打包只会替换 exe，不会动到 `config/`、`data/`、`logs/`。

### 改版流程

1. 修改代码
2. 改 `VERSION`（例如 `1.3.1`）并在 `CHANGELOG.md` 记一笔
3. 运行 `build.bat` 打包
4. `git commit`，并用 GitHub Releases 发布 `dist/Teacher-TimeTable/` 的压缩档

### 直接运行（不打包）

```sh
TEACHER_TIMETABLE_HOME=dist/Teacher-TimeTable go run ./cmd/teacher-timetable
```

`TEACHER_TIMETABLE_HOME` 指定运行时文件夹；不设时为 exe 所在的文件夹。

### 格式

- Go：`gofmt -w .`、`go vet ./...`
- 网页：`npx prettier --write "web/**/*.{html,css,js}"`

## 隐私

`班级课表.html`、`场地课表_English.html` 含老师名字；`data/`、`dist/`、`logs/` 已列入 `.gitignore`，不会上传到 GitHub。
