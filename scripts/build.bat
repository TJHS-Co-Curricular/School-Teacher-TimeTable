@echo off
chcp 65001 >nul
setlocal EnableExtensions
title 打包 Teacher-TimeTable.exe

rem ============================================================
rem  循人课表 - 打包
rem  输出：dist\Teacher-TimeTable\（完整的运行时文件夹）
rem     Teacher-TimeTable.exe、config\、data\、logs\、firewall-allow.bat
rem  版本号读取项目根目录的 VERSION 文件。
rem ============================================================

for %%i in ("%~dp0..") do set "ROOT=%%~fi"
set "MODULE=github.com/TJHS-Co-Curricular/School-Teacher-TimeTable"
set "EXE_NAME=Teacher-TimeTable.exe"
set "OUT=%ROOT%\dist\Teacher-TimeTable"

echo.
echo ==============================================
echo   循人课表 - 打包 %EXE_NAME%
echo ==============================================
echo.

if not exist "%ROOT%\go.mod" (
  echo [错误] 找不到 go.mod。build.bat 要放在项目的 scripts 文件夹里。
  goto :fail
)
set /p VER=<"%ROOT%\VERSION"
set "VER=%VER: =%"

rem ---------- 1. 检查 Go ----------
call :find_go
if not defined GO_OK (
  echo [提示] 这台电脑还没有安装 Go（打包需要，用 exe 的人不需要）。
  echo.
  where winget >nul 2>nul
  if errorlevel 1 goto :no_winget
  choice /c YN /n /m "要现在用 winget 自动安装 Go 吗？[Y/N] "
  if errorlevel 2 goto :no_winget
  winget install --id GoLang.Go -e --accept-source-agreements --accept-package-agreements
  call :find_go
  if not defined GO_OK (
    echo [提示] Go 已安装，但这个窗口还读不到。请关闭窗口后重新运行 build.bat。
    goto :fail
  )
)
for /f "tokens=3" %%v in ('go version') do set "GO_VER=%%v"
echo [1/5] Go %GO_VER%，版本 v%VER%

rem ---------- 2. 版本资讯 ----------
set "COMMIT="
where git >nul 2>nul && for /f %%c in ('git -C "%ROOT%" rev-parse --short HEAD 2^>nul') do set "COMMIT=%%c"
for /f %%d in ('powershell -NoProfile -Command "Get-Date -Format yyyy-MM-dd"') do set "BDATE=%%d"

rem ---------- 3. 结束正在运行的 exe（否则无法覆盖） ----------
tasklist /fi "imagename eq %EXE_NAME%" 2>nul | find /i "%EXE_NAME%" >nul
if not errorlevel 1 (
  echo [2/5] %EXE_NAME% 正在运行，先把它关闭……
  taskkill /im "%EXE_NAME%" /f >nul 2>nul
  timeout /t 1 /nobreak >nul
) else (
  echo [2/5] 没有正在运行的 %EXE_NAME%
)

rem ---------- 4. 检查代码 ----------
pushd "%ROOT%"
echo [3/5] 检查代码……
go vet ./...
if errorlevel 1 (
  popd
  echo [错误] 代码检查失败，请看上面的讯息。
  goto :fail
)

rem ---------- 5. 编译 ----------
echo [4/5] 编译中……
set "GOOS=windows"
set "GOARCH=amd64"
set "CGO_ENABLED=0"
go build -trimpath -ldflags "-s -w -X %MODULE%/internal/buildinfo.Version=%VER% -X %MODULE%/internal/buildinfo.Commit=%COMMIT% -X %MODULE%/internal/buildinfo.BuildDate=%BDATE%" -o "%OUT%\%EXE_NAME%" ./cmd/teacher-timetable
set "BUILD_ERR=%errorlevel%"
popd
if not "%BUILD_ERR%"=="0" (
  echo [错误] 编译失败，请看上面的讯息。
  goto :fail
)

rem ---------- 6. 运行时文件夹 ----------
echo [5/5] 准备运行时文件夹……
for %%d in (config data logs) do if not exist "%OUT%\%%d" mkdir "%OUT%\%%d"
copy /y "%ROOT%\scripts\firewall-allow.bat" "%OUT%\" >nul

for %%f in ("%OUT%\%EXE_NAME%") do set "SIZE=%%~zf"
set /a SIZE_MB=%SIZE% / 1048576
echo.
echo ==============================================
echo   完成！v%VER%  %COMMIT%
echo   文件夹：%OUT%
echo   大小：约 %SIZE_MB% MB
echo ==============================================
echo.
echo 把课表文件放进 %OUT%\data\，双击 %EXE_NAME% 即可使用。
echo 设置文件会在第一次运行时出现在 config\。
echo.
pause
exit /b 0

rem ============================================================
:find_go
set "GO_OK="
where go >nul 2>nul && (set "GO_OK=1" & goto :eof)
if exist "%ProgramFiles%\Go\bin\go.exe" (
  set "PATH=%ProgramFiles%\Go\bin;%PATH%"
  set "GO_OK=1"
)
goto :eof

:no_winget
echo 请手动安装 Go：https://go.dev/dl/ （下载 windows-amd64.msi），装好后重新运行 build.bat。
start "" "https://go.dev/dl/"
goto :fail

:fail
echo.
pause
exit /b 1
