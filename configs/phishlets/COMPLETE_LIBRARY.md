# 🎯 PhantomProxy Phishlets - Полная библиотека

## 📦 Статистика

- **Всего phishlets**: 21
- **Российские сервисы**: 8
- **Международные сервисы**: 13
- **Универсальных шаблонов**: 1

---

## 🌍 Международные сервисы (13)

### Социальные сети и мессенджеры

| # | Phishlet | Цели | Файл |
|---|----------|------|------|
| 1 | **Facebook** | Facebook, Instagram, WhatsApp | `facebook.yaml` |
| 2 | **Instagram** | Instagram, Threads | `instagram.yaml` |
| 3 | **TikTok** | TikTok, TikTok Ads | `tiktok.yaml` |
| 4 | **Telegram** | Telegram Web | `telegram.yaml` |

### Почтовые сервисы

| # | Phishlet | Цели | Файл |
|---|----------|------|------|
| 5 | **Microsoft 365** | Outlook, OneDrive, Teams, SharePoint, Azure AD | `microsoft365.yaml` |
| 6 | **Google Workspace** | Gmail, Drive, Docs, Calendar, Meet | `googleworkspace.yaml` |

### Маркетплейсы

| # | Phishlet | Цели | Файл |
|---|----------|------|------|
| 7 | **Amazon** | Amazon.com, AWS Console | `amazon.yaml` |
| 8 | **eBay** | eBay.com, eBay Kleinanzeigen | `ebay.yaml` |

### Стриминговые сервисы

| # | Phishlet | Цели | Файл |
|---|----------|------|------|
| 9 | **Netflix** | Netflix.com | `netflix.yaml` |
| 10 | **Disney+** | Disneyplus.com | `disney.yaml` |
| 11 | **Spotify** | Spotify.com | `spotify.yaml` |

### Платежные системы

| # | Phishlet | Цели | Файл |
|---|----------|------|------|
| 12 | **PayPal** | PayPal.com, PayPal Business | `paypal.yaml` |
| 13 | **Stripe** | Stripe.com, Stripe Dashboard | `stripe.yaml` |

---

## 🇷🇺 Российские сервисы (8)

### Почтовые сервисы

| # | Phishlet | Цели | Файл |
|---|----------|------|------|
| 1 | **Yandex** | Яндекс.Паспорт, Почта, Диск, Деньги, Метрика | `yandex.yaml` |
| 2 | **Mail.ru** | Почта Mail.ru, Облако, ICQ, Мой Мир | `mailru.yaml` |

### Социальные сети

| # | Phishlet | Цели | Файл |
|---|----------|------|------|
| 3 | **VK** | ВКонтакте, VK Play, VK Pay | `vk.yaml` |

### Банки

| # | Phishlet | Цели | Файл |
|---|----------|------|------|
| 4 | **SberBank** | СберБанк Онлайн, СберБизнес | `sberbank.yaml` |
| 5 | **Tinkoff** | Тинькофф, Тинькофф Бизнес | `tinkoff_business.yaml` |

### Маркетплейсы

| # | Phishlet | Цели | Файл |
|---|----------|------|------|
| 6 | **Ozon** | Ozon.ru, Ozon Seller | `ozon.yaml` |
| 7 | **Wildberries** | Wildberries.ru, WB Seller | `wildberries.yaml` |

### Госуслуги

| # | Phishlet | Цели | Файл |
|---|----------|------|------|
| 8 | **Gosuslugi** | Госуслуги.ру | `gosuslugi.yaml` |

---

## 🔧 Универсальные шаблоны (1)

| # | Шаблон | Назначение | Файл |
|---|--------|------------|------|
| 1 | **Universal Template** | Клонирование ЛЮБОГО сайта | `universal_template.yaml` |

---

## 🎯 Функционал каждого phishlet

### Перехват данных

✅ **Логин/пароль**
- Перехват в реальном времени
- Отправка на сервер при вводе
- Debounce для оптимизации

✅ **2FA/MFA коды**
- SMS коды
- TOTP коды
- Push подтверждения
- Timeout на ввод

✅ **Сессионные данные**
- Cookies (все типы)
- LocalStorage
- SessionStorage
- Access токены

✅ **Платежные данные** (для маркетплейсов)
- Номер карты
- Срок действия
- CVV/CVC
- Имя держателя

---

## 🛡️ Anti-Detection

Каждый phishlet включает:

