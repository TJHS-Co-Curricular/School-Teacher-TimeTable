# 更新记录

版本号格式：`主版本.次版本.修订`（[语义化版本](https://semver.org/lang/zh-CN/)）。
版本号只写在根目录的 `VERSION`，打包时自动带入 exe、窗口标题、日志与网页页脚。

## v1.3.0 — 2026-09-26

- 项目结构标准化：`cmd/`、`internal/`、`web/`、`config/`、`scripts/`、`build/`
- 运行时文件夹统一：exe 旁边固定为 `config/`（设定档）、`data/`（课表）、`logs/`（日志）
- 设定档移到 `config/`：`settings.ini`（程序）＋ `app.json`（网页显示，改了按 F5 就生效）
- 旧版 `Teacher-TimeTable.ini` 自动搬到 `config/settings.ini`（旧文件改名 `.bak`）
- 新增日志：`logs/YYYY-MM-DD.log`，记录启动、设置、课表读取、同事连线；自动删除旧日志
- 新增版本号：窗口标题、日志、网页页脚都会显示
- 新增设置：`open_browser`、`[log] keep_days`、`[log] level`
- 打包输出改到 `dist/Teacher-TimeTable/`（完整的运行时文件夹）

## v1.2.0 — 2026-09-25

- 局域网分享：同事可用 `http://你的IP:17380/` 打开课表
- `Teacher-TimeTable.ini` 设置（lan、port）；改设置后重新打开会自动接手旧窗口
- `firewall-allow.bat`：清除封锁规则并允许防火墙

## v1.1.0 — 2026-09-25

- 合并英文系统的「场地课表\_English」：英文老师以真实名字显示，班级页列出英文分组

## v1.0.0 — 2026-09-24

- 第一版：按老师 / 按班级 / 老师节数总表，由「班级课表」自动整合
- 打包成单一 exe，每次打开重新读取课表
