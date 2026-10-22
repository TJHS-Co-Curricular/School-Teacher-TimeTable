# HANDOFF.md — 循人课表（TJHS School Teacher TimeTable）交接文档

> 一次性交接：从 Claude app（Cowork）转到 Claude Code。
> 编写日期：2026-10-17；对话期间：2026-09-24 至 2026-09-26。
> 长期规范请看 `CLAUDE.md`；本文件读完、事项处理完后可以删除或移到 `docs/`。
> 标注【待确认】的内容是我无法核实或前后不一致的地方，请不要当成事实。

---

## 0. 盘点：我掌握的全部内容

### 0.1 资料来源

| 来源                 | 内容                                                                                                                               |
| -------------------- | ---------------------------------------------------------------------------------------------------------------------------------- |
| 本项目对话           | 2026-09-24 到 2026-09-26 共约 15 轮需求（见 §0.2）                                                                                 |
| 使用者电脑上的文件夹 | `D:\00-Data\Documents\Github\TJHS-School-Teacher-TimeTable`（通过 Cowork 桥接读写；**现在已无法连接**，最后一次读到是 2026-09-26） |
| eSchool 导出文件     | `班级课表.html`（2,372,570 bytes，列印日期 24-09-2026 14:40）、`场地课表_English.html`（813,815 bytes，列印日期 25-09-2026 12:26） |
| GitHub               | `https://github.com/TJHS-Co-Curricular/School-Teacher-TimeTable`（使用者在 config.js 里填的；我**没有权限读取**这个仓库）          |
| Claude Project 文档  | `claude/CLAUDE.md`（2026-09-26 我写的第一版规范）                                                                                  |

### 0.2 需求时间线（使用者原话的意思 → 我的处理）

| #   | 日期  | 使用者需求                                                                           | 结果                                                                                                                       |
| --- | ----- | ------------------------------------------------------------------------------------ | -------------------------------------------------------------------------------------------------------------------------- |
| 1   | 09-24 | 根据「班级课表」做页面：按老师（整合星期一至五节数）、按班级（照旧），设计跟原版一样 | 单一 HTML，沿用 eSchool 列印样式                                                                                           |
| 2   | 09-24 | 开始页面改成按钮，不要下拉列表                                                       | 按钮网格首页                                                                                                               |
| 3   | 09-24 | 舍弃原设计，重新设计                                                                 | 现在的浅灰绿 + 深绿主色、科目配色设计                                                                                      |
| 4   | 09-24 | 每次启动都读取「班级课表」，打包成可携带 .exe，不需要 Python                         | Go 程序：本机网页服务 + 浏览器内解析                                                                                       |
| 5   | 09-24 | 页面下面加 footer：学校名与版权 + GitHub 链接（当时还没建仓库，先留占位）            | 页脚                                                                                                                       |
| 6   | 09-24 | build.bat 报错 `'go' is not recognized`                                              | 改进 build.bat：检查 Go、失败不再显示「完成」；输出名改 `Teacher-TimeTable.exe`（使用者改的名字）                          |
| 7   | 09-24 | GitHub 链接；把 css/js/html 标准化                                                   | 拆成 index.html + css + 5 个 js，Prettier 格式化                                                                           |
| 8   | 09-24 | 帮我写个 build.bat                                                                   | 自动安装 Go（winget）、关闭运行中的 exe、vet、编译                                                                         |
| 9   | 09-24 | build.bat 报「找不到 source\main.go」                                                | 原因：使用者误删了 Go 源码（我上一条说明写得容易误解）＋把 build.bat 移进 source\；已补回源码、让 build.bat 两个位置都能用 |
| 10  | 09-25 | 加了「场地课表\_English」，英文系统和教务处分开，把英文 merge 进来                   | 英文合并（v1.1）                                                                                                           |
| 11  | 09-25 | 同事打我的 IP 无法访问                                                               | 局域网分享、ini 设置、firewall-allow.bat（v1.2）                                                                           |
| 12  | 09-25 | `Teacher-TimeTable.ini` 没有功能，重新校对代码                                       | 根本原因是使用者电脑上的 main.go 是旧版（写入失败，见 §4.1）；顺便修了 6 个设置相关 bug                                    |
| 13  | 09-26 | 项目标准化：代码结构、运行时文件夹统一、版本号与日志；ini/json 放 config/            | 重构为 v1.3.0（现在的结构）                                                                                                |
| 14  | 09-26 | 写成 prompt，方便不同电脑的 Claude 同步                                              | 第一版 CLAUDE.md（仓库根目录 + Claude Project）                                                                            |
| 15  | 10-17 | 转到 Claude Code，生成交接包                                                         | 本文件 + 新版 CLAUDE.md                                                                                                    |

