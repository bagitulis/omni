@echo off
REM ============================================
REM Build and Deploy Pipeline (Auto-Fix Edition)
REM ============================================
REM 
REM This is a wrapper script that calls the PowerShell
REM build system with full auto-fix capabilities.
REM
REM All errors are automatically detected and fixed:
REM   - Docker Windows/Linux mode switching
REM   - WSL mount cache issues
REM   - Network errors
REM   - Build cache corruption
REM   - Disk space cleanup
REM
REM Usage:
REM   build.bat              - Interactive menu
REM   build.bat quick        - Quick restart
REM   build.bat smart        - Cache deps, rebuild code [RECOMMENDED]
REM   build.bat full         - Clean rebuild
REM   build.bat clean        - Aggressive cleanup
REM   build.bat validate     - Validate only
REM
REM Options:
REM   -Spec <spec>          - lowspec/standard/highspec
REM   -ShowBuildOutput      - Show full Docker build output (realtime)
REM
REM Examples:
REM   build.bat smart -Spec standard
REM   build.bat full
REM   build.bat incremental -ShowBuildOutput
REM ============================================

setlocal

REM Get the directory where this script is located
set "SCRIPT_DIR=%~dp0"

REM Check if PowerShell script exists
if not exist "%SCRIPT_DIR%scripts\build\build.ps1" (
    echo.
    echo [ERROR] Build script not found!
    echo Expected: %SCRIPT_DIR%scripts\build\build.ps1
    echo.
    pause
    exit /b 1
)

REM Parse arguments
set "MODE=menu"
set "EXTRA_ARGS="

if "%~1"=="" goto RUN_SCRIPT

REM Check for mode argument
if /i "%~1"=="quick" set "MODE=quick"
if /i "%~1"=="smart" set "MODE=smart"
if /i "%~1"=="full" set "MODE=full"
if /i "%~1"=="clean" set "MODE=clean"
if /i "%~1"=="validate" set "MODE=validate"

REM Collect extra arguments
shift
:PARSE_ARGS
if "%~1"=="" goto RUN_SCRIPT
set "EXTRA_ARGS=%EXTRA_ARGS% %~1"
shift
goto PARSE_ARGS

:RUN_SCRIPT
REM Run PowerShell script with execution policy bypass
powershell.exe -NoProfile -ExecutionPolicy Bypass -File "%SCRIPT_DIR%scripts\build\build.ps1" -Mode %MODE% %EXTRA_ARGS%

REM Capture exit code
set "EXIT_CODE=%ERRORLEVEL%"

REM Keep the window open so users can read the logs
echo.
echo ============================================
if "%EXIT_CODE%"=="0" (
    echo Build completed successfully. Exit code: %EXIT_CODE%
) else (
    echo Build finished with errors. Exit code: %EXIT_CODE%
)
echo ============================================
echo Press any key to close this window...

REM Allow bypassing the pause by setting NO_PAUSE=1
if not defined NO_PAUSE pause >nul

REM Exit with the same code
exit /b %EXIT_CODE%
