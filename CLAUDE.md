# CLAUDE.md — 循人课表（TJHS School Teacher TimeTable）

> 给 Claude 的项目说明。每次开始工作前先读完这份文件；改了结构、规范或需求时，同步更新这份文件。
> 仓库：https://github.com/TJHS-Co-Curricular/School-Teacher-TimeTable
> 当前版本见 `VERSION`，更新记录见 `CHANGELOG.md`。

## 1. 项目是什么

循人中学（Tsun Jin High School，吉隆坡）的课表工具。把 eSchool 导出的课表 HTML 整合成三个视图：

- **按老师**：每位老师星期一至五的课表与每日 / 每周节数
- **按班级**：与 eSchool「班级课表」相同的内容，重新设计的版面
- **老师节数总表**：所有老师的节数，可排序、搜索

打包成单一 `Teacher-TimeTable.exe`（Go，不需要 Python），每次打开 / 按 F5 都重新读取课表文件；可在局域网分享给同事。

使用者：ChongZhiJie（负责维护），使用者是学校老师 / 行政同事。

## 2. 沟通与工作要求

- **一律用简体中文**回复、写界面文字、写注解与文档。
- 使用者会在不同电脑用 Claude；做完重要改动后，更新本文件与 `CHANGELOG.md`。
- 改代码后必须：`gofmt -w .`、`go vet ./...`（含 `GOOS=windows go vet ./...`）、网页用 Prettier 格式化，并实际跑一次验证。
- **写入使用者电脑的文件后，要读回来比对内容**（曾发生写入后内容是旧版本的问题）。
- 不要擅自删除使用者的文件；需要删除时列出清单请使用者自己删。
- 说明要讲清楚「使用者需要自己做什么」。

## 3. 数据来源（eSchool 导出，GBK 编码）

| 文件 | 来源 | 必需 | 每页标签 |
|---|---|---|---|
| `班级课表.html` | 教务处系统 `activity/csche_prt.php` | 是 | `班级：初一A` / `班导师：…` |
| `场地课表_English.html` | 英文系统 `schedule_eng/psche_prt.php` | 否 | `场地：E001` / `课时总数：…` |

共同格式：每页一个 `div.printarea`，`.print-tt__info` 是标题资料，`.print-tt__timetable-box table` 是课表：

- 第 1 行节次（1–11，`6*` 表示第 6 节有特别安排），第 2 行时间，中间 `.print-tt__recess` 是休息栏（只是显示用，**两个系统的休息栏位置不同，要按节次对齐，不要按栏位**）
- 早自习 / 班导师时间 / 晨读是 `rowspan` 的格子，要略过
- 星期一至六，每天 11 节
- 班级课表格子：`科目<br>老师`；英文场地课表格子：`英文组<br>课型<br>老师`（例：`B11A / GR / 陈志祥`）
- `.print-tt__plan` 是每班的「班级/教师 · 节数 · 科目」统计表
- `.print-tt__rmk-box` 是第 6 节午餐安排说明（J1–J3、S1–S3）

### 领域规则

- **节数只算星期一至星期五**，以「时段」计算（同一时段合班只算 1 节）；不含早自习、班导师时间、晨读、联课活动、公共选修、共同备课。
- 班级课表里的 `E1`–`E12` 是英文代号，不是老师；`COMM1`、`Maker7`、`科实A` 等也是分组代号，另列为「分组 / 代号」。
- **英文合并逻辑**：英文是跨班分组上课。把每个英文组（如 `B11A`、`P26C`）的上课时段集合，与每班「英文」时段集合**完全相等**比对，就能找出该组学生来自哪几班（117 组全部可对应）。合并后：
  - 英文老师以真实名字加入老师列表，`E1`–`E12` 移除
  - 同名老师（例：云惟祥）两个系统的节数合并
  - 班级页英文格显示「N 组分班上课」，并列出英文分组、老师、课室
- 没有英文文件时一切照旧，英文以 E 代号显示。
- 课表资料含老师名字：**不可上传到 GitHub**（已在 `.gitignore`）。

## 4. 项目结构（v1.3.0 起的标准）

```
cmd/teacher-timetable/     入口 main.go（只做组装）+ rsrc_windows_amd64.syso（exe 图标/版本）
internal/
  buildinfo/               AppName、Version/Commit/BuildDate（-ldflags 注入）
  logx/                    日志：窗口 + logs/YYYY-MM-DD.log，自动清理旧日志
  netutil/                 本机 IP、来源 IP、是否本机
  paths/                   运行时文件夹结构、旧版设置迁移
  platform/                Windows 命令窗口（UTF-8、标题）、打开浏览器
  server/                  HTTP 路由、单一实例（/__ping /__cfg /__quit）
  settings/                读取 settings.ini（支持 [section]、行尾注解、BOM/UTF-16）与 app.json
  source/                  寻找课表文件（data/ → exe 旁边）
config/                    默认设定档 settings.ini、app.json（embed.go 打包进 exe）
web/                       网页（embed.go 打包进 exe）
  index.html  favicon.ico  css/style.css
  js/config.js → utils.js → parser.js → app.js → main.js（按此顺序载入，共用 window.TT）
build/windows/             app.ico、app.manifest、app.rc.in（@VERSION@ 模板）
scripts/                   build.bat、build.sh、firewall-allow.bat、windres-pp.sh
build.bat                  快捷方式 → scripts/build.bat
VERSION                    版本号唯一来源
CHANGELOG.md  README.md  CLAUDE.md
```