### 0.3 我最后确认过的状态（2026-09-26）

- v1.3.0 所有文件已写进使用者的文件夹，并**逐一读回比对一致**（main.go、server.go、settings.go、logx.go、paths.go、source.go、4 个 js、index.html、settings.ini、build.bat、README、.gitignore、go.mod、exe、dist 的 settings.ini）。
- 使用者原本的设置 `lan = true`、`port = 5001` 已转进 `dist/Teacher-TimeTable/config/settings.ini`。
- 两份课表已复制到 `dist/Teacher-TimeTable/data/`。
- 当时 git 最新 commit 时间约为 2026-09-25 12:41（早于 v1.3.0），**v1.3.0 是否已 commit / push【待确认】**。

---

## 1. 项目目标

把 eSchool 导出的「班级课表」（以及英文系统的「场地课表\_English」）自动整合成**按老师 / 按班级 / 老师节数总表**的网页，打包成一个不需安装任何东西的 Windows exe，并能在校内局域网分享给同事查看。

---

## 2. 当前进度

### 已完成

- [x] 解析 eSchool 班级课表（GBK HTML）：59 个班、每天 11 节、星期一至六
- [x] **按老师**：个人课表、每日节数、每周总节数、任教班级/科目统计；合班同时段显示在同一格
- [x] **按班级**：课表、科目统计、班导师、可互相点击跳转
- [x] **老师节数总表**：可排序、可搜索，显示老师人数、平均节数、最多节数
- [x] 首页为按钮卡片（不用下拉），老师卡片有星期一至五小柱状图
- [x] 重新设计的界面：浅/深色模式、科目固定配色、手机版按天列表、A4 横向打印
- [x] 页脚：© 年份 学校名 · 版权所有 · 版本号 · GitHub 按钮
- [x] 合并英文系统「场地课表\_English」：39 位英文老师以真名加入（老师总数 208），117 个英文组全部对应到班级，E1–E12 代号移除，云惟祥两边节数合并
- [x] Go 单一 exe：本机网页服务，每次请求重新读课表（替换文件按 F5 即更新）
- [x] 局域网分享（`lan = true`），显示同事网址与网卡名称；`firewall-allow.bat`
- [x] 单一实例：设置相同则直接打开；不同则通过 `/__quit` 接手旧窗口
- [x] 设置文件解析：`[section]`、行尾注解、UTF-8 BOM、UTF-16、写错的行会列出
- [x] v1.3.0 标准结构：`cmd/ internal/ web/ config/ scripts/ build/`
- [x] 运行时文件夹 `config/ data/ logs/`；旧版 `Teacher-TimeTable.ini` 自动迁移为 `config/settings.ini`（旧文件改 `.bak`）
- [x] `config/app.json` 控制网页显示，改完按 F5 生效
- [x] 版本号（`VERSION` 唯一来源）注入 exe、窗口标题、日志、页脚；exe 属性版本资讯
- [x] 每日日志 `logs/YYYY-MM-DD.log`，自动清理；记录同事连线
- [x] 打包脚本：`build.bat`（Windows，可自动装 Go）、`scripts/build.sh`（Linux/WSL，含 windres）
- [x] README、CHANGELOG、CLAUDE.md

### 进行中 / 未验证

