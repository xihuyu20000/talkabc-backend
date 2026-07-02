@echo off
chcp 65001 >nul
echo ============================================
echo   TalkABC - Linux 64-bit Cross Build Script
echo ============================================
echo.
echo Cleaning previous build...
if exist talkabc del /f talkabc
if exist talkabc.exe del /f talkabc.exe
echo.
echo Building with GOOS=linux GOARCH=amd64...
echo.
setlocal
set CGO_ENABLED=0
set GOOS=linux
set GOARCH=amd64
go build -ldflags "-w -s" -o talkabc ./cmd/server
endlocal
if %errorlevel% neq 0 (
    echo.
    echo ============================================
    echo   Build failed!
    echo ============================================
    pause
    exit /b 1
)
echo.
echo ============================================
echo   Build succeeded!
echo   Output: talkabc (Linux 64-bit)
echo ============================================
pause