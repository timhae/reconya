@echo off
echo Starting RecoNya with CGO enabled for Windows...
echo.

REM Change to backend directory
cd /d "%~dp0..\backend"

REM Check if .env file exists
if not exist .env (
    echo Error: .env file not found in backend directory!
    echo Please copy .env.example to .env and configure it first.
    pause
    exit /b 1
)

REM Set CGO_ENABLED=1 for SQLite support on Windows
set CGO_ENABLED=1

REM Start the application
echo Starting backend with CGO enabled...
go run ./cmd

pause