- [ ] **v1.3.0 在 Windows 实机运行**：只在 Linux 测试过（Go 交叉编译）。以下未在 Windows 验证【待确认】：
  - 命令窗口 UTF-8 中文与窗口标题（`SetConsoleTitleW`）
  - 关闭窗口时是否写入「窗口已关闭，程序结束」（Go 把 CTRL_CLOSE_EVENT 映射为 SIGTERM，理论上可以）
  - `scripts/build.bat`（`set /p VER=<VERSION`、PowerShell 取日期、`git rev-parse`）
  - `firewall-allow.bat`（自我提权、删除封锁规则）
  - 设置变更后自动接手旧窗口
- [ ] **同事局域网访问是否已成功**：使用者没有回报结果【待确认】
- [ ] **旧结构清理**：请使用者手动删除（我没有删除权限），是否已删【待确认】：
  - `source/` 整个文件夹（v1.2 的源码，含一个 50,213 bytes 的旧 `source/index.html`，内容来源不明【待确认】）
  - 根目录 `Teacher-TimeTable.exe`、`firewall-allow.bat`、`Teacher-TimeTable.ini`
  - 根目录 `班级课表.html`、`场地课表_English.html` 与两个 `_files` 文件夹（已复制到 `dist/.../data/`；在 `.gitignore` 中，留着也不会上传）
  - 早期的 `教师与班级课表.html`（v1 单页快照，可能已被使用者删除【待确认】）
- [ ] **v1.3.0 commit、打 tag、GitHub Release**【待确认】

### 未开始（使用者没有提出，以下为我的建议）

- [ ] 自动化测试（Go 单元测试、JS 解析测试）
- [ ] GitHub Actions 自动打包
- [ ] 场地（课室）视图、星期六课程显示于老师视图
- [ ] 两个系统老师名字的模糊比对

---

## 3. 关键决定

| 决定                 | 选择                                                                                                      | 没选的方案与原因                                                                             |
| -------------------- | --------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------- |
| exe 技术             | **Go**（单一静态 exe，Linux 可交叉编译）                                                                  | Python + PyInstaller：使用者明确要求不需要 Python；体积大、杀毒误报多。Electron：太大        |
| 课表在哪里解析       | **浏览器 JS**（DOMParser + `TextDecoder('gb18030')`）                                                     | Go 解析：标准库没有 GBK 解码，要第三方库；而且 JS 解析让「直接双击 index.html 选文件」也能用 |
| 为什么要本机网页服务 | 浏览器在 `file://` 下不能 fetch 本地文件                                                                  | 每次手动选文件：不符合「每次启动自动读取」                                                   |
| 课表读取时机         | **每次请求都重新读文件**，不缓存                                                                          | 缓存：替换文件后要重开程序                                                                   |
| 网页架构             | 原生 JS + `window.TT` 命名空间 + IIFE，按顺序载入                                                         | ES modules：`file://` 下被 CORS 挡；框架/打包工具：不需要，增加维护成本                      |
| 字体                 | 系统字体（PingFang / 微软雅黑 / Noto Sans SC）                                                            | Google Fonts：要连外网，校内可能离线                                                         |
| 路由                 | `location.hash`（`#teacher=名字` 等）                                                                     | 浏览器「上一页」可返回、可分享网址                                                           |
| 节数算法             | **只算星期一至五，按时段计**（合班只算 1 节）；排除早自习、班导师时间、晨读、联课活动、公共选修、共同备课 | 使用者需求是「星期一到星期五老师上课节数」；排除项是格子里本来就没有老师名字                 |
| 英文组 → 班级        | **英文时段集合完全相等**才对应                                                                            | 子集比对：高一（6 节）会误配到初一（8 节）的组；完全相等 117/117 全部唯一对应                |
| E1–E12               | 有英文文件时移除，没有时保留为「分组 / 代号」                                                             | 它们不是真人（同时段由英文系统的多位老师分组上课）                                           |
| 局域网               | 默认 `lan = true`，监听 0.0.0.0、固定端口                                                                 | 使用者的需求就是同事可访问；`lan = false` 保留给只想自己用的情况                             |
| lan 模式不自动结束   | 一直运行到关闭窗口                                                                                        | 自动结束会让同事突然连不上                                                                   |
| 单一实例             | `/__ping` 辨认 + `/__cfg` 比对设置 + `/__quit`（只接受 127.0.0.1）                                        | 不比对：改了 ini 重开会被旧程序吃掉（v1.2 的真实 bug）                                       |
| 设定档格式           | 程序用 **INI**（可写中文注解，对非开发者友好）；网页用 **JSON**（直接对应 JS 的 `TT.config`）             | 全部 JSON：不能写注解；全部 INI：网页要再写一个解析器                                        |
| 设定档位置           | `config/`（使用者要求）                                                                                   | exe 旁边：v1.2 的做法，文件夹会乱                                                            |
| 运行时文件夹         | `dist/Teacher-TimeTable/` 内含 `config/ data/ logs/`，整个可复制                                          | exe 放仓库根目录：配置、数据、日志会混进仓库                                                 |
| 版本号               | 根目录 `VERSION` 唯一来源，`-ldflags -X` 注入                                                             | 写死在代码里：容易忘了改                                                                     |
| 默认设定档           | `config/` 里的文件用 `embed.go` 打包进 exe，第一次运行写出                                                | 由打包脚本复制：exe 单独复制时会缺设置                                                       |
| exe 图标             | Linux 用 mingw `windres`（自制「直通」预处理器 `windres-pp.sh`，因为没装 gcc）生成 `.syso` 并提交到仓库   | Windows 没有 windres，所以 `build.bat` 直接用仓库里的 `.syso`                                |
| 隐私                 | 课表、`dist/`、`data/`、`logs/` 全部 `.gitignore`                                                         | 课表含老师名字，不能上传                                                                     |
| exe 进 git？         | 不进，用 GitHub Releases 发布                                                                             | exe 6.9 MB，每次改都会膨胀仓库                                                               |

