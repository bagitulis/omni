@echo off
REM ============================================
REM Omni Build System - Python Edition
REM ============================================
REM
REM Modern build automation with intelligent error recovery.
REM
REM Usage:
REM   build.bat              - Auto-run smart build (RECOMMENDED)
REM   build.bat smart        - Cache deps, rebuild code [RECOMMENDED]
REM   build.bat quick        - Quick restart (~10s)
REM   build.bat full         - Clean rebuild (~5-10m)
REM   build.bat validate     - Check environment only
REM   build.bat clean        - Cleanup Docker resources
REM   build.bat status       - Show container/service status
REM
REM Options:
REM   --spec=<level>         - lowspec/standard/highspec
REM   --skip-frontend        - Skip frontend build
REM
REM Examples:
REM   build.bat                        (auto smart-standard)
REM   build.bat smart
REM   build.bat smart --spec=lowspec
REM   build.bat smart --skip-frontend
REM   build.bat full --spec=highspec
REM ============================================

setlocal enabledelayedexpansion

set SCRIPT_DIR=%~dp0
set PYTHON_BUILD_DIR=%SCRIPT_DIR%scripts\python-build
set VENV_DIR=%PYTHON_BUILD_DIR%\.venv
set VENV_PYTHON=%VENV_DIR%\Scripts\python.exe

REM Check if virtual environment exists
if not exist "%VENV_DIR%" (
    echo.
    echo [SETUP] Virtual environment not found - running setup...
    echo.
    python "%PYTHON_BUILD_DIR%\setup.py"
    if errorlevel 1 (
        echo.
        echo [ERROR] Setup failed - please check Python installation
        echo.
        pause
        exit /b 1
    )
    echo.
    echo [SUCCESS] Setup completed
    echo.
)

REM If no arguments provided, show interactive menu
if "%~1"=="" (
    echo.
    echo ============================================================
    echo   OMNI BUILD SYSTEM - Python Edition
    echo ============================================================
    echo.
    echo   Available commands:
    echo.
    echo   1. smart     - Smart build [RECOMMENDED]
    echo   2. quick     - Quick restart (~10s)
    echo   3. full      - Full rebuild (~5-10m)
    echo   4. validate  - Check environment only
    echo   5. clean     - Cleanup Docker
    echo   6. status    - Show container status
    echo.
    echo ============================================================
    echo.
    
    choice /C 123456 /N /M "Select option (1-6) or press Ctrl+C to cancel: "
    
    if errorlevel 6 set "CMD=status"
    if errorlevel 5 set "CMD=clean"
    if errorlevel 4 set "CMD=validate"
    if errorlevel 3 set "CMD=full"
    if errorlevel 2 set "CMD=quick"
    if errorlevel 1 set "CMD=smart"
    
    echo.
    echo Running: build.bat !CMD!
    echo.
    
    "%VENV_PYTHON%" "%PYTHON_BUILD_DIR%\run.py" !CMD!
    set EXIT_CODE=!ERRORLEVEL!
    
    echo.
    pause
    exit /b !EXIT_CODE!
)

REM Run CLI with provided arguments
"%VENV_PYTHON%" "%PYTHON_BUILD_DIR%\run.py" %*

REM Capture exit code
set EXIT_CODE=%ERRORLEVEL%

exit /b %EXIT_CODE%
