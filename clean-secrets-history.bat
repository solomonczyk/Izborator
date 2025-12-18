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
echo [2/4] Удаление всех Google API ключей из всех файлов...
"C:\Program Files\Git\bin\bash.exe" -c "git filter-branch -f --tree-filter 'find . -type f -name \"*.md\" -o -name \"*.go\" -o -name \"*.ts\" -o -name \"*.tsx\" -o -name \"*.js\" -o -name \"*.jsx\" | while read f; do if [ -f \"$f\" ]; then sed -i.bak \"s/AIzaSy[A-Za-z0-9_-]\{35\}/REMOVED_GOOGLE_API_KEY/g\" \"$f\" 2>/dev/null; sed -i.bak \"s/f0fa9f5df0f5a4522/REMOVED_CX_ID/g\" \"$f\" 2>/dev/null; rm -f \"$f.bak\" 2>/dev/null; fi; done' --prune-empty --tag-name-filter cat -- --all"

echo.
echo [3/5] Удаление OpenAI API ключей из всех файлов...
"C:\Program Files\Git\bin\bash.exe" -c "git filter-branch -f --tree-filter 'find . -type f \( -name \"*.env\" -o -name \"*.md\" -o -name \"*.go\" -o -name \"*.ts\" -o -name \"*.tsx\" -o -name \"*.js\" -o -name \"*.jsx\" -o -name \"*.bat\" -o -name \"*.sh\" \) ! -path \"*/.git/*\" ! -path \"*/node_modules/*\" ! -path \"*/.next/*\" | while read f; do if [ -f \"$f\" ]; then sed -i.bak \"s/sk-proj-[A-Za-z0-9_-]\{48,\}/REMOVED_OPENAI_API_KEY/g\" \"$f\" 2>/dev/null; sed -i.bak \"s/OPENAI_API_KEY=sk-[A-Za-z0-9_-]\{48,\}/OPENAI_API_KEY=REMOVED/g\" \"$f\" 2>/dev/null; sed -i.bak \"s/sk-[A-Za-z0-9_-]\{48,\}/REMOVED_OPENAI_API_KEY/g\" \"$f\" 2>/dev/null; rm -f \"$f.bak\" 2>/dev/null; fi; done' --prune-empty --tag-name-filter cat -- --all"

echo.
echo [4/5] Удаление всех .env файлов из истории...
"C:\Program Files\Git\bin\bash.exe" -c "git filter-branch -f --tree-filter 'find . -type f -name \".env\" ! -path \"*/.git/*\" -delete 2>/dev/null || true' --prune-empty --tag-name-filter cat -- --all"

echo.
echo [5/5] Очистка ссылок и оптимизация репозитория...
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

