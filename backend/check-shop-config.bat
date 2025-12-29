@echo off
REM Скрипт для проверки конфигурации магазинов для discovery

cd /d %~dp0

echo ========================================
echo Проверка конфигурации магазинов
echo ========================================
echo.

go run cmd/check-shop-config/main.go

pause

