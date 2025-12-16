# Nginx Reverse Proxy

## Настройка SSL

### Самоподписанный сертификат (для тестирования)

```bash
cd nginx
chmod +x generate-ssl.sh
./generate-ssl.sh
```

Это создаст:
- `ssl/cert.pem` - сертификат
- `ssl/key.pem` - приватный ключ

⚠️ **Важно:** Браузер будет показывать предупреждение о небезопасном соединении. Это нормально для самоподписанных сертификатов.

### Let's Encrypt (для продакшена с доменом)

Когда будет домен, можно использовать certbot:

```bash
# Установка certbot
apt-get update
apt-get install certbot python3-certbot-nginx

# Получение сертификата
certbot --nginx -d izborator.rs -d www.izborator.rs

# Автоматическое обновление
certbot renew --dry-run
```

## Конфигурация

- **HTTP (80)** → автоматический редирект на HTTPS
- **HTTPS (443)** → проксирование на backend и frontend
- **Backend API** → `http://backend:8080/api/`
- **Frontend** → `http://frontend:3000/`

## Порты

После настройки Nginx:
- ✅ `http://IP` → редирект на HTTPS
- ✅ `https://IP` → доступ к приложению
- ❌ `http://IP:8081` → больше не нужен (можно закрыть в firewall)
- ❌ `http://IP:3002` → больше не нужен (можно закрыть в firewall)

## Безопасность

Добавлены security headers:
- `Strict-Transport-Security` - принудительный HTTPS
- `X-Frame-Options` - защита от clickjacking
- `X-Content-Type-Options` - защита от MIME sniffing
- `X-XSS-Protection` - защита от XSS

