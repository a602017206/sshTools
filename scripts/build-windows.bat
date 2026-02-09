@echo off

echo ==================================
echo Building AHaSSHTools for Windows
echo ==================================

for /f "delims=" %%i in ('git describe --tags --abbrev=0 2^>nul') do set GIT_VERSION=%%i

if "%GIT_VERSION%"=="" (
    echo No git tag found, using default version
    set VERSION=
) else (
    set VERSION=-ldflags=-X main.Version=%GIT_VERSION%
    echo Using version from git tag: %GIT_VERSION%
)

echo Cleaning previous build...
if exist build\bin\AHaSSHTools.exe del build\bin\AHaSSHTools.exe

echo Building application...
wails build -clean %VERSION%

if exist build\bin\AHaSSHTools.exe (
    echo.
    echo ==================================
    echo Build complete!
    echo App location: build\bin\AHaSSHTools.exe
    echo ==================================
) else (
    echo.
    echo Error: Build failed, exe not found
    exit /b 1
)
