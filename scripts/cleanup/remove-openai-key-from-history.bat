@echo off
REM Скрипт для удаления OpenAI API ключа из истории Git
REM ВНИМАНИЕ: Это перезапишет всю историю Git!

echo ========================================
echo УДАЛЕНИЕ OPENAI API КЛЮЧА ИЗ ИСТОРИИ GIT
echo ========================================
echo.

REM Проверяем наличие ключа в истории
echo [ШАГ 1/5] Проверка наличия ключа в истории Git...
git log --all -p -S "sk-proj--cbh8PBPLPk" >nul 2>&1
if %errorlevel% == 0 (
    echo [НАЙДЕНО] OpenAI API ключ найден в истории Git!
    echo.
) else (
    echo [OK] OpenAI API ключ не найден в истории Git через поиск.
    echo Но все равно выполним очистку для безопасности.
    echo.
)

REM Проверяем наличие ключа в текущих файлах
echo [ШАГ 2/5] Проверка наличия ключа в текущих файлах...
findstr /S /I /C:"sk-proj--cbh8PBPLPk" *.* >nul 2>&1
if %errorlevel% == 0 (
    echo [ВНИМАНИЕ] Ключ найден в текущих файлах!
    echo Убедись, что backend/.env в .gitignore
    echo.
) else (
    echo [OK] Ключ не найден в текущих файлах (кроме .env, который должен быть в .gitignore)
    echo.
)

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
    exit /b 1
)

set FILTER_BRANCH_SQUELCH_WARNING=1

echo.
echo [ШАГ 3/5] Удаление OpenAI API ключей из всех файлов истории...
"C:\Program Files\Git\bin\bash.exe" -c "git filter-branch -f --tree-filter 'find . -type f \( -name \"*.env\" -o -name \"*.md\" -o -name \"*.go\" -o -name \"*.ts\" -o -name \"*.tsx\" -o -name \"*.js\" -o -name \"*.jsx\" -o -name \"*.bat\" -o -name \"*.sh\" \) ! -path \"*/.git/*\" ! -path \"*/node_modules/*\" ! -path \"*/.next/*\" | while read f; do if [ -f \"$f\" ]; then sed -i.bak \"s/sk-proj-[A-Za-z0-9_-]\{48,\}/REMOVED_OPENAI_API_KEY/g\" \"$f\" 2>/dev/null; sed -i.bak \"s/OPENAI_API_KEY=sk-[A-Za-z0-9_-]\{48,\}/OPENAI_API_KEY=REMOVED/g\" \"$f\" 2>/dev/null; sed -i.bak \"s/sk-[A-Za-z0-9_-]\{48,\}/REMOVED_OPENAI_API_KEY/g\" \"$f\" 2>/dev/null; rm -f \"$f.bak\" 2>/dev/null; fi; done' --prune-empty --tag-name-filter cat -- --all"

if %errorlevel% neq 0 (
    echo [ОШИБКА] Не удалось выполнить filter-branch
    exit /b 1
)

echo.
echo [ШАГ 4/5] Удаление всех .env файлов из истории...
"C:\Program Files\Git\bin\bash.exe" -c "git filter-branch -f --tree-filter 'find . -type f -name \".env\" ! -path \"*/.git/*\" -delete 2>/dev/null || true' --prune-empty --tag-name-filter cat -- --all"

echo.
echo [ШАГ 5/5] Очистка ссылок и оптимизация репозитория...
"C:\Program Files\Git\bin\bash.exe" -c "git for-each-ref --format='delete %%(refname)' refs/original | git update-ref --stdin"
"C:\Program Files\Git\bin\bash.exe" -c "git reflog expire --expire=now --all"
"C:\Program Files\Git\bin\bash.exe" -c "git gc --prune=now --aggressive"

echo.
echo ========================================
echo ОЧИСТКА ЗАВЕРШЕНА
echo ========================================
echo.
echo [ПРОВЕРКА] Проверяем, что ключ удален...
git log --all -p -S "sk-proj--cbh8PBPLPk" >nul 2>&1
if %errorlevel% == 0 (
    echo [ВНИМАНИЕ] Ключ все еще найден в истории! Возможно, нужна дополнительная очистка.
) else (
    echo [OK] Ключ не найден в истории после очистки.
)
echo.

echo СЛЕДУЮЩИЕ ШАГИ:
echo 1. Проверь изменения: git log --all --oneline | head -10
echo 2. Если все ОК, принудительно запушь: git push --force --all
echo 3. Принудительно запушь теги: git push --force --tags
echo.
echo ВНИМАНИЕ: --force перезапишет удаленную историю!
echo Убедись, что все коллабораторы знают об этом.
echo.
echo Также рекомендуется:
echo - Пересоздать OpenAI API ключ в OpenAI Dashboard
echo - Обновить ключ на сервере в .env файле
echo.
pause


