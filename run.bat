@echo off
title Workout Challenge Tracker Launcher
echo ==============================================
echo   Workout Challenge Tracker Launcher
echo ==============================================

echo [1/3] Starting database container...
docker compose up -d
if %errorlevel% neq 0 (
    echo [ERROR] Failed to start Docker container. Make sure Docker Desktop is running!
    pause
    exit /b %errorlevel%
)

echo [2/3] Opening application in your browser...
timeout /t 2 /nobreak >nul
start http://localhost:8080/

echo [3/3] Launching Go Backend Server...
echo ----------------------------------------------
go run main.go
if %errorlevel% neq 0 (
    echo [ERROR] Failed to run backend server.
    pause
    exit /b %errorlevel%
)
