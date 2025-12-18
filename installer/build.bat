@echo off
REM smaqit installer build script for Windows

setlocal enabledelayedexpansion

REM Detect version from git tags, fallback to "dev"
for /f "tokens=*" %%i in ('git describe --tags --always --dirty 2^>nul') do set VERSION=%%i
if "%VERSION%"=="" set VERSION=dev

set BINARY_NAME=smaqit
set DIST_DIR=dist
set LDFLAGS=-ldflags "-X main.Version=%VERSION%"

REM Main command dispatcher
if "%1"=="" goto build
if /i "%1"=="build" goto build
if /i "%1"=="build-all" goto build_all
if /i "%1"=="linux" goto build_linux
if /i "%1"=="darwin-intel" goto build_darwin_intel
if /i "%1"=="darwin-arm" goto build_darwin_arm
if /i "%1"=="windows" goto build_windows
if /i "%1"=="version" goto show_version
if /i "%1"=="clean" goto clean
if /i "%1"=="help" goto help
if /i "%1"=="--help" goto help
if /i "%1"=="-h" goto help

echo Unknown command: %1
echo.
goto help

:build
echo Building %BINARY_NAME% version %VERSION% for current platform...
if not exist %DIST_DIR% mkdir %DIST_DIR%
go build %LDFLAGS% -o %DIST_DIR%\%BINARY_NAME%.exe .
echo Built: %DIST_DIR%\%BINARY_NAME%.exe
goto end

:build_all
echo Building %BINARY_NAME% version %VERSION% for all platforms...
call :build_linux
call :build_darwin_intel
call :build_darwin_arm
call :build_windows
echo All builds complete.
goto end

:build_linux
echo Building for linux/amd64...
if not exist %DIST_DIR% mkdir %DIST_DIR%
set GOOS=linux
set GOARCH=amd64
go build %LDFLAGS% -o %DIST_DIR%\%BINARY_NAME%_linux_amd64 .
goto :eof

:build_darwin_intel
echo Building for darwin/amd64...
if not exist %DIST_DIR% mkdir %DIST_DIR%
set GOOS=darwin
set GOARCH=amd64
go build %LDFLAGS% -o %DIST_DIR%\%BINARY_NAME%_darwin_amd64 .
goto :eof

:build_darwin_arm
echo Building for darwin/arm64...
if not exist %DIST_DIR% mkdir %DIST_DIR%
set GOOS=darwin
set GOARCH=arm64
go build %LDFLAGS% -o %DIST_DIR%\%BINARY_NAME%_darwin_arm64 .
goto :eof

:build_windows
echo Building for windows/amd64...
if not exist %DIST_DIR% mkdir %DIST_DIR%
set GOOS=windows
set GOARCH=amd64
go build %LDFLAGS% -o %DIST_DIR%\%BINARY_NAME%_windows_amd64.exe .
goto :eof

:show_version
echo %VERSION%
goto end

:clean
echo Cleaning build artifacts...
if exist %DIST_DIR% rmdir /s /q %DIST_DIR%
echo Done.
goto end

:help
echo smaqit installer build script
echo.
echo Usage: build.bat [command]
echo.
echo Commands:
echo   build          - Build for current platform
echo   build-all      - Build for all platforms
echo   linux          - Build for Linux (amd64)
echo   darwin-intel   - Build for macOS Intel (amd64)
echo   darwin-arm     - Build for macOS Apple Silicon (arm64)
echo   windows        - Build for Windows (amd64)
echo   version        - Show version that would be built
echo   clean          - Remove build artifacts
echo   help           - Show this help message
echo.
echo Current version: %VERSION%
goto end

:end
endlocal
