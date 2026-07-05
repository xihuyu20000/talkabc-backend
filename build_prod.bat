@echo off
chcp 65001 >nul
echo ============================================
echo   TalkABC - Production Build Script
echo ============================================
echo.
echo Environment: Production
echo.

echo Cleaning previous build...
if exist talkabc (
    echo Found existing talkabc file, deleting...
    del /f /q talkabc
)
if exist talkabc.exe (
    echo Found existing talkabc.exe file, deleting...
    del /f /q talkabc.exe
)
echo.

echo Building for Production environment...
echo.
setlocal
set CGO_ENABLED=0
set GOOS=linux
set GOARCH=amd64

for /f "tokens=*" %%a in ('git rev-parse --short HEAD 2^>nul') do set GIT_COMMIT=%%a
if "%GIT_COMMIT%"=="" set GIT_COMMIT=unknown

for /f "tokens=1-4 delims=/ " %%a in ('date /t') do set DATE_STR=%%c%%b%%a
for /f "tokens=1-2 delims=:" %%a in ('time /t') do set TIME_STR=%%a%%b
set BUILD_TIME=%DATE_STR%-%TIME_STR%

echo Git Commit: %GIT_COMMIT%
echo Build Time: %BUILD_TIME%
echo.

go build -ldflags "-w -s -X main.Version=1.0.0 -X main.BuildTime=%BUILD_TIME% -X main.GitCommit=%GIT_COMMIT%" -o talkabc ./cmd/server
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
echo   Environment: Production
echo   Version: 1.0.0
echo ============================================
pause
