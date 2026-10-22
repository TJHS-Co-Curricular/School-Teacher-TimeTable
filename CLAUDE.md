# CLAUDE.md — 循人课表（TJHS School Teacher TimeTable）

> Claude Code 每次都会读这份文件。改了结构、规范、命令或约束时，同步更新这里。
> 一次性的交接资讯已归档在 `docs/archive/HANDOFF.md`；版本记录在 `CHANGELOG.md`；当前版本见 `VERSION`。
> 仓库：https://github.com/TJHS-Co-Curricular/School-Teacher-TimeTable

## 项目一句话

把循人中学 eSchool 导出的「班级课表」（＋英文系统「场地课表\_English」）整合成 **按老师 / 按班级 / 老师节数总表** 的网页，打包成免安装的 Windows exe，可在校内局域网分享。维护者：ChongZhiJie；使用者：学校老师与行政同事。

## 沟通要求

- **一律使用简体中文**：回复、界面文字、代码注解、日志讯息、文档。
- 说明要讲清楚「使用者自己需要做什么」。
- 做完重要改动：更新 `CHANGELOG.md`，结构或规范有变就更新本文件。

---

## 1. 技术栈与版本

| 项目     | 内容                                                                                          |
| -------- | --------------------------------------------------------------------------------------------- |
| 后端     | **Go 1.22+**，只用标准库（`embed`、`net/http`…），**没有第三方依赖**                          |
| 前端     | 原生 HTML / CSS / JavaScript（ES2020），无框架、无打包工具、无 ES modules                     |
| 解析     | 浏览器 `DOMParser` + `TextDecoder('gb18030')`（eSchool 文件是 GBK）                           |
| 目标平台 | Windows 10/11 x64；浏览器 Chrome / Edge 111+（用了 CSS `color-mix()`）                        |
| 打包     | `GOOS=windows GOARCH=amd64 CGO_ENABLED=0`，单一 exe（约 7 MB）                                |
| exe 资源 | mingw-w64 `windres` 生成 `.syso`（只在 Linux/WSL；可选）                                      |
| 格式化   | `gofmt`；Prettier 3（`.prettierrc.json`：printWidth 120、singleQuote、trailingComma all、LF） |
| Module   | `github.com/TJHS-Co-Curricular/School-Teacher-TimeTable`                                      |
| 硬件     | 无特殊硬件                                                                                    |

## 2. 目录结构

```
cmd/teacher-timetable/
  main.go                    入口，只做组装（启动流程见文件开头注解）
  rsrc_windows_amd64.syso    exe 图标/manifest/版本（由 scripts/build.sh 生成，已提交）
internal/
  buildinfo/                 AppName、ExeName、AppID、Version/Commit/BuildDate（-ldflags 注入）
  logx/                      日志：窗口 + logs/YYYY-MM-DD.log、等级、自动清理
  netutil/                   局域网 IP、ClientIP、IsLoopback
  paths/                     运行时文件夹结构（config/ data/ logs/）、旧版 ini 位置
  platform/                  Windows 命令窗口 UTF-8 与标题、打开浏览器
  server/                    HTTP 路由、读课表、/config.json、单一实例（IsRunning/RunningKey/StopRunning）
  settings/                  settings.ini 解析与 Render、app.json（LoadApp）
  source/                    课表来源 Class / English，data/ → 根目录查找
config/                      默认设定档 settings.ini、app.json + embed.go（打包进 exe）
web/                         网页 + embed.go（打包进 exe）
  index.html  favicon.ico
  css/style.css
  js/config.js → utils.js → parser.js → app.js → main.js   （按此顺序载入）
build/windows/               app.ico、app.manifest、app.rc.in（@VERSION@ 模板）
scripts/                     build.bat、build.sh、windres-pp.sh、firewall-allow.bat
build.bat                    快捷方式 → scripts\build.bat
VERSION  CHANGELOG.md  README.md  CLAUDE.md
dist/                        打包输出（不进 git）
```

### 运行时文件夹（exe 所在处；打包输出 `dist/Teacher-TimeTable/`）

```
Teacher-TimeTable.exe
firewall-allow.bat
config/   settings.ini（程序）、app.json（网页）——不存在时由 exe 写出默认值
data/     班级课表.html（必需）、场地课表_English.html 或 场地课表.html（可选）
logs/     YYYY-MM-DD.log
```

- `settings.ini`：`[server] lan / port / open_browser`，`[log] keep_days / level`。改完重开 exe 生效（新进程会接手旧进程）。
- `app.json`：`appName`、`headerTitle`、`schoolName`、`schoolNameEn`、`githubUrl`。经 `/config.json` 提供，按 F5 生效。
- 环境变量 `TEACHER_TIMETABLE_HOME` 可指定运行时文件夹（开发用）。

### HTTP 路由

