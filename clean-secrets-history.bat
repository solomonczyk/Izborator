@echo off
REM Скрипт для очистки истории Git от всех секретов и ключей
REM ВНИМАНИЕ: Это перезапишет всю историю Git! Используй только если уверен.

echo ========================================
echo ОЧИСТКА ИСТОРИИ GIT ОТ СЕКРЕТОВ
echo ========================================
echo.
echo ВНИМАНИЕ: Этот скрипт перезапишет всю историю Git!
echo Убедись, что:
echo 1. Все изменения закоммичены
echo 2. Создана резервная копия репозитория
echo 3. Все коллабораторы уведомлены
echo.
pause

set FILTER_BRANCH_SQUELCH_WARNING=1

echo.
echo [1/3] Удаление Google API ключей из DEVELOPMENT_LOG.md...
"C:\Program Files\Git\bin\bash.exe" -c "git filter-branch -f --tree-filter 'if [ -f DEVELOPMENT_LOG.md ]; then sed -i.bak \"s/AIzaSy[A-Za-z0-9_-]\{35\}/REMOVED_GOOGLE_API_KEY/g\" DEVELOPMENT_LOG.md 2>/dev/null; sed -i.bak \"s/f0fa9f5df0f5a4522/REMOVED_CX_ID/g\" DEVELOPMENT_LOG.md 2>/dev/null; rm -f DEVELOPMENT_LOG.md.bak 2>/dev/null; fi' --prune-empty --tag-name-filter cat -- --all"

echo.
echo [2/3] Удаление всех Google API ключей из всех файлов...
"C:\Program Files\Git\bin\bash.exe" -c "git filter-branch -f --tree-filter 'find . -type f -name \"*.md\" -o -name \"*.go\" -o -name \"*.ts\" -o -name \"*.tsx\" -o -name \"*.js\" -o -name \"*.jsx\" | while read f; do if [ -f \"$f\" ]; then sed -i.bak \"s/AIzaSy[A-Za-z0-9_-]\{35\}/REMOVED_GOOGLE_API_KEY/g\" \"$f\" 2>/dev/null; sed -i.bak \"s/f0fa9f5df0f5a4522/REMOVED_CX_ID/g\" \"$f\" 2>/dev/null; rm -f \"$f.bak\" 2>/dev/null; fi; done' --prune-empty --tag-name-filter cat -- --all"

echo.
echo [3/3] Очистка ссылок и оптимизация репозитория...
"C:\Program Files\Git\bin\bash.exe" -c "git for-each-ref --format='delete %%(refname)' refs/original | git update-ref --stdin"
"C:\Program Files\Git\bin\bash.exe" -c "git reflog expire --expire=now --all"
"C:\Program Files\Git\bin\bash.exe" -c "git gc --prune=now --aggressive"

echo.
echo ========================================
echo ОЧИСТКА ЗАВЕРШЕНА
echo ========================================
echo.
echo СЛЕДУЮЩИЕ ШАГИ:
echo 1. Проверь изменения: git log --all
echo 2. Если все ОК, принудительно запушь: git push --force --all
echo 3. Принудительно запушь теги: git push --force --tags
echo.
echo ВНИМАНИЕ: --force перезапишет удаленную историю!
echo Убедись, что все коллабораторы знают об этом.
echo.
pause

