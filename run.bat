@echo off
:: Desktop Surveillance Camera - Quick Run Script
:: This script runs the surveillance camera with default settings

setlocal enabledelayedexpansion

echo ====================================
echo Desktop Surveillance Camera
echo ====================================
echo.

:: Check if binary exists
if not exist "build\surveillance-camera.exe" (
    if not exist "surveillance-camera.exe" (
        echo ERROR: surveillance-camera.exe not found!
        echo.
        echo Please build the application first:
        echo   1. Run build.bat to compile the application
        echo   2. Or run: go build -o surveillance-camera.exe .
        echo.
        pause
        exit /b 1
    )
    set BINARY=surveillance-camera.exe
) else (
    set BINARY=build\surveillance-camera.exe
)

:: Show current configuration
echo Current settings:
echo Binary: !BINARY!

if exist config.json (
    echo Config: config.json (custom)
) else (
    echo Config: Default settings will be used
    echo   - Server: http://0.0.0.0:9981
    echo   - Mode: On-demand
)
echo.

:: Get local IP for easy access
echo Detecting network information...
for /f "tokens=2 delims=:" %%a in ('ipconfig ^| findstr /i "IPv4"') do (
    set ip=%%a
    set ip=!ip: =!
    if not "!ip!"=="" (
        echo Local IP: !ip!
        echo Access URL: http://!ip!:9981
        goto ip_found
    )
)
:ip_found
echo.

:: Ask for run mode
echo Select run mode:
echo 1) Start server (default)
echo 2) Test screenshot functionality
echo 3) Show help
echo 4) Custom command
echo.
set /p mode="Enter choice (1-4) or press Enter for default: "

if "%mode%"=="" set mode=1
if "%mode%"=="1" goto start_server
if "%mode%"=="2" goto test_screenshot  
if "%mode%"=="3" goto show_help
if "%mode%"=="4" goto custom_command

echo Invalid choice, starting server...
goto start_server

:start_server
echo.
echo ====================================
echo Starting Desktop Surveillance Camera
echo ====================================
echo.
echo Server starting...
echo - Web interface: http://localhost:9981
echo - API endpoint: http://localhost:9981/last
echo.
echo Press Ctrl+C to stop the server
echo.
pause
echo Starting server...
"!BINARY!"
goto end

:test_screenshot
echo.
echo ====================================
echo Testing Screenshot Functionality
echo ====================================
echo.
"!BINARY!" -test
echo.
echo Test completed. Check the generated screenshot file.
pause
goto end

:show_help
echo.
echo ====================================
echo Application Help
echo ====================================
echo.
"!BINARY!" -help
echo.
pause
goto end

:custom_command
echo.
echo Enter custom command arguments (without the executable name):
echo Example: -config my-config.json
echo.
set /p args="Arguments: "
echo.
echo Running: "!BINARY!" !args!
echo.
"!BINARY!" !args!
pause
goto end

:end
echo.
echo Desktop Surveillance Camera stopped.
echo.