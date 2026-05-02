@echo off
REM Omni Build System - Windows Batch Wrapper
REM Usage: build.bat [command] [options]
REM   build.bat              - Interactive menu
REM   build.bat smart        - Smart build
REM   build.bat sync-export  - Export DB to NDJSON (git-friendly)
REM   build.bat sync-import  - Import DB from NDJSON
python "%~dp0build.py" %*
