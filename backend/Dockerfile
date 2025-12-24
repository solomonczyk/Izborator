# Этап 1: Сборка (Builder)
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Кэшируем зависимости
COPY go.mod go.sum ./
RUN go mod download

# Копируем исходный код
COPY . .

# Компилируем статические бинарники
# CGO_ENABLED=0 делает бинарник полностью независимым от системных библиотек
RUN CGO_ENABLED=0 GOOS=linux go build -o api ./cmd/api/main.go
RUN CGO_ENABLED=0 GOOS=linux go build -o worker ./cmd/worker/main.go
RUN CGO_ENABLED=0 GOOS=linux go build -o migrate ./cmd/migrate/main.go
RUN CGO_ENABLED=0 GOOS=linux go build -o indexer ./cmd/indexer/main.go
RUN CGO_ENABLED=0 GOOS=linux go build -o classifier ./cmd/classifier/main.go
RUN CGO_ENABLED=0 GOOS=linux go build -o discovery ./cmd/discovery/main.go
RUN CGO_ENABLED=0 GOOS=linux go build -o autoconfig ./cmd/autoconfig/main.go

# Этап 2: Финальный образ (Runner)
FROM alpine:latest

WORKDIR /app

# Устанавливаем сертификаты (для HTTPS запросов к магазинам)
RUN apk --no-cache add ca-certificates tzdata wget

# Копируем бинарники из этапа сборки
COPY --from=builder /app/api .
COPY --from=builder /app/worker .
COPY --from=builder /app/migrate .
COPY --from=builder /app/indexer .
COPY --from=builder /app/classifier .
COPY --from=builder /app/discovery .
COPY --from=builder /app/autoconfig .

# Копируем миграции и локали (они нужны в рантайме)
COPY --from=builder /app/migrations ./migrations
COPY --from=builder /app/internal/i18n/locales ./internal/i18n/locales

# Открываем порт
EXPOSE 8080

# По умолчанию запускаем API (но можно переопределить в docker-compose)
CMD ["./api"]

