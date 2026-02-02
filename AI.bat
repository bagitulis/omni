@echo off
setlocal EnableDelayedExpansion

set "CONFIG_DIR=%~dp0opencode-configs"
set "TARGET_DIR=%USERPROFILE%\.config\opencode"

REM ============================================
REM  AI.bat - OpenCode Provider Switcher
REM  Smart Sync: newer file always wins
REM ============================================

if "%1"=="" goto :menu
if /i "%1"=="copilot" goto :copilot
if /i "%1"=="antigravity" goto :antigravity
if /i "%1"=="mix" goto :mix
if /i "%1"=="current" goto :current
goto :menu

:menu
cls
echo.
echo  +===========================================================+
echo  ^|           AI.bat - OpenCode Provider Switcher             ^|
echo  +===========================================================+
echo  ^|                                                           ^|
echo  ^|   [1] Copilot      - GitHub Copilot (Claude + Gemini)     ^|
echo  ^|   [2] Antigravity  - Antigravity (Claude + Gemini)        ^|
echo  ^|   [3] Mix          - Copilot Opus + Antigravity Gemini    ^|
echo  ^|                                                           ^|
echo  ^|   [Q] Quit                                                ^|
echo  ^|                                                           ^|
echo  +===========================================================+
echo.

call :detect_current
echo   Current Provider: !CURRENT_PROVIDER!
echo.

set /p "choice=  Select [1-3, Q]: "

if /i "%choice%"=="1" goto :copilot
if /i "%choice%"=="2" goto :antigravity
if /i "%choice%"=="3" goto :mix
if /i "%choice%"=="q" goto :end
goto :menu

:copilot
set "PROVIDER=Copilot"
set "SOURCE_FILE=oh-my-opencode-copilot.json"
set "SYNC_ACCOUNTS=0"
goto :apply

:antigravity
set "PROVIDER=Antigravity"
set "SOURCE_FILE=oh-my-opencode-antigravity.json"
set "SYNC_ACCOUNTS=1"
goto :apply

:mix
set "PROVIDER=Mix"
set "SOURCE_FILE=oh-my-opencode-mix.json"
set "SYNC_ACCOUNTS=1"
goto :apply

:apply
if not exist "%CONFIG_DIR%\%SOURCE_FILE%" (
    echo.
    echo   [ERROR] Config not found: %SOURCE_FILE%
    pause
    goto :menu
)

if not exist "%TARGET_DIR%" mkdir "%TARGET_DIR%"

REM Copy oh-my-opencode config (provider/model settings)
copy /y "%CONFIG_DIR%\%SOURCE_FILE%" "%TARGET_DIR%\oh-my-opencode.json" >nul

REM Copy opencode.json (model definitions)
if exist "%CONFIG_DIR%\opencode.json" (
    copy /y "%CONFIG_DIR%\opencode.json" "%TARGET_DIR%\opencode.json" >nul
)

REM Copy antigravity.json (multi-account behavior settings)
if exist "%CONFIG_DIR%\antigravity.json" (
    copy /y "%CONFIG_DIR%\antigravity.json" "%TARGET_DIR%\antigravity.json" >nul
)

REM Smart sync antigravity-accounts.json only for Antigravity/Mix/Test-Flash/Test-Pro
if "%SYNC_ACCOUNTS%"=="1" (
    call :smart_sync_accounts
)

echo.
echo   [OK] Switched to %PROVIDER%
echo.
echo   Starting OpenCode...
echo.

opencode
goto :end

REM ============================================
REM  SMART SYNC for antigravity-accounts.json
REM  Sync across 3 locations - newest file wins
REM  Locations: opencode-configs, .config\opencode, AppData\Roaming\opencode
REM ============================================
:smart_sync_accounts
set "LOC1=%CONFIG_DIR%\antigravity-accounts.json"
set "LOC2=%USERPROFILE%\.config\opencode\antigravity-accounts.json"
set "LOC3=%APPDATA%\opencode\antigravity-accounts.json"

REM Ensure target directories exist
if not exist "%USERPROFILE%\.config\opencode" mkdir "%USERPROFILE%\.config\opencode"
if not exist "%APPDATA%\opencode" mkdir "%APPDATA%\opencode"

REM Get ticks for each file (0 if not exists)
set "TICKS1=0"
set "TICKS2=0"
set "TICKS3=0"
if exist "%LOC1%" for /f %%i in ('powershell -NoProfile -Command "(Get-Item '%LOC1%').LastWriteTime.Ticks"') do set "TICKS1=%%i"
if exist "%LOC2%" for /f %%i in ('powershell -NoProfile -Command "(Get-Item '%LOC2%').LastWriteTime.Ticks"') do set "TICKS2=%%i"
if exist "%LOC3%" for /f %%i in ('powershell -NoProfile -Command "(Get-Item '%LOC3%').LastWriteTime.Ticks"') do set "TICKS3=%%i"

REM No files exist anywhere
if "%TICKS1%"=="0" if "%TICKS2%"=="0" if "%TICKS3%"=="0" (
    echo   [SKIP] No antigravity-accounts.json found in any location
    goto :eof
)

REM Find newest file using PowerShell
for /f %%i in ('powershell -NoProfile -Command "$t1=%TICKS1%;$t2=%TICKS2%;$t3=%TICKS3%;if($t1 -ge $t2 -and $t1 -ge $t3){1}elseif($t2 -ge $t1 -and $t2 -ge $t3){2}else{3}"') do set "NEWEST=%%i"

if "%NEWEST%"=="1" (
    set "NEWEST_FILE=%LOC1%"
    set "NEWEST_NAME=opencode-configs"
) else if "%NEWEST%"=="2" (
    set "NEWEST_FILE=%LOC2%"
    set "NEWEST_NAME=.config\opencode"
) else (
    set "NEWEST_FILE=%LOC3%"
    set "NEWEST_NAME=AppData\Roaming\opencode"
)

echo   [SYNC] Newest: %NEWEST_NAME%

REM Copy newest to all other locations
if not "%NEWEST%"=="1" (
    copy /y "%NEWEST_FILE%" "%LOC1%" >nul
    echo   [OK] -^> opencode-configs
)
if not "%NEWEST%"=="2" (
    copy /y "%NEWEST_FILE%" "%LOC2%" >nul
    echo   [OK] -^> .config\opencode
)
if not "%NEWEST%"=="3" (
    copy /y "%NEWEST_FILE%" "%LOC3%" >nul
    echo   [OK] -^> AppData\Roaming\opencode
)
goto :eof

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

REM Check for Mix (has both github-copilot and antigravity)
findstr /C:"github-copilot" "%TARGET_DIR%\oh-my-opencode.json" >nul 2>&1
if %errorlevel%==0 (
    findstr /C:"antigravity" "%TARGET_DIR%\oh-my-opencode.json" >nul 2>&1
    if !errorlevel!==0 (
        set "CURRENT_PROVIDER=Mix"
        goto :eof
    )
    set "CURRENT_PROVIDER=Copilot"
    goto :eof
)

findstr /C:"antigravity" "%TARGET_DIR%\oh-my-opencode.json" >nul 2>&1
if %errorlevel%==0 (
    set "CURRENT_PROVIDER=Antigravity"
    goto :eof
)
goto :eof

:end
endlocal