| Функция | Описание |
|---------|----------|
| **Hide Webdriver** | Скрытие navigator.webdriver |
| **Spoof Navigator** | Подмена параметров браузера |
| **Anti DevTools** | Обход обнаружения DevTools |
| **Random Timings** | Рандомизация таймингов |
| **Canvas Spoof** | Canvas fingerprint spoofing |
| **WebGL Spoof** | WebGL fingerprint spoofing |
| **Audio Spoof** | AudioContext spoofing |
| **Timezone Spoof** | Подмена часового пояса |
| **Language Spoof** | Подмена языка |

---

## ⚡ Производительность

Оптимизация во всех phishlets:

| Оптимизация | Описание |
|-------------|----------|
| **Cache Static** | Кэширование статики |
| **Compress Responses** | Brotli/Gzip сжатие |
| **Lazy Images** | Lazy loading изображений |
| **Minify JS/CSS** | Минификация кода |
| **HTTP/2 Push** | Предзагрузка ресурсов |
| **TLS 1.3** | Современное шифрование |

---

## 🚀 Быстрый старт

### 1. Активация phishlet

```bash
curl -X POST http://localhost:8080/api/v1/phishlets/{id}/enable \
  -H "X-API-Key: your-api-key"
```

### 2. Просмотр списка

```bash
curl http://localhost:8080/api/v1/phishlets \
  -H "X-API-Key: your-api-key"
```

### 3. Через веб-панель

1. Откройте http://localhost:8080
2. Перейдите в раздел "Phishlets"
3. Нажмите кнопку включения

---

## 📊 Веб-панель управления

### Разделы

| Раздел | Описание |
|--------|----------|
| **Dashboard** | Статистика и графики в реальном времени |
| **Phishlets** | Управление всеми phishlets |
| **Сессии** | Активные сессии жертв |
| **Credentials** | Перехваченные логины/пароли |
| **Логи** | Live логи системы |
| **Настройки** | Конфигурация системы |

### Функции

✅ **Real-time мониторинг**
- Активные сессии
- Перехваченные credentials
- Запросы к прокси
- Статус сервисов

✅ **Управление phishlets**
- Включение/выключение
- Просмотр деталей
- Статус каждого

✅ **Экспорт данных**
- CSV экспорт credentials
- JSON экспорт сессий
- Логи в текстовом формате

✅ **SSE уведомления**
- Мгновенные уведомления о новых credentials
- Уведомления о создании сессий
- Live логи

---

## 🎨 UI/UX Особенности

### Тёмная тема
- Профессиональный дизайн
- Снижение нагрузки на глаза
- Modern look & feel

### Адаптивность
- Desktop (1920x1080+)
- Tablet (768x1024)
- Mobile (375x667)

### Интерактивность
- Hover эффекты
- Анимации переходов
- Loading индикаторы
- Toast уведомления

---

## 📱 Горячие клавиши

| Клавиша | Действие |
|---------|----------|
| `R` | Refresh all data |
| `D` | Перейти на Dashboard |
| `P` | Перейти к Phishlets |
| `S` | Перейти к Sessions |
| `C` | Перейти к Credentials |
| `L` | Перейти к Logs |
| `N` | Перейти к Settings |

---

## 🔐 Безопасность

### Рекомендации

1. ⚠️ **Смените API Key** по умолчанию
2. 🔒 **Используйте HTTPS** для панели
3. 🌐 **Ограничьте доступ** по IP
4. 🔑 **Включите 2FA** (в разработке)
5. 📝 **Логируйте доступ** к панели

---

## 🛠️ Кастомизация

### Создание нового phishlet

1. Скопируйте `universal_template.yaml`
2. Измените `target.primary`
3. Настройте `sub_filters`
4. Добавьте нужные `triggers`
5. Укажите `cookies` для перехвата

### Изменение UI

1. Откройте `frontend/js/panel.js`
2. Измените CSS variables
3. Добавьте новые виджеты
4. Настройте графики Chart.js

---

## 📞 Поддержка

- **Документация**: `./docs/`
- **API Docs**: http://localhost:8080/api/docs
- **Issues**: https://github.com/phantom-proxy/phantom-proxy/issues
- **Telegram**: @phantomproxy

---

## 📈 Roadmap

### Q2 2026
- [ ] LinkedIn phishlet
- [ ] Twitter/X phishlet
- [ ] Snapchat phishlet
- [ ] Discord phishlet

### Q3 2026
- [ ] Apple iCloud phishlet
- [ ] Dropbox phishlet
- [ ] Slack phishlet
- [ ] Zoom phishlet

### Q4 2026
- [ ] AI-powered content generation
- [ ] Auto-phishlet creator
- [ ] Advanced analytics
- [ ] Multi-language support

---

**Версия библиотеки**: 2.0  
**Phishlets**: 21  
**Веб-панель**: 1.0.0  
**Последнее обновление**: Февраль 2026
