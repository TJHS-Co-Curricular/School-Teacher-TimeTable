# 循人课表 · TJHS School Teacher TimeTable

把 eSchool 导出的「班级课表」整合成 **按老师**、**按班级** 和 **老师节数总表** 三种视图。
每次打开都会重新读取「班级课表.html」，替换文件即可更新。

## 使用

1. 在 eSchool 打开「班级课表」列印页面，浏览器「另存为」→ `班级课表.html`。
2. 把 `班级课表.html` 和 `Teacher-TimeTable.exe` 放在同一个文件夹。
3. 双击 `Teacher-TimeTable.exe`，浏览器会自动打开课表。
4. 更新课表：覆盖 `班级课表.html`，在浏览器按 F5。

- 黑色窗口关闭即结束；浏览器页面全部关闭约 3 分钟后也会自动结束。
- 不用 exe 也可以：直接打开 `source/web/index.html`，再选择 `班级课表.html`。

## 项目结构

```
.
├── README.md
├── .editorconfig / .gitattributes / .gitignore / .prettierrc.json
└── source/
    ├── main.go                  本机网页服务：提供 web/ 与 /source.html（即 班级课表.html）
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
            ├── parser.js        解析 eSchool 班级课表 HTML（GBK）
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

`班级课表.html` 含老师名字，已列入 `.gitignore`，不会上传到 GitHub。
