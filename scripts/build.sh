#!/bin/sh
# 打包 Teacher-TimeTable.exe（Linux / WSL / macOS，需要 Go 1.22+）
# 输出：dist/Teacher-TimeTable/（完整的运行时文件夹）
# 装有 mingw-w64 的 windres 时，会依 VERSION 更新 exe 的图标与版本资讯。
set -e
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"
MODULE="github.com/TJHS-Co-Curricular/School-Teacher-TimeTable"
VER="$(tr -d ' \r\n' < VERSION)"
COMMIT="$(git rev-parse --short HEAD 2>/dev/null || true)"
DATE="$(date +%Y-%m-%d)"
OUT="dist/Teacher-TimeTable"

if command -v x86_64-w64-mingw32-windres >/dev/null 2>&1; then
  VC="$(echo "$VER" | tr '.' ',' ),0"
  sed -e "s/@VERSION@/$VER/g" -e "s/@VERSION_COMMA@/$VC/g" build/windows/app.rc.in > build/windows/app.rc
  (cd build/windows && x86_64-w64-mingw32-windres --preprocessor="$ROOT/scripts/windres-pp.sh" -c 65001 -O coff \
    -i app.rc -o "$ROOT/cmd/teacher-timetable/rsrc_windows_amd64.syso")
  rm -f build/windows/app.rc
fi

go vet ./...
mkdir -p "$OUT/config" "$OUT/data" "$OUT/logs"
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -trimpath \
  -ldflags "-s -w -X $MODULE/internal/buildinfo.Version=$VER -X $MODULE/internal/buildinfo.Commit=$COMMIT -X $MODULE/internal/buildinfo.BuildDate=$DATE" \
  -o "$OUT/Teacher-TimeTable.exe" ./cmd/teacher-timetable
cp scripts/firewall-allow.bat "$OUT/"
echo "完成：$OUT/Teacher-TimeTable.exe（v$VER）"
