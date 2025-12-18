@echo off
REM Скрипт для удаления OpenAI API ключа из истории Git
REM Более безопасная версия - только OpenAI ключи

echo ========================================
echo УДАЛЕНИЕ OPENAI API КЛЮЧА ИЗ ИСТОРИИ GIT
echo ========================================
echo.
echo ВНИМАНИЕ: Этот скрипт перезапишет всю историю Git!
echo Убедись, что:
echo 1. Все изменения закоммичены
echo 2. Создана резервная копия репозитория
echo 3. Все коллабораторы уведомлены
echo.
echo Продолжить? (Y/N)
set /p confirm=
if /i not "%confirm%"=="Y" (
    echo Отменено.
    exit /b
)

set FILTER_BRANCH_SQUELCH_WARNING=1

echo.
echo [1/4] Поиск OpenAI ключей в истории...
"C:\Program Files\Git\bin\bash.exe" -c "git log --all --full-history -p -S 'sk-proj-' | head -20"
echo.
echo Найдены ли ключи выше? (Y/N)
set /p found=
if /i not "%found%"=="Y" (
    echo Ключи не найдены в истории. Проверяю текущие файлы...
    "C:\Program Files\Git\bin\bash.exe" -c "grep -r 'sk-proj-' . --exclude-dir=.git --exclude-dir=node_modules 2>/dev/null | head -5"
    echo.
    echo Продолжить удаление из истории? (Y/N)
    set /p continue=
    if /i not "%continue%"=="Y" (
        echo Отменено.
        exit /b
    )
)

echo.
echo [2/4] Удаление OpenAI API ключей из всех файлов...
"C:\Program Files\Git\bin\bash.exe" -c "git filter-branch -f --tree-filter 'find . -type f ! -path \"*/.git/*\" ! -path \"*/node_modules/*\" ! -path \"*/.next/*\" ! -path \"*/dist/*\" ! -path \"*/build/*\" | while read f; do if [ -f \"$f\" ]; then sed -i.bak \"s/sk-proj-[A-Za-z0-9_-]\{48,\}/REMOVED_OPENAI_API_KEY/g\" \"$f\" 2>/dev/null; sed -i.bak \"s/OPENAI_API_KEY=sk-[A-Za-z0-9_-]\{48,\}/OPENAI_API_KEY=REMOVED/g\" \"$f\" 2>/dev/null; sed -i.bak \"s/sk-[A-Za-z0-9_-]\{48,\}/REMOVED_OPENAI_API_KEY/g\" \"$f\" 2>/dev/null; rm -f \"$f.bak\" 2>/dev/null; fi; done' --prune-empty --tag-name-filter cat -- --all"

echo.
echo [3/4] Удаление всех .env файлов из истории...
"C:\Program Files\Git\bin\bash.exe" -c "git filter-branch -f --tree-filter 'find . -type f -name \".env\" ! -path \"*/.git/*\" -delete 2>/dev/null || true' --prune-empty --tag-name-filter cat -- --all"

echo.
echo [4/4] Очистка ссылок и оптимизация репозитория...
"C:\Program Files\Git\bin\bash.exe" -c "git for-each-ref --format='delete %%(refname)' refs/original | git update-ref --stdin"
"C:\Program Files\Git\bin\bash.exe" -c "git reflog expire --expire=now --all"
"C:\Program Files\Git\bin\bash.exe" -c "git gc --prune=now --aggressive"

echo.
echo ========================================
echo ОЧИСТКА ЗАВЕРШЕНА
echo ========================================
echo.
echo СЛЕДУЮЩИЕ ШАГИ:
echo 1. Проверь изменения: git log --all --oneline | head -10
echo 2. Проверь, что ключи удалены: git log --all -p | findstr /C:"sk-proj-" /C:"OPENAI_API_KEY"
echo 3. Если все ОК, принудительно запушь: git push --force --all
echo 4. Принудительно запушь теги: git push --force --tags
echo.
echo ВНИМАНИЕ: --force перезапишет удаленную историю!
echo Убедись, что все коллабораторы знают об этом.
echo.
echo После push пересоздай API ключ в OpenAI Dashboard!
echo.
pause

