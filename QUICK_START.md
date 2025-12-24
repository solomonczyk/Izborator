# 🚀 QUICK START: С ЧЕГО НАЧАТЬ?

**Время чтения:** 5 минут  
**Для кого:** Разработчики, которые хотят срочно начать улучшение проекта  

---

## 📍 ВЫ НАХОДИТЕСЬ ЗДЕСЬ

Проект `Izborator` содержит 77 файлов в корне и требует организации. Анализ выявил **13 основных проблем**, из которых **4 критичны**.

---

## 🎯 ПЛАН НА НЕДЕЛЮ (16 часов)

### ✅ Результат после этой недели:
- Корень проекта чист (было 77 → будет 10 файлов)
- Все скрипты организованы в папке `scripts/`
- Документация полная и актуальная
- Код без дублирования в handlers
- Проект готов к следующему этапу улучшений

---

## 📋 ДЕНЬ 1 (4 часа) - ОРГАНИЗАЦИЯ СКРИПТОВ

### Утро (2 часа)

**Шаг 1: Создать папку scripts/**
```bash
cd f:/Dev/Projects/Izborator
mkdir -p scripts/{setup,start,stop,check,test,fix,deploy,cleanup}
```

**Шаг 2: Переместить скрипты**
```bash
# Проверка
mv check-*.sh scripts/check/

# Исправление
mv fix-*.sh scripts/fix/

# Тестирование
mv test-*.sh scripts/test/
mv quick-test-api.sh scripts/test/

# Запуск
mv start-*.sh scripts/start/
mv start-*.bat scripts/start/
mv start-*.ps1 scripts/start/

# Остановка
mv stop-*.ps1 scripts/stop/

# Деплой
mv deploy.sh scripts/deploy/

# Очистка
mv remove-*.sh scripts/cleanup/
mv remove-*.bat scripts/cleanup/
mv clean-*.bat scripts/cleanup/
mv update-openai-key.sh scripts/cleanup/

# Остальное
mv run-*.sh scripts/
mv rebuild-and-test-classifier.sh scripts/test/
mv install-pre-commit-hook.bat scripts/setup/
mv fix-app-init-order.py scripts/fix/
```

**Проверка результата:**
```bash
# В корне больше не должно быть .sh, .bat, .ps1 файлов
ls -la *.sh *.bat *.ps1 2>/dev/null | wc -l
# Результат: 0 (никаких файлов)
```

---

### Полдень (2 часа)

**Шаг 3: Удалить .trigger файлы**
```bash
# Удалить из корня
rm .trigger-*

# ИЛИ переместить в .github
mkdir -p .github/workflows/triggers
mv .trigger-* .github/workflows/triggers/ 2>/dev/null || true
```

**Шаг 4: Обновить .gitignore**

Добавить в .gitignore (если еще нет):
```
# Environment
.env
.env.local
.env.*.local

# Node/Frontend
frontend/node_modules/
frontend/.next/

# Go/Backend
backend/bin/
backend/*.test

# IDE
.vscode/
.idea/

# Docker
/db_data/
/redis_data/

# Logs
*.log
/tmp/
```

---

## 📚 ДЕНЬ 2 (4 часа) - СОЗДАНИЕ ДОКУМЕНТАЦИИ

### Утро (4 часа) - Создать 4 файла в корне

**Готовые файлы находятся в:**
- `STRATEGY.md` — копировать из примера ниже
- `STATUS.md` — копировать из примера ниже
- `SECURITY_GUIDELINES.md` — копировать из примера ниже
- `START_COMMANDS.md` — копировать из примера ниже

Все файлы уже созданы! Проверьте что они есть в корне проекта.

---

## 🔧 ДЕНЬ 3-4 (8 часов) - РЕФАКТОРИНГ HANDLERS

### Утро (День 3)

**Шаг 5: Создать BaseHandler**

Создайте файл `backend/internal/http/handlers/base.go`:

```go
package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"izborator/internal/logger"
	"izborator/internal/storage"
)

type BaseHandler struct {
	Logger  logger.Logger
	Storage storage.Storage
}

func (h *BaseHandler) RespondJSON(w http.ResponseWriter, code int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(data)
}

func (h *BaseHandler) RespondError(w http.ResponseWriter, code int, message string) {
	h.RespondJSON(w, code, map[string]interface{}{
		"error": message,
	})
}

func (h *BaseHandler) ParseDaysParam(r *http.Request) int {
	days := r.URL.Query().Get("days")
	if days == "" {
		return 30
	}
	d, _ := strconv.Atoi(days)
	if d < 1 {
		return 1
	}
	if d > 365 {
		return 365
	}
	return d
}
```

### Полдень (День 3) - День 4

**Шаг 6: Обновить handlers**

Для каждого файла в `backend/internal/http/handlers/`:

```go
// ДО:
type ProductsHandler struct {
    logger logger.Logger
    storage storage.Storage
}

func (h *ProductsHandler) GetProduct(w http.ResponseWriter, r *http.Request) {
    // ... код ...
    respondJSON(w, 200, product)
}

// ПОСЛЕ:
type ProductsHandler struct {
    *BaseHandler
}

func (h *ProductsHandler) GetProduct(w http.ResponseWriter, r *http.Request) {
    // ... код ...
    h.RespondJSON(w, 200, product)
}
```

Файлы для обновления:
- [ ] `products.go`
- [ ] `categories.go`
- [ ] `cities.go`
- [ ] `stats.go`
- [ ] `scraper.go` (если есть)

**Шаг 7: Запустить тесты**
```bash
cd backend
go test ./internal/http/handlers -v
```

---

## 🔨 ДЕНЬ 5 - ВЕРСИИ И ФИНАЛИЗАЦИЯ

### Утро (2 часа)

**Шаг 8: Исправить версии NPM**

```bash
cd frontend

# Обновить package.json (вручную или через npm)
npm update next@15 react@18 react-dom@18

# Переустановить
rm -rf node_modules package-lock.json
npm install

# Проверить
npm run build
```

**Шаг 9: Добавить Go зависимости**

```bash
cd backend
go get github.com/stretchr/testify/assert
go get github.com/go-playground/validator/v10
go mod tidy
go test ./...
```

### Полдень (2 часа)

**Шаг 10: Обновить документацию**

1. Обновить `README.md` в корне:
   - Добавить ссылки на новые файлы (STRATEGY.md, STATUS.md, etc.)
   - Обновить раздел "Быстрые ссылки"

2. Обновить `docs/README.md`:
   - Исправить ссылки на файлы
   - Удалить ссылки на несуществующие файлы

3. Обновить `PROJECT_DEVELOPMENT_PLAN.md`:
   - Отметить в Этапе 1 что выполнено

---

## ✅ КОНТРОЛЬНЫЙ СПИСОК

Проверьте что все сделано:

### День 1: Организация ✓
- [ ] Создана папка `scripts/` с подпапками
- [ ] Все .sh скрипты перемещены в scripts/check, fix, test, etc.
- [ ] Все .bat и .ps1 файлы перемещены
- [ ] .trigger файлы удалены из корня
- [ ] .gitignore обновлен
- [ ] В корне больше нет скриптов

### День 2: Документация ✓
- [ ] Созданы STRATEGY.md, STATUS.md, SECURITY_GUIDELINES.md, START_COMMANDS.md
- [ ] README.md ссылается на все файлы
- [ ] docs/README.md обновлен

### День 3-4: Рефакторинг ✓
- [ ] Создан backend/internal/http/handlers/base.go
- [ ] Все handlers обновлены для использования BaseHandler
- [ ] Удален дублирующийся код
- [ ] go test ./internal/http/handlers проходит

### День 5: Версии ✓
- [ ] NPM версии обновлены и совпадают
- [ ] frontend npm run build работает
- [ ] Go зависимости добавлены
- [ ] go test ./... проходит

---

## 🎯 РЕЗУЛЬТАТ

После выполнения всех шагов:

```
Корень проекта:
БЫЛО: 77 файлов (3+, 12.bat, 12.ps1, 20.trigger, 26.md)
СТАЛО: ~10 файлов (README.md, docker-compose.yml, LICENSE, .git*, docs/, backend/, frontend/, scripts/, nginx/, config/)

Статус:
✅ Проект организован
✅ Документация полная
✅ Код без дублирования
✅ Готов к следующему этапу
```

---

## 📞 ЧТО ДАЛЬШЕ?

После завершения этой недели:

1. **Неделя 2-3:** Рефакторинг adapters и добавление Swagger
2. **Неделя 4+:** Расширение тестов и мониторинга

Подробный план в файле: **DEVELOPMENT_PLAN_DETAILED.md**

---

## 🔗 ПОЛЕЗНЫЕ ССЫЛКИ

- **Полный анализ:** DEEP_ANALYSIS_REPORT.md
- **Детальный план:** DEVELOPMENT_PLAN_DETAILED.md
- **Это резюме:** ANALYSIS_SUMMARY.md

---

**Удачи! Начните с Шага 1 прямо сейчас! 🚀**
