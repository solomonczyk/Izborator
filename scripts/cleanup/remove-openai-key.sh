#!/bin/bash
# Скрипт для удаления OpenAI API ключа из истории Git

set -e

echo "========================================"
echo "УДАЛЕНИЕ OPENAI API КЛЮЧА ИЗ ИСТОРИИ GIT"
echo "========================================"
echo ""

# Устанавливаем переменную для подавления предупреждений
export FILTER_BRANCH_SQUELCH_WARNING=1

echo "[ШАГ 1/5] Проверка наличия ключа в истории Git..."
if git log --all -p -S "sk-proj--cbh8PBPLPk" | head -5 | grep -q "sk-proj"; then
    echo "[НАЙДЕНО] OpenAI API ключ найден в истории Git!"
else
    echo "[OK] OpenAI API ключ не найден в истории через поиск."
    echo "Но все равно выполним очистку для безопасности."
fi
echo ""

echo "[ШАГ 2/5] Удаление OpenAI API ключей из всех файлов истории..."
git filter-branch -f --tree-filter '
  find . -type f \( -name "*.env" -o -name "*.md" -o -name "*.go" -o -name "*.ts" -o -name "*.tsx" -o -name "*.js" -o -name "*.jsx" -o -name "*.bat" -o -name "*.sh" \) \
    ! -path "*/.git/*" ! -path "*/node_modules/*" ! -path "*/.next/*" | \
  while read f; do
    if [ -f "$f" ]; then
      sed -i.bak "s/sk-proj-[A-Za-z0-9_-]\{48,\}/REMOVED_OPENAI_API_KEY/g" "$f" 2>/dev/null || true
      sed -i.bak "s/OPENAI_API_KEY=sk-[A-Za-z0-9_-]\{48,\}/OPENAI_API_KEY=REMOVED/g" "$f" 2>/dev/null || true
      sed -i.bak "s/sk-[A-Za-z0-9_-]\{48,\}/REMOVED_OPENAI_API_KEY/g" "$f" 2>/dev/null || true
      rm -f "$f.bak" 2>/dev/null || true
    fi
  done
' --prune-empty --tag-name-filter cat -- --all

echo ""
echo "[ШАГ 3/5] Удаление всех .env файлов из истории..."
git filter-branch -f --tree-filter 'find . -type f -name ".env" ! -path "*/.git/*" -delete 2>/dev/null || true' --prune-empty --tag-name-filter cat -- --all

echo ""
echo "[ШАГ 4/5] Очистка ссылок..."
git for-each-ref --format='delete %(refname)' refs/original | git update-ref --stdin || true

echo ""
echo "[ШАГ 5/5] Оптимизация репозитория..."
git reflog expire --expire=now --all
git gc --prune=now --aggressive

echo ""
echo "========================================"
echo "ОЧИСТКА ЗАВЕРШЕНА"
echo "========================================"
echo ""

echo "[ПРОВЕРКА] Проверяем, что ключ удален..."
if git log --all -p -S "sk-proj--cbh8PBPLPk" | head -5 | grep -q "sk-proj"; then
    echo "[ВНИМАНИЕ] Ключ все еще найден в истории! Возможно, нужна дополнительная очистка."
else
    echo "[OK] Ключ не найден в истории после очистки."
fi
echo ""

echo "СЛЕДУЮЩИЕ ШАГИ:"
echo "1. Проверь изменения: git log --all --oneline | head -10"
echo "2. Если все ОК, принудительно запушь: git push --force --all"
echo "3. Принудительно запушь теги: git push --force --tags"
echo ""
echo "ВНИМАНИЕ: --force перезапишет удаленную историю!"
echo "Убедись, что все коллабораторы знают об этом."
echo ""


