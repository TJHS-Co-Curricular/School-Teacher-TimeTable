@echo off
chcp 65001 >nul
setlocal EnableExtensions
title 打包 Teacher-TimeTable.exe

rem ============================================================
rem  循人课表 - 一键打包 Teacher-TimeTable.exe
rem  用法：双击本文件。输出在同一个文件夹的 Teacher-TimeTable.exe
rem ============================================================

rem build.bat 放在项目根目录或 source\ 里都可以
if exist "%~dp0main.go" (
  set "SRC=%~dp0."
  for %%i in ("%~dp0..") do set "ROOT=%%~fi\"
) else (
  set "SRC=%~dp0source"
  set "ROOT=%~dp0"
)
set "OUT=%ROOT%Teacher-TimeTable.exe"
set "EXE_NAME=Teacher-TimeTable.exe"

echo.
echo ==============================================
echo   循人课表 - 打包 %EXE_NAME%
echo ==============================================
echo.

rem ---------- 1. 检查源代码 ----------
if not exist "%SRC%\main.go" (
  echo [错误] 找不到 main.go。build.bat 要放在项目根目录或 source 文件夹里，
echo        而且 source 里要有 main.go、go.mod 等 Go 源代码。
  goto :fail
)

rem ---------- 2. 检查 Go ----------
call :find_go
if not defined GO_OK (
  echo [提示] 这台电脑还没有安装 Go（打包需要，用 exe 的人不需要）。
  echo.
  where winget >nul 2>nul
  if errorlevel 1 goto :no_winget
  choice /c YN /n /m "要现在用 winget 自动安装 Go 吗？[Y/N] "
  if errorlevel 2 goto :no_winget
  echo.
  echo 正在安装 Go，请稍候……
  winget install --id GoLang.Go -e --accept-source-agreements --accept-package-agreements
  call :find_go
  if not defined GO_OK (
    echo.
    echo [提示] Go 已安装，但这个窗口还读不到。请关闭窗口后重新双击 build.bat。
    goto :fail
  )
)
for /f "tokens=3" %%v in ('go version') do set "GO_VER=%%v"
echo [1/4] Go 版本：%GO_VER%

rem ---------- 3. 如果 exe 正在运行，先结束它（否则无法覆盖） ----------
tasklist /fi "imagename eq %EXE_NAME%" 2>nul | find /i "%EXE_NAME%" >nul
if not errorlevel 1 (
  echo [2/4] %EXE_NAME% 正在运行，先把它关闭……
  taskkill /im "%EXE_NAME%" /f >nul 2>nul
  timeout /t 1 /nobreak >nul
) else (
  echo [2/4] 没有正在运行的 %EXE_NAME%
)

rem ---------- 4. 检查代码 ----------
pushd "%SRC%"
echo [3/4] 检查代码……
go vet .
if errorlevel 1 (
  popd
  echo [错误] 代码检查失败，请看上面的讯息。
  goto :fail
)

rem ---------- 5. 编译 ----------
echo [4/4] 编译中……
set "GOOS=windows"
set "GOARCH=amd64"
set "CGO_ENABLED=0"
go build -trimpath -ldflags "-s -w" -o "%OUT%" .
set "BUILD_ERR=%errorlevel%"
popd
if not "%BUILD_ERR%"=="0" (
  echo [错误] 编译失败，请看上面的讯息。
  goto :fail
)

for %%f in ("%OUT%") do set "SIZE=%%~zf"
set /a SIZE_MB=%SIZE% / 1048576
echo.
echo ==============================================
echo   完成！
echo   文件：%OUT%
echo   大小：约 %SIZE_MB% MB
echo ==============================================
echo.
echo 把 %EXE_NAME% 和「班级课表.html」放在同一个文件夹，双击即可使用。
echo.
pause
exit /b 0

rem ============================================================
:find_go
set "GO_OK="
where go >nul 2>nul && (set "GO_OK=1" & goto :eof)
rem 刚装好 Go 时 PATH 还没更新，直接找默认安装位置
if exist "%ProgramFiles%\Go\bin\go.exe" (
  set "PATH=%ProgramFiles%\Go\bin;%PATH%"
  set "GO_OK=1"
)
goto :eof

:no_winget
echo 请手动安装 Go：
echo   1. 打开 https://go.dev/dl/
echo   2. 下载 go1.xx.windows-amd64.msi 并安装
echo   3. 关闭这个窗口，重新双击 build.bat
start "" "https://go.dev/dl/"
goto :fail

:fail
echo.
pause
exit /b 1
