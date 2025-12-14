@echo off
REM Start Frontend on port 3003 (3000 is occupied by another project)
REM IMPORTANT: This script sets environment variables for Next.js
cd /d %~dp0
echo Starting Frontend on port 3003...
echo API Base URL: http://localhost:3002
echo.
set PORT=3003
set NEXT_PUBLIC_API_BASE=http://localhost:3002
npm run dev
pause

