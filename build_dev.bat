@echo off
chcp 65001 >nul
echo ============================================
echo   TalkABC - Development Build Script
echo ============================================
echo.
echo Environment: Development
echo.

echo Cleaning previous build...
if exist talkabc-dev (
    echo Found existing talkabc-dev file, deleting...
    del /f /q talkabc-dev
)
if exist talkabc-dev.exe (
    echo Found existing talkabc-dev.exe file, deleting...
    del /f /q talkabc-dev.exe
)
echo.

echo Building for Development environment...
echo.
setlocal
set CGO_ENABLED=0
set GOOS=linux
set GOARCH=amd64
go build -ldflags "-w -s" -o talkabc-dev ./cmd/server
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
echo   Output: talkabc-dev (Linux 64-bit)
echo   Environment: Development
echo ============================================
pause