| 路由              | 用途                                             |
| ----------------- | ------------------------------------------------ |
| `/`               | 网页（embed）                                    |
| `/source.html`    | 班级课表（每次重新读文件）                       |
| `/source_en.html` | 英文场地课表                                     |
| `/config.json`    | app.json + `version`                             |
| `/__ping`         | 心跳；回传 AppID `tjhs-timetable` 用来辨认本程序 |
| `/__cfg`          | 运行中的设置 `lan=…;port=…`                      |
| `/__version`      | 版本号                                           |
| `/__quit`         | 结束本进程（**只接受 127.0.0.1**）               |

## 3. 领域知识（改解析或统计前必读）

### eSchool 文件格式（GBK）

| 文件                                          | 网址标记（文件开头 `saved from url`） | 每页标签                                   |
| --------------------------------------------- | ------------------------------------- | ------------------------------------------ |
| `班级课表.html`                               | `activity/csche_prt.php`              | `班级：初一A`、`班导师：…`、`列印日期：…`  |
| `场地课表_English.html`（或 `场地课表.html`） | `schedule_eng/psche_prt.php`          | `场地：E001`、`课时总数：…`、`列印日期：…` |

- 每页一个 `div.printarea`；标题在 `.print-tt__info`；课表在 `.print-tt__timetable-box table`。
- 第 1 行节次 1–11（`6*` 有特别安排，说明在 `.print-tt__rmk-box`），第 2 行时间；星期一至六每天 11 节。
- `.print-tt__recess` 是休息栏，**两个系统的位置不同 → 一律按节次索引对齐，不要按栏位**。
- 早自习 / 班导师时间 / 晨读是 `rowspan` 格子，解析时跳过。
- 班级课表格子：`科目<br>老师`；英文场地格子：`英文组<br>课型<br>老师`（如 `B11A / GR / 陈志祥`）。
- `.print-tt__plan` 是每班「班级/教师 · 节数 · 科目」统计表。
- 班级课表除了 59 个真正的班级，还会导出教学分组页（如 `跨阅(1)`、`S3C小专题1`、`ART1`、`MAA`；2026-09-26 的资料共 106 页）。**使用者要求照常显示在「按班级」**，不要过滤。
- 同名的页（如两页「力学」）由 `parser.js` 依序改名为 `力学（2）`，避免按名字查找时互相覆盖。

### 统计规则

- **节数只算星期一至星期五，按时段计**：同一时段合班只算 1 节。
- 不计：早自习、班导师时间、晨读、联课活动、公共选修、共同备课（格子里没有老师）。
- `E1`–`E12` 是英文代号；`COMM1`、`COoMm1`、`Maker7`、`科实A` 等是分组代号 → 归入「分组 / 代号」，不当老师。

### 英文合并

- 英文是跨班分组上课。英文组（如 `B11A`）的上课时段集合 **完全等于** 某几班「英文」时段集合 → 该组学生来自这几班（不要用子集比对）。
- 有英文文件时：英文老师以真名加入、`E1`–`E12` 从老师列表移除、同名老师节数合并；班级页英文格显示「N 组分班上课」并列出英文分组。
- 没有英文文件时：一切照旧，英文显示 E 代号。

### 前端数据结构（`web/js/app.js`）

- `DATA`（parser 输出）：`{ kind, periods, recess, times, pre, classes: [{ n, hr, g, p }], printed, sem, note }`
  - `g[day][period]` = 格子文字数组；`p` = 统计表行 `[名字, 节数, 科目]`
- `T[老师]`：`{ slots: { 'd-p': [{ s, c, x }] }, pair: { '班级\u0001科目': n }, day[5], total, subj[], cls[], eg[], group, homeroom[] }`
- `EG[英文组]`：`{ code, rooms, teachers, comps, slots, classes }`；`EBLOCK[班级]` = 英文组代号数组

## 4. 编码规范与命名约定

### Go

- 每个功能一个 `internal/<功能>/` 包，包名小写单字；`main.go` 只做组装，不放业务逻辑。
- 注解、错误讯息、日志用简体中文；错误用 `fmt.Errorf("…：%w", err)` 包装。
- 日志一律经 `logx`：`Infof`（窗口+文件）、`Warnf`（⚠）、`Errorf`（✖）、`Filef`（只写文件）、`Debugf`（只写文件，level=debug）、`Say`（只印窗口的排版文字）。
- 不加第三方依赖；需要时先问使用者。
- 平台相关代码放 `internal/platform`，用 `_windows.go` / `//go:build !windows` 区分。
- 提交前：`gofmt -w .`、`go vet ./...`、`GOOS=windows go vet ./...`。

### JavaScript

- 每个文件是 IIFE：`(function (TT) { 'use strict'; … TT.xxx = {…}; })((window.TT = window.TT || {}));`
- 共用命名空间 `window.TT`：`TT.config`、`TT.util`、`TT.parser`、`TT.app`。
- **不用 ES modules**（`file://` 下会被挡），不用框架。
- 用模板字串产生 HTML，**所有资料都要经过 `esc()`**。
- 变量 camelCase；常量大写（`WEEKDAYS`、`GROUP_RE`）；中文排序用 `zh()`。

