@echo off
REM health-check.bat - Simple script to check the health of the watcher
REM Returns 0 if the watcher is healthy, 1 otherwise

setlocal enabledelayedexpansion

REM Path to the metrics file
if "%METRICS_FILE%"=="" (
    set METRICS_FILE=watcher_metrics.json
)

REM Check if the metrics file exists
if not exist "%METRICS_FILE%" (
    echo ERROR: Metrics file not found: %METRICS_FILE%
    exit /b 1
)

REM Check the health status from the metrics file
findstr /c:"\"health\":\"true\"" "%METRICS_FILE%" >nul
if errorlevel 1 (
    echo ERROR: Watcher reports unhealthy status
    exit /b 1
)

REM Check when the last event was processed
for /f "tokens=2 delims=:," %%a in ('findstr /c:"\"last_event\":" "%METRICS_FILE%"') do (
    set LAST_EVENT=%%a
    set LAST_EVENT=!LAST_EVENT:"=!
    set LAST_EVENT=!LAST_EVENT: =!
)

REM Output the status
echo OK: Watcher is healthy
echo Last event processed: %LAST_EVENT%
exit /b 0