---

## 4. 已知问题、限制与踩过的坑

### 4.1 踩过的坑（历史）

1. **Cowork 桥接写入后内容是旧版本**：v1.2 时 `source/main.go` 写入后，设备上的内容竟是上一版，导致使用者打包出来的 exe 没有 ini 功能。原因不明。之后每次写入都读回比对。**转到 Claude Code 后直接在本机改文件，这个问题应该不会再出现。**
2. **说明写得模糊导致误删**：我写「删掉 source/index.html 和 source/app.manifest 以外的旧文件」，使用者理解成「只保留这两个」，把 Go 源码全删了。之后列删除清单要写成「要删的就是这几个」。
3. **从大文件截取代码时少了最后一行**：两次把单页 HTML 拆分时漏掉了 `render();`，页面停在「正在读取」。拆分后一定要实际打开测试。
4. **`pkill -f` 杀到自己的 shell**：测试时用 `pkill -f "rt/tt"` 会匹配到包含这个字串的命令本身。用 `pgrep -x`。
5. **行尾注解**：README 的 ini 例子写了 `lan = true ; 说明`，旧解析器会读成 false。已修。
6. **两个系统的休息栏位置不同**：班级课表 `1 2 | 3 4 | 5 6* | 7 8 9 | 10 11`，英文系统 `1 2 | 3 4 | 5 6* 7 | 8 9 | 10 11`。时间相同，所以按节次索引对齐、不要按栏位。
7. **windres 需要 C 预处理器**：没有 gcc 时会失败，`.rc` 里也不能写 `//` 注解（直通预处理器不会去掉）。

### 4.2 现有限制 / 可能的 bug