Go module：`github.com/TJHS-Co-Curricular/School-Teacher-TimeTable`，Go 1.22+。

### 运行时文件夹（exe 所在位置，打包输出为 `dist/Teacher-TimeTable/`）

```
Teacher-TimeTable.exe
firewall-allow.bat
config/   settings.ini（程序）、app.json（网页）——第一次运行自动建立
data/     班级课表.html、场地课表_English.html
logs/     YYYY-MM-DD.log
```

- 所有 ini / json 设定档一律放 `config/`。
- `settings.ini`：`[server] lan / port / open_browser`、`[log] keep_days / level`；改完重开 exe 生效（新程序会通过 `/__quit` 自动接手旧窗口）。
- `app.json`：`appName`、`headerTitle`、`schoolName`、`schoolNameEn`、`githubUrl`；经 `/config.json` 提供，改完按 F5 生效。
- 开发时可用环境变量 `TEACHER_TIMETABLE_HOME` 指定运行时文件夹。

### HTTP 路由

`/` 网页 · `/source.html` 班级课表 · `/source_en.html` 英文场地课表 · `/config.json` app.json+version · `/__ping` 心跳 · `/__cfg` 设置比对 · `/__version` · `/__quit`（只接受本机）

### 运行行为

- `lan = true`：监听 0.0.0.0，窗口显示「同事网址」，一直运行到关闭窗口
- `lan = false`：只监听 127.0.0.1，页面全部关闭约 3 分钟后自动结束
- 同事第一次连线写日志：「同事 IP 开始查看课表」
- 直接双击 `web/index.html`（file://）也能用：会出现选择文件画面，可一次选两个文件

## 5. 版本与发布规范

- 版本号只改 `VERSION`（语义化版本），同时在 `CHANGELOG.md` 记一笔。
- 版本号会显示在：窗口标题、日志启动行、网页页脚、exe 属性（需用 `scripts/build.sh` + windres 重新产生 .syso）。
- 打包：Windows 双击 `build.bat`；Linux/WSL `sh scripts/build.sh`。输出 `dist/Teacher-TimeTable/`，重新打包不动 `config/ data/ logs/`。
- exe 不进 git；用 GitHub Releases 发布 `dist/Teacher-TimeTable/` 压缩档。
- 新增功能时维持结构：新 Go 功能放 `internal/<功能>/`，入口保持精简；网页逻辑放对应的 js 文件。

## 6. 网页设计规范

- 简体中文界面；浅灰绿底 `#f3f5f2`、白色卡片、主色深绿 `#1f5c47`；支持深色模式（`prefers-color-scheme` + `data-theme`）。
- 科目固定色相：华文红、英文蓝、马来文橙、数学紫、理科绿、史地褐、商科青、电脑蓝绿、艺术粉、体育黄绿、辅导灰蓝、专题靛；联课活动等用中性灰。
- 首页是按钮卡片（不要下拉列表）；老师卡片显示总节数与星期一至五小柱状图。
- 课表格：时间用 24 小时制，休息用虚线分隔；手机宽度改为按天列表。
- 页脚：© 年份 循人中学 Tsun Jin High School · 版权所有 · 版本号 · GitHub 按钮。
- 可打印（A4 横向），打印时隐藏工具列。

## 7. 已知注意事项

- 使用者的 Windows 电脑：`D:\00-Data\Documents\Github\TJHS-School-Teacher-TimeTable`
- Windows 防火墙：第一次要按「允许」；按过取消就以系统管理员身份跑 `firewall-allow.bat`（会先删除该程序的封锁规则）。学校 Wi-Fi 可能有 AP isolation。
- `build.bat` 使用仓库里已编译的 `.syso`（Windows 没有 windres）；改版号后 exe 属性里的版本需在 Linux/WSL 跑 `build.sh` 更新。
- `.bat` 文件用 CRLF 与 `chcp 65001`；其余文件 LF（见 `.gitattributes`、`.editorconfig`）。
- 旧版本 exe（v1.2 以前）不认识 `/__quit`，接手失败时请使用者手动关闭旧窗口。

## 8. 版本历史摘要

- v1.0.0 按老师 / 按班级 / 节数总表，单一 exe
- v1.1.0 合并英文场地课表
- v1.2.0 局域网分享、ini 设置、防火墙脚本
- v1.3.0 项目结构标准化、运行时文件夹 config/data/logs、版本号与日志、app.json