### CSS

- 颜色与尺寸用 `:root` 变量；深色模式在 `@media (prefers-color-scheme: dark)` 与 `[data-theme="dark"]` 重新定义变量。
- 科目颜色用 `.k-<类别>`（`k-cn k-en k-bm k-ma k-sc k-hi k-bz k-ict k-art k-pe k-gd k-prj k-oth`）设定 `--h` 色相；活动类用 `.act`。
- 类名简短、按组件命名：`.tcard`（老师卡）、`.ccard`（班级卡）、`.dhead`（详情头）、`.tt`（课表）、`.prow`（统计行）、`.sum`（总表）。
- 手机（≤760px）课表改为按天列表 `.days`；打印 A4 横向，隐藏 `.top .dnav .noprint`。

### 网页设计规范（使用者确认过）

- 简体中文；浅灰绿底 `#f3f5f2`、白卡片、主色深绿 `#1f5c47`；系统字体（不连外网）。
- 首页是**按钮卡片，不用下拉列表**；老师卡片显示总节数 + 星期一至五小柱状图。
- 时间 24 小时制；休息时段用虚线分隔。
- 页脚：© 年份 学校名 · 版权所有 · 版本号 · GitHub 按钮。

### 其他文件

- `.bat`：CRLF、开头 `chcp 65001 >nul`；其余文件 LF、UTF-8（见 `.gitattributes`、`.editorconfig`）。
- 设定档（ini / json）**只放 `config/`**。
- `.rc` 模板不能写 `//` 注解（windres 直通预处理器不会去掉）。

## 5. 常用命令

```bash
# 检查与格式化
gofmt -w .
go vet ./...
GOOS=windows go vet ./...
npx prettier --write "web/**/*.{html,css,js}"

# 打包（输出 dist/Teacher-TimeTable/，不会动到 config/ data/ logs/）
build.bat                 # Windows（没装 Go 会提示用 winget 安装；会先关闭运行中的 exe）
sh scripts/build.sh       # Linux / WSL / macOS（有 windres 时会依 VERSION 重建 .syso）

# 不打包直接运行（需要 data/ 里有课表）
TEACHER_TIMETABLE_HOME=dist/Teacher-TimeTable go run ./cmd/teacher-timetable          # bash
set TEACHER_TIMETABLE_HOME=dist\Teacher-TimeTable && go run ./cmd/teacher-timetable    # cmd

# 快速检查服务
curl http://127.0.0.1:17380/__version
curl http://127.0.0.1:17380/config.json
```

- 测试：**目前没有自动化测试**。改动后至少手动验证：三个视图、英文合并（以 2026-09-26 的课表：老师 212 位、英文组 118 个全部对应、E1–E12 不出现）、改 settings.ini 重开会接手、F5 读到新课表。新增测试时课表样本必须**匿名化**。
- 发布：改 `VERSION` → 写 `CHANGELOG.md` → 打包 → commit → `git tag vX.Y.Z` → push → GitHub Release 上传 `dist/Teacher-TimeTable` 压缩档（压缩前清空 `data/`、`logs/`）。
- 改了版本号想更新 exe「属性」里的版本：在 Linux/WSL 跑 `scripts/build.sh`（需要 `x86_64-w64-mingw32-windres`），提交新的 `.syso`。

## 6. 绝对不能改动 / 必须遵守

1. **课表资料不可进 git**：`班级课表*.html`、`场地课表*.html`、`*_files/`、`data/`、`dist/`、`logs/` 含老师名字（`.gitignore` 已设定，不可移除）。测试样本必须匿名化。
2. **exe 必须单一文件、免安装**：不可引入 Python、Node 或其他运行时依赖（使用者明确要求）。
3. **统计规则不可擅改**：星期一至五、按时段计、排除项（见 §3）。要改先问使用者。
4. **英文组对应必须用「时段集合完全相等」**，不可改成子集比对。
5. **路由与 AppID 保持兼容**：`/__ping` 回传 `tjhs-timetable`、`/__cfg`、`/__quit`（只接受本机）——新旧版本 exe 靠它们互相辨认与接手。
6. **设定档键名与位置保持兼容**：`config/settings.ini` 的 `lan`、`port`、`open_browser`、`keep_days`、`level`；默认端口 17380（使用者实际用 5001）；继续支持旧版无 section 的 ini 与 `Teacher-TimeTable.ini` 迁移。
7. **运行时文件夹结构**固定为 exe + `config/ data/ logs/`；继续支持课表放在 exe 旁边（旧放法）。
8. **版本号只来自 `VERSION`**，不在代码里写死。
9. **`file://` 模式要能用**：直接双击 `web/index.html` 可选择文件（所以前端不用 ES modules、不依赖服务端才有的功能）。
10. **界面一律简体中文**；首页不用下拉列表。
11. **不要擅自删除使用者的文件**；需要删除时列出清单让使用者自己删。
12. 不加第三方 Go 依赖、不加前端框架，除非使用者同意。
