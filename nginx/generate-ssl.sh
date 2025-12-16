#!/bin/bash
# Генерация самоподписанного SSL сертификата для тестирования

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SSL_DIR="$SCRIPT_DIR/ssl"

mkdir -p "$SSL_DIR"

# Генерируем приватный ключ
openssl genrsa -out "$SSL_DIR/key.pem" 2048

# Генерируем самоподписанный сертификат (действителен 365 дней)
openssl req -new -x509 -key "$SSL_DIR/key.pem" -out "$SSL_DIR/cert.pem" -days 365 -subj "/C=RS/ST=Serbia/L=Belgrade/O=Izborator/CN=localhost"

# Устанавливаем правильные права доступа
chmod 600 "$SSL_DIR/key.pem"
chmod 644 "$SSL_DIR/cert.pem"

echo "✅ SSL сертификат создан:"
echo "   - $SSL_DIR/cert.pem"
echo "   - $SSL_DIR/key.pem"
echo ""
echo "⚠️  Это самоподписанный сертификат. Браузер будет показывать предупреждение."
echo "   Для продакшена используй Let's Encrypt с реальным доменом."

