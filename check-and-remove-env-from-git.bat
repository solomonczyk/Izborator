@echo off
REM Скрипт для проверки и удаления backend/.env из истории Git

echo Проверка, попал ли backend/.env в Git...
echo.

REM Проверяем, есть ли файл в Git
git ls-files backend/.env >nul 2>&1
if %errorlevel% == 0 (
    echo [КРИТИЧНО] Файл backend/.env найден в Git репозитории!
    echo.
    echo Удаляем из индекса...
    git rm --cached backend/.env
    echo.
    echo Коммитим удаление...
    git commit -m "Remove backend/.env from repository"
    echo.
    echo [ВАЖНО] Теперь нужно удалить из истории Git:
    echo Запусти: clean-secrets-history.bat
) else (
    echo [OK] Файл backend/.env не в Git репозитории
    echo.
    echo Проверяем историю Git на наличие API ключей...
    git log --all --full-history --source -- "*backend/.env" | findstr /C:"commit" >nul 2>&1
    if %errorlevel% == 0 (
        echo [ВНИМАНИЕ] Файл backend/.env найден в истории Git!
        echo Нужно очистить историю через clean-secrets-history.bat
    ) else (
        echo [OK] Файл backend/.env не найден в истории Git
    )
)

echo.
echo Проверка завершена.

