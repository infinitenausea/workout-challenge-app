# Kanban Board

## 🚀 Спринт №7 (Завершен)
**Цель спринта:** Интеграция с Telegram Mini Apps (TMA) - безопасная авторизация, адаптация темы и нативный UX.

---

### TO DO

---

### IN PROGRESS

---

### QA / REVIEW

---

### DONE

- [x] **QA-8: Тестирование интеграции Telegram Mini Apps (Epic: US-9, US-10)**
  * **API Тесты:**
    - TC-9.1 (Успешная валидация `initData` по тестовому токену).
    - TC-9.2 (Проверка на HTTP 401 при неверном хэше).
    - TC-9.3 (Успешная отработка Dev Bypass при отсутствии переменной `TELEGRAM_BOT_TOKEN`).
  * **UI/UX Тесты:**
    - TC-10.1 (Проверка Network Requests на наличие корректного заголовка `Authorization` или `X-User-Id`).
    - TC-10.2 (Смена темы в клиенте Telegram автоматически перекрашивает элементы приложения).

- [x] **FE-19: Адаптация CSS под темы Telegram (Задача 18)**
  * **Файлы:** `frontend/css/main.css`
  * **Описание:**
    1. Заменить жесткие цвета на fallback-переменные для `--tg-theme-...`.
    2. Убедиться, что интерфейс моментально реагирует на событие `themeChanged` в Telegram.
  * **Ограничения:** Тема должна оставаться адекватной (читабельной) при локальной разработке без Telegram-окружения.

- [x] **FE-18: Инициализация Telegram SDK (Задача 17)**
  * **Файлы:** `frontend/index.html`, `frontend/js/telegram.js` (NEW), `frontend/js/api.js`, `frontend/js/app.js`
  * **Описание:**
    1. Добавить скрипт `<script src="https://telegram.org/js/telegram-web-app.js"></script>` в `index.html`.
    2. Создать `js/telegram.js` для инкапсуляции работы с `window.Telegram.WebApp`.
    3. Вызывать `Telegram.WebApp.ready()` and `Telegram.WebApp.expand()` при старте в `app.js`.
    4. Обновить `api.js`: считывать `window.Telegram.WebApp.initData` и передавать `Authorization: Bearer <initData>`. Если пусто — использовать `X-User-Id: default_user_1` (fallback).
  * **Ограничения:** Избегать крашей при запуске вне Telegram.

- [x] **BE-15: Создание TelegramAuthMiddleware (Задача 13)**
  * **Файлы:** `internal/handlers/middleware.go` (NEW), `internal/handlers/router.go`
  * **Описание:**
    1. Реализовать `TelegramAuthMiddleware(next http.Handler, botToken string) http.Handler`.
    2. Проверять наличие заголовка `Authorization: Bearer <initData>`.
    3. При пустом токене бота (fallback-режим) считывать `X-User-Id` и пропускать запрос.
    4. При заданном токене: парсить `initData`, сверять `hash` через `HMAC-SHA256("WebAppData", TELEGRAM_BOT_TOKEN)`, проверять `auth_date` (< 24 часов).
    5. При успехе извлекать `user.id` и писать его в `X-User-Id` / `context`. При ошибке отдавать `401 Unauthorized`.
    6. Обернуть обработчики `/api/*` в `router.go` в этот Middleware.
  * **Ограничения:** Алгоритм должен строго соответствовать докам Telegram. Логировать 401 ошибки без утечки токена.

---
