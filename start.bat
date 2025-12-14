@echo off
REM Double-click this file to start Izborator (Backend + Frontend)
echo Starting Izborator...
echo.

REM Get script directory
set "ROOT_DIR=%~dp0"
set "BACKEND_DIR=%ROOT_DIR%backend"
set "FRONTEND_DIR=%ROOT_DIR%frontend"

REM Check if folders exist
if not exist "%BACKEND_DIR%" (
    echo ERROR: Backend folder not found!
    pause
    exit /b 1
)

if not exist "%FRONTEND_DIR%" (
    echo ERROR: Frontend folder not found!
    pause
    exit /b 1
)

REM Clear frontend cache
echo Clearing frontend cache...
if exist "%FRONTEND_DIR%\.next" (
    rmdir /s /q "%FRONTEND_DIR%\.next" 2>nul
)

REM Start Backend in new window
echo Starting Backend API (port 8081)...
start "Izborator Backend" cmd /k "%BACKEND_DIR%\start-backend.bat"

REM Wait a bit
timeout /t 3 /nobreak >nul

REM Start Frontend in new window
echo Starting Frontend (Next.js)...
start "Izborator Frontend" cmd /k "cd /d %FRONTEND_DIR% && npm run dev"

echo.
echo ===============================================================
echo Izborator is starting!
echo.
echo Backend API:  http://localhost:8081
echo Frontend:     http://localhost:3000 (or 3001, 3002)
echo.
echo Two windows opened - one for Backend, one for Frontend
echo Close windows to stop services
echo ===============================================================
echo.
pause
