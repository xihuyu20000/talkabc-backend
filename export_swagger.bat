@echo off
chcp 65001 >nul
echo ============================================
echo   TalkABC - Swagger Export Script
echo ============================================
echo.

setlocal enabledelayedexpansion

set CHECK_ONLY=0
set VERBOSE=0
set OUTPUT_DIR=./swagger

:parse_args
if "%~1"=="" goto end_parse
if "%~1"=="--check" (
    set CHECK_ONLY=1
    shift
    goto parse_args
)
if "%~1"=="--verbose" (
    set VERBOSE=1
    shift
    goto parse_args
)
if "%~1"=="--output" (
    set OUTPUT_DIR=%~2
    shift
    shift
    goto parse_args
)
shift
goto parse_args
:end_parse

for /f "delims=" %%i in ('go env GOPATH') do set GOPATH=%%i
set PATH=%GOPATH%\bin;%PATH%

echo Checking swag tool...
where swag >nul 2>&1
if %errorlevel% neq 0 (
    echo swag not found, installing...
    go install github.com/swaggo/swag/cmd/swag@latest
    if %errorlevel% neq 0 (
        echo.
        echo ============================================
        echo   ERROR: Failed to install swag!
        echo ============================================
        pause
        exit /b 1
    )
    echo swag installed successfully.
) else (
    echo swag is already installed.
)

if %CHECK_ONLY% equ 1 (
    echo.
    echo ============================================
    echo   Check completed: swag is available
    echo ============================================
    pause
    exit /b 0
)

echo.
echo Generating Swagger documentation...
echo Output directory: %OUTPUT_DIR%
echo.

swag init --dir ./cmd/server,./internal/handler --output %OUTPUT_DIR%
if %errorlevel% neq 0 (
    echo.
    echo ============================================
    echo   ERROR: Failed to generate Swagger docs!
    echo ============================================
    pause
    exit /b 1
)

echo Swagger documentation generated successfully!
echo.

if %VERBOSE% equ 1 (
    echo Generated files:
    echo   - %OUTPUT_DIR%\swagger.json
    echo   - %OUTPUT_DIR%\swagger.yaml
    echo   - %OUTPUT_DIR%\docs.go
)

echo.
echo ============================================
echo   Export completed!
echo.
echo   Swagger UI: http://localhost:8080/swagger/index.html
echo   Swagger JSON: %OUTPUT_DIR%\swagger.json
echo   Apifox Import: %OUTPUT_DIR%\swagger.json
echo ============================================
pause