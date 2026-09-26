@echo off
chcp 65001 >nul
setlocal
title 允许同事访问课表（Windows 防火墙）

rem ============================================================
rem  让同一网络的同事可以打开 Teacher-TimeTable.exe 的课表
rem  会删除这个程序旧的防火墙规则（包括之前按「取消」产生的封锁规则），
rem  再新增一条「允许」规则。需要系统管理员权限，会跳出确认视窗。
rem ============================================================

net session >nul 2>&1
if errorlevel 1 (
  echo 需要系统管理员权限，正在请求……
  powershell -NoProfile -Command "Start-Process -FilePath '%~f0' -Verb RunAs"
  exit /b
)

set "EXE=%~dp0Teacher-TimeTable.exe"
if not exist "%EXE%" (
  echo [错误] 在这个文件夹找不到 Teacher-TimeTable.exe：
  echo        %~dp0
  echo        请把 firewall-allow.bat 放在 exe 旁边再运行。
  pause
  exit /b 1
)

echo 清除旧规则……
netsh advfirewall firewall delete rule name=all program="%EXE%" >nul 2>&1

echo 新增允许规则……
netsh advfirewall firewall add rule name="Teacher-TimeTable 课表" dir=in action=allow program="%EXE%" enable=yes profile=any
if errorlevel 1 (
  echo [错误] 新增规则失败。
  pause
  exit /b 1
)

echo.
echo 完成！重新打开 Teacher-TimeTable.exe，同事就可以用窗口里的「同事网址」打开课表。
echo.
pause
