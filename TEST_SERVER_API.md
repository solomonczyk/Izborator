# 🧪 Тестирование API на сервере

## Информация о сервере
- **IP:** 152.53.227.37
- **Hostname:** v2202508292476370494.powersrv.de
- **Проект:** Запущен через Docker Compose

## Варианты тестирования

### Вариант 1: Тестирование через SSH на сервере (рекомендуется)

```bash
# Подключись к серверу
ssh root@152.53.227.37

# Перейди в директорию проекта
cd ~/Izborator

# Проверь статус контейнеров
docker-compose ps

# Запусти тестирование внутри backend контейнера
docker-compose exec backend bash -c "curl -s http://backend:8080/api/health"

# Или используй скрипт
docker-compose exec backend bash test-browse-api-server.sh
```

### Вариант 2: Тестирование с локальной машины

**PowerShell (Windows):**
```powershell
# Установи callback для игнорирования SSL ошибок
[System.Net.ServicePointManager]::ServerCertificateValidationCallback = {$true}

# Тест 1: Health check
Invoke-WebRequest -Uri "https://152.53.227.37/api/health" -UseBasicParsing

# Тест 2: Browse без фильтра
Invoke-WebRequest -Uri "https://152.53.227.37/api/v1/products/browse?page=1&per_page=5" -UseBasicParsing

# Тест 3: Browse с категорией
Invoke-WebRequest -Uri "https://152.53.227.37/api/v1/products/browse?category=mobilni-telefoni&page=1&per_page=5" -UseBasicParsing

# Или используй скрипт (после обновления)
.\test-browse-api-server.ps1
```

**Bash (Linux/Mac):**
```bash
# Тест 1: Health check
curl -k https://152.53.227.37/api/health

# Тест 2: Browse без фильтра
curl -k "https://152.53.227.37/api/v1/products/browse?page=1&per_page=5"

# Тест 3: Browse с категорией
curl -k "https://152.53.227.37/api/v1/products/browse?category=mobilni-telefoni&page=1&per_page=5"

# Или используй скрипт
bash test-browse-api-server.sh
```

### Вариант 3: Тестирование через браузер

Открой в браузере (игнорируй предупреждение о сертификате):
- Health: https://152.53.227.37/api/health
- Browse: https://152.53.227.37/api/v1/products/browse?page=1&per_page=5
- Browse с категорией: https://152.53.227.37/api/v1/products/browse?category=mobilni-telefoni&page=1&per_page=5

## Чек-лист тестирования

- [ ] Health check возвращает `{"status":"ok"}`
- [ ] GET /api/v1/products/browse (без фильтра) возвращает JSON с полями: items, total, page, per_page
- [ ] GET /api/v1/products/browse?category=mobilni-telefoni работает
- [ ] GET /api/v1/products/browse?category=laptopovi работает
- [ ] GET /api/v1/products/browse?category=neexistujuca-kategorija возвращает пустой результат или ошибку (fallback)

## После успешного тестирования

1. Обнови `ROADMAP_CURRENT_STEP.md` - отметь выполненные тесты
2. Обнови `DEVELOPMENT_LOG.md` - запиши результаты тестирования
3. Переходи к следующей задаче из `PLAN.md`

