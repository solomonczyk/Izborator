# Миграции базы данных

Этот каталог содержит SQL миграции для базы данных PostgreSQL.

## Структура

- `*.up.sql` - миграция вверх (применение изменений)
- `*.down.sql` - миграция вниз (откат изменений)

## Применение миграций

### Используя встроенный инструмент

```bash
# Перейти в директорию backend
cd backend

# Применить все неприменённые миграции
go run cmd/migrate/main.go -up

# Показать статус миграций
go run cmd/migrate/main.go -status

# Показать текущую версию
go run cmd/migrate/main.go -version

# Откатить N последних миграций
go run cmd/migrate/main.go -down 1
```

### Сборка инструмента

```bash
cd backend
go build -o migrate.exe cmd/migrate/main.go

# Использование
./migrate.exe -up
./migrate.exe -status
./migrate.exe -down 1
```

### Используя migrate CLI (альтернатива)

```bash
# Установка migrate
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Применение миграций
migrate -path ./migrations -database "postgres://user:password@localhost/izborator?sslmode=disable" up

# Откат последней миграции
migrate -path ./migrations -database "postgres://user:password@localhost/izborator?sslmode=disable" down 1
```

## Порядок миграций

Миграции применяются в порядке их номеров:
- `0001_initial_schema` - создание начальной схемы БД

## Формат имён файлов

Миграции должны следовать формату:
- `NNNN_name.up.sql` - применение миграции
- `NNNN_name.down.sql` - откат миграции

Где `NNNN` - четырёхзначный номер версии (0001, 0002, и т.д.)

