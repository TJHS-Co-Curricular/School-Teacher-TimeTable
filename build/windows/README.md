# build/windows

exe 的图标、manifest 与版本资讯。

- `app.rc.in`：资源描述模板，`@VERSION@` / `@VERSION_COMMA@` 由 `scripts/build.sh` 用 `VERSION` 取代。
- 用 mingw-w64 的 `windres` 编译成 `cmd/teacher-timetable/rsrc_windows_amd64.syso`，Go 编译时会自动带入。
- Windows 的 `build.bat` 没有 windres，会直接使用仓库里已编译好的 `.syso`；改版号后请在 Linux/WSL 跑一次 `scripts/build.sh` 更新它（只影响 exe「属性 → 详细信息」里的版本号，程序内显示的版本号一律来自 `VERSION`）。
