@echo off
chcp 65001 >nul
echo ============================================
echo   TalkABC - Test Build Script
echo ============================================
echo.
echo Environment: Test
echo.

echo Cleaning previous build...
if exist talkabc-test (
    echo Found existing talkabc-test file, deleting...
    del /f /q talkabc-test
)
if exist talkabc-test.exe (
    echo Found existing talkabc-test.exe file, deleting...
    del /f /q talkabc-test.exe
)
echo.

echo Building for Test environment...
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

go build -ldflags "-w -s -X main.Version=1.0.0-test -X main.BuildTime=%BUILD_TIME% -X main.GitCommit=%GIT_COMMIT%" -o talkabc-test ./cmd/server
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
echo   Output: talkabc-test (Linux 64-bit)
echo   Environment: Test
echo ============================================
pause
