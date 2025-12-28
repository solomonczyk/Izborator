# Этап 1: Базовый образ
FROM node:20-alpine AS base

# Этап 2: Установка зависимостей
FROM base AS deps
WORKDIR /app
COPY package.json package-lock.json ./
RUN npm ci

# Этап 3: Сборка
FROM base AS builder
WORKDIR /app
COPY --from=deps /app/node_modules ./node_modules
COPY . .

# Отключаем телеметрию Next.js для скорости
ENV NEXT_TELEMETRY_DISABLED 1

# Переменные окружения для сборки (ARG для передачи из docker-compose)
ARG NEXT_PUBLIC_API_BASE
ENV NEXT_PUBLIC_API_BASE=${NEXT_PUBLIC_API_BASE}

# Собираем проект
RUN npm run build

# Этап 4: Продакшен образ
FROM base AS runner
WORKDIR /app

ENV NODE_ENV production
ENV NEXT_TELEMETRY_DISABLED 1

RUN addgroup --system --gid 1001 nodejs
RUN adduser --system --uid 1001 nextjs

# Копируем только нужное для запуска (Standalone mode)
COPY --from=builder /app/public ./public
COPY --from=builder --chown=nextjs:nodejs /app/.next/standalone ./
COPY --from=builder --chown=nextjs:nodejs /app/.next/static ./.next/static

USER nextjs

EXPOSE 3000

ENV PORT 3000
# ВАЖНО: Host должен быть 0.0.0.0 для Docker
ENV HOSTNAME "0.0.0.0"

CMD ["node", "server.js"]

