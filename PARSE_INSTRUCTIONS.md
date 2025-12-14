# ⚠️ ВАЖНО: Нужны прямые URL товаров!

## Проблема

Ты дал мне страницы **категорий** (списки товаров), а парсер работает только с URL **конкретных товаров**.

---

## Что нужно сделать

### 1. Для мобильных телефонов

**Страница категории:** https://gigatron.rs/mobilni-telefoni-tableti-i-oprema/mobilni-telefoni?Brend=Motorola

**Действия:**
1. Открой эту страницу в браузере
2. Кликни на **любой товар** (например, "MOTOROLA moto g15 Power 8/256GB Gravity Grey")
3. Скопируй URL из адресной строки
4. **Пример правильного URL:** `https://gigatron.rs/mobilni-telefoni-tableti-i-oprema/mobilni-telefoni/motorola-moto-g15-power-8-256gb-gravity-grey-XXXXX`

---

### 2. Для ноутбуков

**Страница категории:** https://gigatron.rs/racunari-i-komponente/laptop-racunari/poslovni-korisnici

**Действия:**
1. Открой эту страницу в браузере
2. Кликни на **любой товар** (например, "HP ZBook Power 16 G11")
3. Скопируй URL из адресной строки
4. **Пример правильного URL:** `https://gigatron.rs/racunari-i-komponente/laptop-racunari/hp-zbook-power-16-g11-XXXXX`

---

### 3. Для телевизоров

**Страница категории:** https://gigatron.rs/tv-audio-video/televizori

**Действия:**
1. Открой эту страницу в браузере
2. Кликни на **любой товар**
3. Скопируй URL из адресной строки
4. **Пример правильного URL:** `https://gigatron.rs/tv-audio-video/televizori/[название-телевизора]`

---

## После получения URL

Как только ты дашь мне **три прямых URL товаров**, я запущу парсер:

```powershell
# Для каждого товара:
go run cmd/worker/main.go -url "ТВОЙ_URL_ТОВАРА" -shop "shop-001"
```

---

## Альтернатива: Я могу попробовать найти URL сам

Если хочешь, я могу попробовать найти товары на страницах через браузер, но это займет больше времени и может не сработать.

**Что предпочитаешь:**
1. Ты найдешь URL сам (быстрее и надежнее)
2. Я попробую найти через браузер (может занять время)

