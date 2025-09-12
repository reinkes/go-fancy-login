@echo off
REM Batch wrapper for fancy-login-go
REM For Command Prompt users

fancy-login-go.exe %*
if %errorlevel% equ 0 (
    if exist "%TEMP%\aws_profile.bat" (
        call "%TEMP%\aws_profile.bat"
        echo ✅ AWS_PROFILE environment variable updated
    )
)