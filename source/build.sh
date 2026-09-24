#!/bin/sh
# 重新打包 Teacher-TimeTable.exe（Linux / WSL / macOS，需要 Go 1.22+）
# 若装有 mingw-w64 的 windres，会一并更新 exe 图标与版本资讯；否则沿用现成的 rsrc_windows_amd64.syso。
set -e
cd "$(dirname "$0")"
if command -v x86_64-w64-mingw32-windres >/dev/null 2>&1; then
  x86_64-w64-mingw32-windres --preprocessor="$PWD/pp.sh" -c 65001 -O coff -i app.rc -o rsrc_windows_amd64.syso
fi
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -ldflags "-s -w" -o ../Teacher-TimeTable.exe .
echo "完成：../Teacher-TimeTable.exe"
