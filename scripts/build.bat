@echo off
setlocal EnableExtensions
title Build Teacher-TimeTable.exe

rem ============================================================
rem  TJHS Teacher TimeTable - build script
rem  Output: dist\Teacher-TimeTable\ (complete runtime folder)
rem     Teacher-TimeTable.exe, config\, data\, logs\, firewall-allow.bat
rem  Version is read from the VERSION file in the project root.
rem  Keep this file pure ASCII + CRLF (cmd misreads UTF-8 lines).
rem ============================================================

for %%i in ("%~dp0..") do set "ROOT=%%~fi"
set "MODULE=github.com/TJHS-Co-Curricular/School-Teacher-TimeTable"
set "EXE_NAME=Teacher-TimeTable.exe"
set "OUT=%ROOT%\dist\Teacher-TimeTable"

echo.
echo ==============================================
echo   TJHS Teacher TimeTable - build %EXE_NAME%
echo ==============================================
echo.

if not exist "%ROOT%\go.mod" (
  echo [ERROR] go.mod not found. build.bat must stay in the project's scripts folder.
  goto :fail
)
set /p VER=<"%ROOT%\VERSION"
set "VER=%VER: =%"

rem ---------- 1. Check Go ----------
call :find_go
if not defined GO_OK (
  echo [INFO] Go is not installed on this computer. It is only needed to build; users of the exe do not need it.
  echo.
  where winget >nul 2>nul
  if errorlevel 1 goto :no_winget
  choice /c YN /n /m "Install Go now with winget? [Y/N] "
  if errorlevel 2 goto :no_winget
  winget install --id GoLang.Go -e --accept-source-agreements --accept-package-agreements
  call :find_go
  if not defined GO_OK (
    echo [INFO] Go is installed, but this window cannot see it yet. Close this window and run build.bat again.
    goto :fail
  )
)
for /f "tokens=3" %%v in ('go version') do set "GO_VER=%%v"
echo [1/5] Go %GO_VER%, version v%VER%

rem ---------- 2. Version info ----------
set "COMMIT="
where git >nul 2>nul && for /f %%c in ('git -C "%ROOT%" rev-parse --short HEAD 2^>nul') do set "COMMIT=%%c"
for /f %%d in ('powershell -NoProfile -Command "Get-Date -Format yyyy-MM-dd"') do set "BDATE=%%d"

rem ---------- 3. Close the running exe (otherwise it cannot be overwritten) ----------
tasklist /fi "imagename eq %EXE_NAME%" 2>nul | find /i "%EXE_NAME%" >nul
if not errorlevel 1 (
  echo [2/5] %EXE_NAME% is running, closing it...
  taskkill /im "%EXE_NAME%" /f >nul 2>nul
  timeout /t 1 /nobreak >nul
) else (
  echo [2/5] %EXE_NAME% is not running
)

rem ---------- 4. Check code ----------
pushd "%ROOT%"
echo [3/5] Checking code...
go vet ./...
if errorlevel 1 (
  popd
  echo [ERROR] Code check failed. See the messages above.
  goto :fail
)

rem ---------- 5. Compile ----------
echo [4/5] Compiling...
set "GOOS=windows"
set "GOARCH=amd64"
set "CGO_ENABLED=0"
go build -trimpath -ldflags "-s -w -X %MODULE%/internal/buildinfo.Version=%VER% -X %MODULE%/internal/buildinfo.Commit=%COMMIT% -X %MODULE%/internal/buildinfo.BuildDate=%BDATE%" -o "%OUT%\%EXE_NAME%" ./cmd/teacher-timetable
set "BUILD_ERR=%errorlevel%"
popd
if not "%BUILD_ERR%"=="0" (
  echo [ERROR] Compile failed. See the messages above.
  goto :fail
)

rem ---------- 6. Runtime folder ----------
echo [5/5] Preparing runtime folder...
for %%d in (config data logs) do if not exist "%OUT%\%%d" mkdir "%OUT%\%%d"
copy /y "%ROOT%\scripts\firewall-allow.bat" "%OUT%\" >nul

for %%f in ("%OUT%\%EXE_NAME%") do set "SIZE=%%~zf"
set /a SIZE_MB=%SIZE% / 1048576
echo.
echo ==============================================
echo   Done! v%VER%  %COMMIT%
echo   Folder: %OUT%
echo   Size:   about %SIZE_MB% MB
echo ==============================================
echo.
echo Put the timetable files in %OUT%\data\ and double-click %EXE_NAME%.
echo The settings file appears in config\ after the first run.
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
echo Please install Go manually: https://go.dev/dl/ (download windows-amd64.msi), then run build.bat again.
start "" "https://go.dev/dl/"
goto :fail

:fail
echo.
pause
exit /b 1
