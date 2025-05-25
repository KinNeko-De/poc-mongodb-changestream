@echo off
REM reset-resume-token.bat - Script to reset the resume token for the watcher

REM Default file path
SET RESUME_TOKEN_FILE=watcher\resume_token.json

REM Check if file exists
IF EXIST "%RESUME_TOKEN_FILE%" (
    echo Removing resume token file: %RESUME_TOKEN_FILE%
    del /f "%RESUME_TOKEN_FILE%"
    echo Resume token has been reset. The watcher will start with a new change stream.
) ELSE (
    echo No resume token file found at %RESUME_TOKEN_FILE%
    echo The watcher will start with a new change stream on next run.
)
