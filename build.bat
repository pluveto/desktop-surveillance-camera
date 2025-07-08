@echo off
setlocal enabledelayedexpansion

:: Desktop Surveillance Camera - Windows Build Script
:: Author: Desktop Surveillance Camera Project
:: License: MIT

echo ====================================
echo Desktop Surveillance Camera Builder
echo ====================================
echo.

:: Check if Go is installed
where go >nul 2>nul
if %errorlevel% neq 0 (
    echo ERROR: Go is not installed or not in PATH
    echo Please install Go from https://golang.org/dl/
    pause
    exit /b 1
)

:: Check if CGO is available (gcc)
where gcc >nul 2>nul
if %errorlevel% neq 0 (
    echo WARNING: GCC not found in PATH
    echo CGO compilation may fail. Please install MinGW-w64 or TDM-GCC
    echo Download from: https://www.mingw-w64.org/downloads/
    echo.
    set /p continue="Continue anyway? (y/N): "
    if /i not "!continue!"=="y" (
        echo Build cancelled.
        pause
        exit /b 1
    )
)

:: Show Go version
echo Go version:
go version
echo.

:: Set build variables
set BINARY_NAME=surveillance-camera.exe
set BUILD_DIR=build
set VERSION=1.0.0

:: Create build directory
if not exist "%BUILD_DIR%" (
    mkdir "%BUILD_DIR%"
)

echo Building Desktop Surveillance Camera...
echo Binary: %BINARY_NAME%
echo Build Directory: %BUILD_DIR%
echo Version: %VERSION%
echo.

:: Download dependencies
echo [1/3] Downloading dependencies...
go mod tidy
if %errorlevel% neq 0 (
    echo ERROR: Failed to download dependencies
    pause
    exit /b 1
)

:: Build the application
echo [2/3] Building application...
set CGO_ENABLED=1
set GOOS=windows
go build -ldflags "-X main.version=%VERSION%" -o "%BUILD_DIR%\%BINARY_NAME%" .
if %errorlevel% neq 0 (
    echo ERROR: Build failed
    echo.
    echo Common solutions:
    echo 1. Install MinGW-w64 or TDM-GCC for CGO support
    echo 2. Ensure gcc is in your PATH
    echo 3. Check Go installation
    pause
    exit /b 1
)

:: Test the build
echo [3/3] Testing build...
if exist "%BUILD_DIR%\%BINARY_NAME%" (
    echo SUCCESS: Build completed successfully!
    echo.
    echo Binary location: %BUILD_DIR%\%BINARY_NAME%
    
    :: Get file size
    for %%A in ("%BUILD_DIR%\%BINARY_NAME%") do (
        set size=%%~zA
    )
    echo Binary size: !size! bytes
) else (
    echo ERROR: Binary not found after build
    pause
    exit /b 1
)

echo.
echo ====================================
echo Build completed successfully!
echo ====================================
echo.

:: Ask user what to do next
echo What would you like to do next?
echo 1) Test screenshot functionality
echo 2) Run the server with default settings
echo 3) Show help information
echo 4) Create example config file
echo 5) Exit
echo.
set /p choice="Enter your choice (1-5): "

if "%choice%"=="1" goto test_screenshot
if "%choice%"=="2" goto run_server
if "%choice%"=="3" goto show_help
if "%choice%"=="4" goto create_config
if "%choice%"=="5" goto end

echo Invalid choice. Exiting...
goto end

:test_screenshot
echo.
echo Testing screenshot functionality...
echo ====================================
"%BUILD_DIR%\%BINARY_NAME%" -test
echo.
echo Screenshot test completed.
pause
goto end

:run_server
echo.
echo Starting server with default settings...
echo ====================================
echo Server will start on http://localhost:9981
echo Press Ctrl+C to stop the server
echo.
pause
"%BUILD_DIR%\%BINARY_NAME%"
goto end

:show_help
echo.
echo Application help:
echo ====================================
"%BUILD_DIR%\%BINARY_NAME%" -help
echo.
pause
goto end

:create_config
echo.
echo Creating example configuration file...
echo ====================================
(
echo {
echo   "server": {
echo     "host": "0.0.0.0",
echo     "port": 9981
echo   },
echo   "capture": {
echo     "mode": "ondemand",
echo     "interval": "5s"
echo   }
echo }
) > config.json.example

if exist config.json.example (
    echo Example configuration created: config.json.example
    echo You can rename it to config.json and modify as needed.
) else (
    echo ERROR: Failed to create example configuration
)
echo.
pause
goto end

:end
echo.
echo Thanks for using Desktop Surveillance Camera!
echo.
echo Quick start commands:
echo   %BUILD_DIR%\%BINARY_NAME% -test          # Test screenshot
echo   %BUILD_DIR%\%BINARY_NAME%                # Run server
echo   %BUILD_DIR%\%BINARY_NAME% -help          # Show help
echo.
echo Visit: http://localhost:9981 to access the web interface
echo.
pause