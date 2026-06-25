# Backend Tasks

## Epic: US-1 Создание упражнения

**Цель:** Реализовать API для получения списка упражнений и добавления новых кастомных упражнений пользователем.

### Задача 1: Инициализация сервера, конфигурации и инфраструктуры БД
* **Файлы:** `main.go`, `internal/config/config.go`, `internal/database/db.go`
* **Описание:** 1. Создать минимальный каркас Go-приложения (`main.go`).
    2. Реализовать пакет `internal/config` для безопасного чтения строки подключения к БД (DSN) и порта сервера из переменных окружения (без хардкода).
    3. В пакете `internal/database` реализовать подключение к PostgreSQL через пул соединений `github.com/jackc/pgx/v5/pgxpool`.
    4. Написать функцию `RunMigrations()`, которая при старте проверяет наличие таблиц и выполняет DDL-скрипт (создание таблиц `exercises`, `challenges`, `workouts`, `user_achievements` и индексов из `architecture.md`).
* **Ограничения:** Использовать чистый SQL. Сервер не должен падать, если база данных временно недоступна при старте (сделать 3 попытки переподключения).

### Задача 2: API Handlers
* **Файл(ы):** `internal/handlers/exercise_handler.go`, `internal/handlers/router.go`
* **Описание:**
  1. Реализовать `POST /api/exercises`.
     * Извлечь `X-User-Id` из заголовков (фоллбэк на `default_user_1`, если пусто).
     * Валидация: поле `name` не должно быть пустым. Если пусто — вернуть HTTP `400 Bad Request`.
     * Вызов `CreateExercise`. При успешном создании установить `is_custom = true` и вернуть HTTP `201 Created` с JSON созданного объекта. Если дубликат — вернуть `409 Conflict`.
  2. Реализовать `GET /api/exercises`.
     * Извлечь `X-User-Id`.
     * Вызов `GetExercises` и возврат массива JSON.
* **Ограничения:**
  * Писать подробные логи ошибок на сервере.

---

## Epic: US-2 Создание челленджа

**Цель:** Реализовать API для создания и получения списка челленджей.

### Задача 3: API Создания и Получения Челленджей
* **Файлы:** `internal/models/challenge.go`, `internal/database/challenge.go`, `internal/handlers/challenge_handler.go`, `internal/handlers/router.go`
* **Описание:**
  1. Создать структуру `Challenge` в `internal/models/challenge.go` в соответствии со спецификацией:
     * `id` (int), `user_id` (string), `name` (string), `exercise_id` (int), `target_value` (int), `current_progress` (int), `start_date` (time.Time / string), `end_date` (time.Time / string), `status` (string).
  2. В `internal/database/challenge.go` реализовать методы:
     * `CreateChallenge(ctx, userID, challenge)`: вставить запись с дефолтным статусом `active`.
     * `GetChallenges(ctx, userID)`: получить список всех челленджей пользователя.
     * `GetChallengeByID(ctx, userID, id)`: получить детальную информацию о челлендже.
  3. В `internal/handlers/challenge_handler.go` реализовать эндпоинты:
     * `POST /api/challenges` — Создание челленджа.
       * Валидация: `name` не пустое, `target_value > 0`, `end_date >= start_date`. При нарушении — вернуть `400 Bad Request`.
     * `GET /api/challenges` — Получить список челленджей пользователя (считывать `X-User-Id` с заголовка).
     * `GET /api/challenges/:id` — Детальная информация по конкретному челленджу.
  4. Зарегистрировать эндпоинты в `internal/handlers/router.go`.
* **Ограничения:**
  * Использовать чистый SQL.
  * Писать логи ошибок при операциях с БД.
