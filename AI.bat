@echo off
setlocal EnableDelayedExpansion

set "CONFIG_DIR=%~dp0opencode-configs"
set "TARGET_DIR=%USERPROFILE%\.config\opencode"

REM ============================================
REM  AI.bat - OpenCode Provider Switcher
REM ============================================

if "%1"=="" goto :menu
if /i "%1"=="copilot" goto :copilot
if /i "%1"=="antigravity" goto :antigravity
if /i "%1"=="gemini" goto :gemini
if /i "%1"=="openai" goto :openai
if /i "%1"=="current" goto :current
goto :menu

:menu
cls
echo.
echo  +===========================================================+
echo  ^|           AI.bat - OpenCode Provider Switcher             ^|
echo  +===========================================================+
echo  ^|                                                           ^|
echo  ^|   [1] Copilot      - GitHub Copilot (Claude)              ^|
echo  ^|   [2] Antigravity  - Antigravity (Claude via Google)      ^|
echo  ^|   [3] Gemini       - Google Gemini Native                 ^|
echo  ^|   [4] OpenAI       - OpenAI GPT-5.2                       ^|
echo  ^|                                                           ^|
echo  ^|   [Q] Quit                                                ^|
echo  ^|                                                           ^|
echo  +===========================================================+
echo.

call :detect_current
echo   Current: !CURRENT_PROVIDER!
echo.

set /p "choice=  Select [1-4, Q]: "

if /i "%choice%"=="1" goto :copilot
if /i "%choice%"=="2" goto :antigravity
if /i "%choice%"=="3" goto :gemini
if /i "%choice%"=="4" goto :openai
if /i "%choice%"=="q" goto :end
goto :menu

:copilot
set "PROVIDER=Copilot"
set "SOURCE_FILE=oh-my-opencode-copilot.json"
goto :apply

:antigravity
set "PROVIDER=Antigravity"
set "SOURCE_FILE=oh-my-opencode-full-claude.json"
goto :apply

:gemini
set "PROVIDER=Gemini"
set "SOURCE_FILE=oh-my-opencode-full-gemini.json"
goto :apply

:openai
set "PROVIDER=OpenAI"
set "SOURCE_FILE=oh-my-opencode-openai.json"
goto :apply

:apply
if not exist "%CONFIG_DIR%\%SOURCE_FILE%" (
    echo.
    echo   [ERROR] Config not found: %SOURCE_FILE%
    pause
    goto :menu
)

if not exist "%TARGET_DIR%" mkdir "%TARGET_DIR%"

copy /y "%CONFIG_DIR%\%SOURCE_FILE%" "%TARGET_DIR%\oh-my-opencode.json" >nul

echo.
echo   [OK] Switched to %PROVIDER%
echo.
echo   Starting OpenCode...
echo.

opencode
goto :end

:current
call :detect_current
echo.
echo   Current Provider: !CURRENT_PROVIDER!
echo.
pause
goto :menu

:detect_current
set "CURRENT_PROVIDER=[None]"
if not exist "%TARGET_DIR%\oh-my-opencode.json" goto :eof

findstr /C:"github-copilot" "%TARGET_DIR%\oh-my-opencode.json" >nul 2>&1
if %errorlevel%==0 (
    set "CURRENT_PROVIDER=Copilot"
    goto :eof
)

findstr /C:"antigravity-claude" "%TARGET_DIR%\oh-my-opencode.json" >nul 2>&1
if %errorlevel%==0 (
    set "CURRENT_PROVIDER=Antigravity"
    goto :eof
)

findstr /C:"antigravity-gemini" "%TARGET_DIR%\oh-my-opencode.json" >nul 2>&1
if %errorlevel%==0 (
    set "CURRENT_PROVIDER=Gemini"
    goto :eof
)

findstr /C:"openai/" "%TARGET_DIR%\oh-my-opencode.json" >nul 2>&1
if %errorlevel%==0 (
    set "CURRENT_PROVIDER=OpenAI"
    goto :eof
)
goto :eof

:end
endlocal
