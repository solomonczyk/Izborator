FROM nginx:alpine

# Копируем конфигурацию
COPY nginx.conf /etc/nginx/nginx.conf

# Создаем директорию для SSL сертификатов
# Сертификаты монтируются через volume в docker-compose.yml
RUN mkdir -p /etc/nginx/ssl

# Открываем порты
EXPOSE 80 443

CMD ["nginx", "-g", "daemon off;"]

