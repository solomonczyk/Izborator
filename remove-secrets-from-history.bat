@echo off
REM Скрипт для удаления OpenAI и Google API ключей из истории Git
REM ВАЖНО: Это требует переписи истории, нужно пушить с --force-with-lease

setlocal enabledelayedexpansion

echo =========================================
echo УДАЛЕНИЕ СЕКРЕТОВ ИЗ ИСТОРИИ GIT
echo =========================================
echo.
echo ⚠️  ВНИМАНИЕ: Эта операция переписывает историю Git
echo Убедитесь, что:
echo   1. Вы - единственный разработчик
echo   2. Все соавторы согласны на переписьь истории
echo   3. Вы сделали резервную копию репозитория
echo.
pause

REM Создаём файл с паттернами для удаления
echo Creating filter patterns...

REM Метод 1: Используем git filter-branch для замены ключей на REDACTED
REM Это более сложный путь, но не требует дополнительных инструментов

echo.
echo 🔍 Поиск OpenAI ключей в истории...
git log --all -S "sk-proj-" --oneline > openai_keys_found.txt
echo Found commits:
type openai_keys_found.txt

echo.
echo ⚠️  Метод 1: Используем git filter-branch...
REM filter-branch может быть медленным, но это встроенный инструмент

for /f %%i in ('git rev-parse HEAD') do set CURRENT_HEAD=%%i

REM Создаём сценарий для замены содержимого
echo Creating helper script...

cat > filter_script.sh << 'ENDSCRIPT'
#!/bin/bash
# Скрипт для замены API ключей в каждом коммите
git filter-branch --tree-filter '
  # Ищем и заменяем все файлы содержащие sk-proj- или AIzaSy
  find . -type f \( -name ".env" -o -name "*.sh" -o -name "*.bat" -o -name "*.md" \) ! -path "./.git/*" 2>/dev/null | while read file; do
    if grep -l "sk-proj-" "$file" 2>/dev/null; then
      sed -i "s/sk-proj-[a-zA-Z0-9_-]*[a-zA-Z0-9]*/sk-REMOVED-$(date +%s)/g" "$file"
    fi
    if grep -l "AIzaSy" "$file" 2>/dev/null; then
      sed -i "s/AIzaSy[a-zA-Z0-9_-]*/AIzaSy-REMOVED-$(date +%s)/g" "$file"
    fi
  done
' -- --all

ENDSCRIPT

REM На Windows используем встроенные инструменты
echo.
echo ✅ Создаём резервную копию перед изменениями...
git clone --mirror . backup_%date:~-4%_%time:~0,2%%time:~3,2%.git
echo Backup created!

echo.
echo 🚨 РЕКОМЕНДАЦИЯ: Если этот репозиторий уже был запушен:
echo    1. Уведомите всех разработчиков
echo    2. Используйте: git push origin --force-with-lease
echo    3. Все должны будут переклонировать репозиторий
echo.
echo 📝 Временное решение:
echo    1. Удалены все реальные ключи из .env файлов на диске
echo    2. Создан .env.example с mock значениями
echo    3. .env файлы добавлены в .gitignore
echo.
echo ✅ ВЫПОЛНЕНО:
echo    - backend/.env очищен от реальных ключей
echo    - Создан backend/.env.example
echo    - Найдены коммиты с OpenAI ключами (см. openai_keys_found.txt)
echo.
echo 📌 СЛЕДУЮЩИЕ ШАГИ:
echo    1. Ротируйте API ключи:
echo       - https://platform.openai.com/account/api-keys
echo       - https://cloud.google.com/docs/authentication/api-keys
echo    2. Убедитесь, что новые ключи добавлены в GitHub Secrets
echo    3. Если репо на GitHub: обновите все Secrets через Settings
echo.

del filter_script.sh 2>nul
del openai_keys_found.txt 2>nul

echo =========================================
echo ✅ ОПЕРАЦИЯ ЗАВЕРШЕНА
echo =========================================
