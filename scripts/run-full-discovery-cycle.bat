@echo off
REM Полный цикл Discovery: Discovery -> Classifier -> AutoConfig
REM Использование: scripts\run-full-discovery-cycle.bat [limit]

setlocal

set LIMIT_AUTOCONFIG=%1
if "%LIMIT_AUTOCONFIG%"=="" set LIMIT_AUTOCONFIG=5

echo [START] Запуск полного цикла Discovery
echo.

cd /d "%~dp0..\backend"

REM Шаг 1: Discovery
echo [STEP 1] Запуск Discovery (поиск кандидатов)...
echo.

if exist discovery.exe (
    discovery.exe
) else if exist discovery (
    discovery
) else (
    echo [ERROR] discovery не найден!
    exit /b 1
)

if errorlevel 1 (
    echo [WARNING] Discovery завершился с предупреждениями
) else (
    echo [OK] Discovery завершен успешно
)
echo.

REM Шаг 2: Classifier
echo [STEP 2] Запуск Classifier (классификация)...
echo.

if exist classifier.exe (
    classifier.exe -classify-all
) else if exist classifier (
    classifier -classify-all
) else (
    echo [ERROR] classifier не найден!
    exit /b 1
)

if errorlevel 1 (
    echo [WARNING] Classifier завершился с предупреждениями
) else (
    echo [OK] Classifier завершен успешно
)
echo.

REM Шаг 3: AutoConfig
echo [STEP 3] Запуск AutoConfig (AI генерация селекторов)...
echo Обрабатываем %LIMIT_AUTOCONFIG% кандидатов...
echo.

if exist autoconfig.exe (
    autoconfig.exe -limit %LIMIT_AUTOCONFIG%
) else if exist autoconfig (
    autoconfig -limit %LIMIT_AUTOCONFIG%
) else (
    echo [ERROR] autoconfig не найден!
    exit /b 1
)

if errorlevel 1 (
    echo [WARNING] AutoConfig завершился с предупреждениями
) else (
    echo [OK] AutoConfig завершен успешно
)
echo.

echo [COMPLETE] Полный цикл Discovery завершен!
echo.

endlocal

