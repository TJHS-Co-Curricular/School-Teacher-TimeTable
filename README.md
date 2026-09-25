# 循人课表 · TJHS School Teacher TimeTable

把 eSchool 导出的「班级课表」整合成 **按老师**、**按班级** 和 **老师节数总表** 三种视图，
并合并英文系统的「场地课表_English」（英文课跨班分组上课）。
每次打开都会重新读取这两个文件，替换文件即可更新。

## 使用

1. 在 eSchool 打开「班级课表」列印页面，浏览器「另存为」→ `班级课表.html`。
2. （可选）在英文系统打开「场地课表」列印页面，「另存为」→ `场地课表_English.html`。
3. 把这两个文件和 `Teacher-TimeTable.exe` 放在同一个文件夹。
4. 双击 `Teacher-TimeTable.exe`，浏览器会自动打开课表。
5. 更新课表：覆盖对应的文件，在浏览器按 F5。

- 黑色窗口关闭即结束；浏览器页面全部关闭约 3 分钟后也会自动结束。
- 不用 exe 也可以：直接打开 `source/web/index.html`，再选择 `班级课表.html`（可同时选 `场地课表_English.html`）。

## 让同事访问（局域网分享）

默认开启。启动后黑色窗口会显示「同事网址」，例如 `http://192.168.0.25:17380/`，同一个网络的同事用浏览器打开即可。

- 第一次运行时 Windows 防火墙会询问，请按 **允许访问**。
- 之前按过「取消」、或同事还是连不上：以系统管理员身份运行根目录的 `firewall-allow.bat`（会清掉旧的封锁规则再允许）。
- 同事必须和你在同一个网络；有些学校 Wi-Fi 会隔离设备（AP isolation），那种情况只能改用有线网络或请网管开放。
- 分享模式下程序会一直运行，关闭黑色窗口才结束。

设置在 exe 旁边的 `Teacher-TimeTable.ini`（第一次运行自动建立）：

```ini
lan = true     ; false = 只限本机，页面关闭 3 分钟后自动结束
port = 17380
```

改完设置后重新双击 exe 即可生效：程序会自动关闭旧的窗口，再用新设置启动。启动时黑色窗口会显示读到的设置；写错的行会列出来并略过。

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
├── README.md
├── .editorconfig / .gitattributes / .gitignore / .prettierrc.json
└── source/
    ├── main.go                  本机网页服务：提供 web/、/source.html（班级课表）、/source_en.html（英文场地课表）
    ├── settings.go              读取 Teacher-TimeTable.ini、列出本机 IP
    ├── console_windows.go       Windows 命令窗口 UTF-8
    ├── console_other.go
    ├── go.mod
    ├── app.rc / app.ico / app.manifest / rsrc_windows_amd64.syso   exe 图标与版本资讯
    ├── build.bat                一键打包 exe（Windows）
    ├── build.sh / pp.sh         打包脚本（Linux / macOS）
    └── web/                     网页（编译时打包进 exe）
        ├── index.html
        ├── favicon.ico
        ├── css/style.css
        └── js/
            ├── config.js        设置：学校名称、GitHub 网址等
            ├── utils.js         共用小工具
            ├── parser.js        解析 eSchool 班级课表 / 英文场地课表 HTML（GBK）
            ├── app.js           界面与三个视图
            └── main.js          启动：读取课表、文件选择、心跳
```

## 修改

- 学校名称、GitHub 网址：`source/web/js/config.js`
- 样式：`source/web/css/style.css`
- 格式化：`npx prettier --write "source/web/**/*.{html,css,js}"`

## 打包 exe

需要 [Go](https://go.dev/dl/)（Windows 也可用 `winget install GoLang.Go`）。

- Windows：双击 `source/build.bat`（没装 Go 时会提示用 winget 自动安装；exe 正在运行时会先关闭它）
- Linux / macOS：`sh source/build.sh`

输出为根目录的 `Teacher-TimeTable.exe`，单一文件，不需要安装 Python 或其他软件。

## 隐私

`班级课表.html`、`场地课表_English.html` 含老师名字，已列入 `.gitignore`，不会上传到 GitHub。
