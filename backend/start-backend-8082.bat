@echo off
REM Start Backend with port 8082
cd /d %~dp0
set DB_HOST=localhost
set DB_PORT=5433
set DB_USER=postgres
set DB_PASSWORD=postgres
set DB_NAME=izborator
set SERVER_PORT=8082
go run cmd/api/main.go
pause

