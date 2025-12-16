FROM nginx:alpine

# Копируем конфигурацию
COPY nginx.conf /etc/nginx/nginx.conf

# Создаем директорию для SSL сертификатов
RUN mkdir -p /etc/nginx/ssl

# Копируем SSL сертификаты (если есть)
COPY ssl/*.pem /etc/nginx/ssl/ 2>/dev/null || true

# Открываем порты
EXPOSE 80 443

CMD ["nginx", "-g", "daemon off;"]