- **未在 Windows 实机测试 v1.3.0**（见 §2）。
- `build.bat` 不会更新 exe 属性里的版本号（用仓库里的 `.syso`），需在 Linux/WSL 跑 `scripts/build.sh`。现在 `.syso` 里是 1.3.0。
- v1.2 及更早的 exe 不认识 `/__quit`，新版接手会失败 → 提示手动关闭旧窗口。
- lan 模式下端口被别的程序占用会直接报错退出（故意的：同事需要固定端口）。
- 接手旧窗口时，新旧两个进程会短暂写同一个日志文件（无害）。
- 老师名字跨系统比对是**完全相同的字串**；若两个系统写法不同（大小写、空格、繁简），会变成两位老师【待确认是否有这种情况】。
- 英文组对应不到任何班级时（时段不完全相同），目前**不会提示**，该组的「班级」会是空的。
- `场地课表_English` 每页的「课时总数」没有使用，也没有用来核对。
- 老师视图只显示星期一至五；星期六只有联课活动，目前视为不需要【待确认】。
- `COMM1` / `COoMm1`（两种拼法都出现在 eSchool 资料里）、`Maker7`、`科实A` 仍当作「分组 / 代号」显示。
- 直接双击 `web/index.html`（file:// 模式）每次都要重新选文件；页脚不显示版本号。
- 页面 CSS 用了 `color-mix()`，需要 Chrome/Edge 111+ 等较新的浏览器。
- exe 没有数字签名，Windows SmartScreen 第一次会警告（点「更多信息 → 仍要运行」）。
- 学校 Wi-Fi 若有 AP isolation，开了防火墙同事也连不上。
- **没有自动化测试**。只做过手动 / Playwright 截图测试。

### 4.3 数据观察（2026-09 的课表）

- 班级：59 个（初一 A–J、初二 A–L、初三 A–J、高一 商一–七 / 理一–三、高二 商一–六 / 理一–三、高三 商一–五 / 理一–三）
- 合并前 170 位老师 + 28 个代号；合并后 208 位老师
- 英文场地：21 个（E001–E006、E102、E201–E207、E301–E307）；英文组 117 个（8 节或 6 节）
- 英文课型：GR、WR、SP/RD、WB、RD/SPM、RD/SUEC
- 当时没有老师在同一时段被排进两个班（核对过）

---

## 5. 下一步（按优先级）

