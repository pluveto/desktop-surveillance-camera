@echo off
:: Desktop Surveillance Camera - Development Setup Script
:: This script sets up the development environment on Windows

setlocal enabledelayedexpansion

echo ====================================
echo Desktop Surveillance Camera
echo Development Setup
echo ====================================
echo.

:: Check if Go is installed
echo [1/4] Checking Go installation...
where go >nul 2>nul
if %errorlevel% neq 0 (
    echo ERROR: Go is not installed!
    echo.
    echo Please install Go from: https://golang.org/dl/
    echo After installation, restart this script.
    pause
    exit /b 1
)

go version
echo Go is installed ✓
echo.

:: Check if Git is installed
echo [2/4] Checking Git installation...
where git >nul 2>nul
if %errorlevel% neq 0 (
    echo WARNING: Git is not installed
    echo Git is recommended for version control
    echo Download from: https://git-scm.com/download/win
) else (
    git --version
    echo Git is installed ✓
)
echo.

:: Check CGO environment
echo [3/4] Checking CGO environment...
where gcc >nul 2>nul
if %errorlevel% neq 0 (
    echo WARNING: GCC not found in PATH
    echo.
    echo CGO is required for Windows API calls.
    echo Please install one of the following:
    echo.
    echo 1. TDM-GCC: https://jmeubank.github.io/tdm-gcc/
    echo 2. MinGW-w64: https://www.mingw-w64.org/downloads/
    echo 3. MSYS2: https://www.msys2.org/
    echo.
    echo After installation, make sure 'gcc' is in your PATH
    echo.
    set /p continue="Continue setup without CGO? (y/N): "
    if /i not "!continue!"=="y" (
        echo Setup cancelled. Please install a C compiler first.
        pause
        exit /b 1
    )
) else (
    gcc --version | findstr "gcc"
    echo GCC is installed ✓
)
echo.

:: Initialize Go module and download dependencies
echo [4/4] Setting up Go module...
if not exist go.mod (
    echo Initializing Go module...
    go mod init desktop-surveillance-camera
)

echo Downloading dependencies...
go mod tidy
if %errorlevel% neq 0 (
    echo ERROR: Failed to download dependencies
    pause
    exit /b 1
)
echo Dependencies installed ✓
echo.

:: Create necessary directories
echo Creating project directories...
if not exist build mkdir build
if not exist logs mkdir logs
echo Directories created ✓
echo.

:: Create example configuration
echo Creating example configuration...
if not exist config.json.example (
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
    echo Example config created: config.json.example ✓
)
echo.

:: Create .gitignore if using Git
if exist .git (
    if not exist .gitignore (
        echo Creating .gitignore...
        (
        echo # Binaries
        echo *.exe
        echo build/
        echo.
        echo # Logs
        echo *.log
        echo logs/
        echo.
        echo # Screenshots
        echo screenshot_*.png
        echo.
        echo # Config files
        echo config.json
        echo.
        echo # IDE files
        echo .vscode/
        echo *.swp
        echo *.swo
        echo.
        echo # OS files
        echo Thumbs.db
        echo .DS_Store
        ) > .gitignore
        echo .gitignore created ✓
    )
)

echo.
echo ====================================
echo Development Setup Complete!
echo ====================================
echo.
echo What you can do now:
echo.
echo 1. Build the project:
echo    build.bat
echo.
echo 2. Run the application:
echo    run.bat
echo.
echo 3. Manual commands:
echo    go build -o surveillance-camera.exe .
echo    surveillance-camera.exe -test
echo    surveillance-camera.exe
echo.
echo 4. Access web interface:
echo    http://localhost:9981
echo.

set /p action="Would you like to build the project now? (y/N): "
if /i "!action!"=="y" (
    echo.
    echo Starting build process...
    call build.bat
) else (
    echo.
    echo Setup completed. You can run 'build.bat' when ready to build.
)

echo.
pause