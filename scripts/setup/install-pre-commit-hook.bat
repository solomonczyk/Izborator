@echo off
REM Скрипт для установки pre-commit hook на локальной машине
REM Это предотвратит случайный коммит .env файлов и API ключей

echo ============================================
echo Installing pre-commit hook...
echo ============================================
echo.

REM Проверяем наличие .git директории
if not exist ".git" (
    echo ERROR: .git directory not found
    echo Run this script from the root of the repository
    exit /b 1
)

REM Копируем pre-commit hook
if exist ".githooks\pre-commit" (
    echo Copying .githooks/pre-commit to .git/hooks/pre-commit...
    
    REM На Windows используем PowerShell
    powershell -Command "Copy-Item '.githooks\pre-commit' '.git\hooks\pre-commit' -Force"
    
    if %ERRORLEVEL% EQU 0 (
        echo.
        echo ✅ Pre-commit hook installed successfully!
        echo.
        echo 📝 What it does:
        echo   - Prevents committing .env files
        echo   - Detects hardcoded API keys (OpenAI, Google)
        echo   - Checks for hardcoded passwords
        echo.
        echo 🚀 Next time you run 'git commit', the hook will run automatically
        echo.
        echo ⚠️  If you need to bypass (not recommended):
        echo   git commit --no-verify
        echo.
    ) else (
        echo ERROR: Failed to copy pre-commit hook
        exit /b 1
    )
) else (
    echo ERROR: .githooks/pre-commit not found
    exit /b 1
)

echo ============================================
echo ✅ Installation complete!
echo ============================================