1. **在 Windows 实测 v1.3.0**（使用者）
   - 双击 `dist\Teacher-TimeTable\Teacher-TimeTable.exe`：窗口标题有版本号、中文正常、显示 `port = 5001`、`logs\` 产生当天日志
   - 同事用「同事网址」打开；必要时以系统管理员身份跑 `firewall-allow.bat`
   - 改 `config\settings.ini` 后重开，确认旧窗口被接手
   - 双击 `build.bat` 打包一次
2. **清理旧结构**（使用者手动删）：`source\`、根目录的 `Teacher-TimeTable.exe`、`firewall-allow.bat`、`Teacher-TimeTable.ini`；先确认 `source\index.html`（50,213 bytes）是不是有自己改过的内容。
3. **git**：`git status` 确认 v1.3.0 文件都在 → commit → `git tag v1.3.0` → push → GitHub Release 上传 `dist\Teacher-TimeTable` 的压缩档（压缩前把 `data\`、`logs\` 清空，避免泄漏老师名字）。
4. **加测试**（建议）：
   - Go：`internal/settings`（section、行尾注解、BOM、UTF-16、错误行、旧版 Render）、`internal/source`（data/ 优先、标记比对）、`internal/logx`（清理）
   - JS：把 `parser.js` + 合并逻辑做成可在 Node 跑的测试；**测试用的课表要匿名化**（不可提交真名）
5. **GitHub Actions**（建议）：push tag 时跑 vet/test、交叉编译 Windows exe、附到 Release。
6. **英文合并的防呆**（建议）：对应不到班级的英文组、两边名字相近但不同的老师，在页面或日志提示。
7. **之后可考虑**（建议）：课室（场地）视图、星期六显示、导出 Excel、exe 数字签名。

---

## 6. 文件变动

本项目从零开始建立了所有代码。以下是 **v1.3.0 最终版本** 的全部文件（交接包里都有完整内容）：

| 文件                                                                      | 内容                                                                                               |
| ------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------- |
| `go.mod`                                                                  | module `github.com/TJHS-Co-Curricular/School-Teacher-TimeTable`，go 1.22，无第三方依赖             |
| `VERSION`                                                                 | `1.3.0`                                                                                            |
| `CHANGELOG.md`                                                            | v1.0.0–v1.3.0 更新记录                                                                             |
| `README.md`                                                               | 使用、设定档、局域网、日志、英文合并、项目结构、开发流程                                           |
| `CLAUDE.md`                                                               | 长期项目规范（本次重写）                                                                           |
| `HANDOFF.md`                                                              | 本文件                                                                                             |
| `.gitignore`                                                              | dist/、课表、data/、logs/、_.exe、_.bak、旧 ini、windres 中间档                                    |
| `.gitattributes` / `.editorconfig` / `.prettierrc.json`                   | 换行（.bat 用 CRLF）、缩进、Prettier（120 宽、单引号、LF）                                         |
| `build.bat`                                                               | 快捷方式 → `scripts\build.bat`                                                                     |
| `cmd/teacher-timetable/main.go`                                           | 入口：运行时文件夹 → 旧设置迁移 → 读设置/开日志 → 找课表 → 单一实例 → 监听 → 打开浏览器 → 结束条件 |
| `cmd/teacher-timetable/rsrc_windows_amd64.syso`                           | 二进制：exe 图标、manifest、版本 1.3.0                                                             |
| `internal/buildinfo/buildinfo.go`                                         | AppName、ExeName、AppID、Version/Commit/BuildDate                                                  |
| `internal/logx/logx.go`                                                   | 窗口 + 每日日志、等级、清理                                                                        |
| `internal/netutil/netutil.go`                                             | 局域网 IPv4 + 网卡名、ClientIP、IsLoopback                                                         |
| `internal/paths/paths.go`                                                 | Layout（Root/Config/Data/Logs）、Ensure、旧 ini 位置、`TEACHER_TIMETABLE_HOME`                     |
| `internal/platform/*.go`                                                  | 命令窗口 UTF-8 与标题（Windows）、打开浏览器                                                       |
| `internal/server/server.go`                                               | 路由、课表读取（记录同事 IP）、`/config.json`、单一实例辅助函数                                    |
| `internal/settings/settings.go`                                           | settings.ini 解析、Render（旧版转新格式）、LoadApp（app.json）                                     |
| `internal/source/source.go`                                               | Class / English 两种来源、data/ → 根目录查找、内容标记比对                                         |
| `config/embed.go` / `settings.ini` / `app.json`                           | 默认设定档（打包进 exe）                                                                           |
| `web/embed.go`                                                            | 打包网页                                                                                           |
| `web/index.html`                                                          | 页面骨架（顶栏、main、页脚）                                                                       |
| `web/css/style.css`                                                       | 全部样式（设计变量、浅/深色、科目色、响应式、打印）                                                |
| `web/js/config.js`                                                        | 网页默认设置（会被 /config.json 覆盖）                                                             |
| `web/js/utils.js`                                                         | esc、zh 排序、$、to24                                                                              |
| `web/js/parser.js`                                                        | 解析 eSchool HTML（班级 / 场地两种）                                                               |
| `web/js/app.js`                                                           | 整合老师课表、英文合并、三个视图、路由                                                             |
| `web/js/main.js`                                                          | 读取 /config.json、自动载入课表、文件选择、心跳                                                    |
| `web/favicon.ico`、`build/windows/app.ico`                                | 二进制：日历图标                                                                                   |
| `build/windows/app.manifest` / `app.rc.in` / `README.md`                  | exe 资源与说明                                                                                     |
| `scripts/build.bat` / `build.sh` / `windres-pp.sh` / `firewall-allow.bat` | 打包与防火墙脚本                                                                                   |

使用者电脑上**应删除的旧文件**（v1.2 以前，已被上面取代）：`source/`（全部）、根目录 `Teacher-TimeTable.exe`、`firewall-allow.bat`、`Teacher-TimeTable.ini`。

使用者电脑上 `dist/Teacher-TimeTable/` 由我生成的内容（不进 git）：`Teacher-TimeTable.exe`（v1.3.0）、`firewall-allow.bat`、`config/settings.ini`（lan=true, port=5001）、`config/app.json`、`data/班级课表.html`、`data/场地课表_English.html`、`data/请把课表放在这里.txt`。

---

## 7. 待确认事项汇总

1. v1.3.0 是否已 commit / push 到 GitHub？是否打了 tag？
2. v1.3.0 在 Windows 实机运行是否正常（中文窗口、标题、日志、接手、关闭时写日志、build.bat、firewall-allow.bat）？
3. 同事通过局域网访问是否已经成功？
4. 旧结构（`source/`、根目录 exe / bat / ini）是否已删除？
5. 设备上 `source/index.html`（50,213 bytes，2026-09-24 修改）是什么内容？我写过的单页版本约 40 KB，大小对不上，可能是使用者自己的版本。
6. 早期的 `教师与班级课表.html` 是否还在、还需要吗？
7. 使用者电脑上的 Go 版本？（`go.mod` 要求 1.22+；使用者曾成功用 build.bat 打包，所以已安装 Go，版本未知）
8. 两个系统是否有同一位老师名字写法不同的情况？（目前只发现「云惟祥」完全相同）
9. 老师视图是否需要显示星期六（目前只有联课活动）？
10. `E1`–`E12` 我判断为英文代号（数据比对 117/117 支持），但没有向使用者或教务处确认过正式定义。
11. 英文组代号的含义（例如 B/H/P 开头、数字）未确认；目前只当作代号显示，不影响功能。
12. 课表中 `6*` 的特殊安排以 `.print-tt__rmk-box` 的内容为准（J1–J3 先午餐、S1–S3 先上课），J/S 是否代表初中/高中未确认。
13. GitHub 仓库名是 `School-Teacher-TimeTable`，本地文件夹名是 `TJHS-School-Teacher-TimeTable`，两者不同是否有意？

---

## 8. 2026-10-17 Claude Code 核对结果（取代上面部分【待确认】）

| #   | 结论                                                                                                                                                       |
| --- | ---------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 1   | v1.3.0 已 commit 并 push（69ef8f4 = `origin/main`），但**没有 `v1.3.0` tag**；旧的 `Release` tag 指向 1.0 commit，挂着名为「V1.1」的 GitHub Release        |
| 2   | 已在 Windows 实机运行（日志 09-26、09-28、10-13）：日志、关闭窗口写日志、「设置相同直接打开」、`/__quit` 接手都正常；build.bat / firewall-allow.bat 未验证 |
| 3   | 仍未确认：09-28 日志里的「同事 100.90.128.8」是本机 Pangolin 网卡自己的 IP                                                                                 |
| 4/6 | 旧结构已清理：`source/` 在 69ef8f4 删除，根目录没有旧的 exe / ini / bat / 课表 / `教师与班级课表.html`                                                     |
| 7   | 这台电脑原本没有 Go；2026-10-17 已安装 Go 1.27.2（`C:\Program Files\Go`）                                                                                  |
| 13  | 本地文件夹实际叫 `School-Teacher-TimeTable`，与仓库同名                                                                                                    |
| —   | 使用者实际端口 **5001**：`dist/.../config/settings.ini` 原本是 17380，已改为 5001                                                                          |
| —   | 英文课表实际文件名是 `场地课表.html`；使用者决定程序和文档都接受这个名字                                                                                   |
| —   | 2026-09-26 的班级课表有 106 页（59 班 + 47 个教学分组页），使用者要求照常显示；两页「力学」已自动编号                                                      |
| —   | 用 2026-09-26 的资料：老师 212 位（华文名 175、英文/马来文名 37）、分组/代号 16 个、英文组 118 个全部对应                                                  |
