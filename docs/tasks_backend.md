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

## Epic: US-2, US-5 Создание челленджа и Дашборд

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

---

## Epic: US-3, US-4, US-6, US-7 Логирование тренировки, Детали, Удаление и Ачивки

**Цель:** Реализовать API для добавления и удаления тренировок (workouts) с транзакционным пересчётом прогресса челленджа, а также систему выдачи достижений (ачивок).

### Задача 4: Модель Workout и структуры ответа

* **Файлы:**
  * `internal/models/workout.go` **(NEW)**
  * `internal/models/achievement.go` **(NEW)**

* **Описание:**
  1. Создать файл `internal/models/workout.go` со структурой `Workout`:
     ```go
     type Workout struct {
         ID          int       `json:"id"`
         UserID      string    `json:"user_id"`
         ChallengeID int       `json:"challenge_id"`
         WorkoutDate time.Time `json:"workout_date"`
         Value       int       `json:"value"`
         CreatedAt   time.Time `json:"created_at"`
     }
     ```
  2. Создать файл `internal/models/achievement.go` со структурой `Achievement` и вспомогательными типами:
     ```go
     type Achievement struct {
         ID              int       `json:"id"`
         UserID          string    `json:"user_id"`
         AchievementCode string    `json:"achievement_code"`
         UnlockedAt      time.Time `json:"unlocked_at"`
     }

     // AchievementDefinition описывает ачивку для отображения на фронте
     type AchievementDefinition struct {
         Code        string `json:"code"`
         Name        string `json:"name"`
         Description string `json:"description"`
         Icon        string `json:"icon"`
     }
     ```
  3. Создать структуру ответа `WorkoutResponse` (может быть в `workout.go` или отдельно):
     ```go
     type WorkoutResponse struct {
         Success              bool     `json:"success"`
         Workout              Workout  `json:"workout"`
         UnlockedAchievements []string `json:"unlocked_achievements"`
     }
     ```

* **Ограничения:**
  * Поля JSON-тегов должны строго соответствовать контракту API из `spec.md`.
  * `WorkoutDate` использует `time.Time` для корректной сериализации в `"YYYY-MM-DD"`.

---

### Задача 5: Database Layer — CRUD Workouts с транзакциями

* **Файлы:**
  * `internal/database/workout.go` **(NEW)**

* **Описание:**
  1. **Метод `CreateWorkout(ctx, userID, challengeID, workout) (*models.Workout, error)`:**
     * Работает **внутри SQL-транзакции** (`db.Pool.Begin(ctx)`).
     * Шаг 1: Проверить, что челлендж с `challenge_id` существует, принадлежит `user_id` и имеет статус `'active'`. Если нет — вернуть ошибку (бэкенд вернёт `400` или `404`). Использовать `SELECT ... FOR UPDATE` для блокировки строки челленджа на время транзакции.
     * Шаг 2: Вставить запись в таблицу `workouts`:
       ```sql
       INSERT INTO workouts (user_id, challenge_id, workout_date, value)
       VALUES ($1, $2, $3, $4)
       RETURNING id, created_at
       ```
     * Шаг 3: Обновить `current_progress` в таблице `challenges`:
       ```sql
       UPDATE challenges
       SET current_progress = current_progress + $1,
           status = CASE
               WHEN current_progress + $1 >= target_value THEN 'completed'
               ELSE status
           END
       WHERE id = $2
       RETURNING current_progress, target_value, status
       ```
     * Шаг 4: Сделать `tx.Commit()`.
     * При любой ошибке — `tx.Rollback()`.

  2. **Метод `DeleteWorkout(ctx, userID, workoutID) (*models.Challenge, error)`:**
     * Работает **внутри SQL-транзакции**.
     * Шаг 1: Получить `workout` по `id` и `user_id`, извлечь `challenge_id` и `value`. Если не найден — вернуть ошибку.
     * Шаг 2: Удалить запись из `workouts`:
       ```sql
       DELETE FROM workouts WHERE id = $1 AND user_id = $2
       ```
     * Шаг 3: Обновить `current_progress` в `challenges` (уменьшить на `value` удалённой тренировки). **Важно:** если `current_progress` после вычитания < `target_value`, а `status` был `'completed'`, нужно вернуть статус в `'active'`:
       ```sql
       UPDATE challenges
       SET current_progress = GREATEST(current_progress - $1, 0),
           status = CASE
               WHEN status = 'completed' AND (current_progress - $2) < target_value THEN 'active'
               ELSE status
           END
       WHERE id = $3
       RETURNING id, user_id, name, exercise_id, target_value, current_progress, start_date, end_date, status
       ```
     * Шаг 4: `tx.Commit()`. При ошибке — `tx.Rollback()`.
     * Возвращает обновлённый объект `Challenge` для пересчёта на фронте.

  3. **Метод `GetWorkoutsByChallenge(ctx, userID, challengeID) ([]models.Workout, error)`:**
     * Простой `SELECT` всех тренировок по `challenge_id` и `user_id`, отсортированных по `workout_date DESC, created_at DESC`.
     * Возвращает `[]models.Workout{}` (пустой слайс, не nil) если записей нет.

* **Ограничения:**
  * **Обязательно** использовать транзакции для `CreateWorkout` и `DeleteWorkout`. Это критично для целостности данных.
  * Не использовать ORM. Только чистый SQL через `pgx`.
  * Логировать ошибки через `log.Printf` перед возвратом.
  * Параметр `value` проходит CHECK-ограничение на уровне БД (`value > 0`), но дополнительно валидировать на уровне хэндлера.

---

### Задача 6: Движок ачивок (Achievement Engine)

* **Файлы:**
  * `internal/database/achievement.go` **(NEW)**

* **Описание:**
  1. **Метод `CheckAndUnlockAchievements(ctx, userID, challengeID, newProgress, targetValue int) ([]string, error)`:**
     * Вызывается **после успешного добавления тренировки** (после коммита транзакции `CreateWorkout`).
     * Проверяет условия для 4 ачивок и возвращает массив `[]string` с кодами только **НОВЫХ** (ранее незаработанных) ачивок.
     * Логика проверок:

       **a) `first_step` — «Первый шаг»:**
       * Условие: у пользователя существует хотя бы одна запись в таблице `workouts`.
       * Запрос:
         ```sql
         SELECT EXISTS(SELECT 1 FROM workouts WHERE user_id = $1 LIMIT 1)
         ```
       * Если `true` и ачивка ещё не разблокирована — разблокировать.

       **b) `equator` — «Экватор»:**
       * Условие: `newProgress >= targetValue / 2` (50% от цели).
       * Достаточно проверить аргументы `newProgress` и `targetValue` — если `newProgress * 2 >= targetValue`, ачивка заработана.
       * Если ачивка ещё не разблокирована — разблокировать.

       **c) `hero` — «Герой»:**
       * Условие: `newProgress >= targetValue` И `end_date` челленджа ещё НЕ прошла (т.е. пользователь завершил челлендж до дедлайна).
       * Запрос: получить `end_date` челленджа и сравнить с `CURRENT_DATE`.
       * Если ачивка ещё не разблокирована — разблокировать.

       **d) `stability` — «Стабильность»:**
       * Условие: у пользователя есть тренировки за 3 **последовательных** календарных дня (не обязательно в рамках одного челленджа).
       * Запрос:
         ```sql
         SELECT DISTINCT workout_date FROM workouts
         WHERE user_id = $1
         ORDER BY workout_date DESC
         LIMIT 10
         ```
       * На уровне Go-кода проверить: есть ли среди дат 3 подряд идущие (разница между соседними = 1 день).
       * Если ачивка ещё не разблокирована — разблокировать.

  2. **Вспомогательный метод `unlockAchievement(ctx, userID, code string) (bool, error)`:**
     * Выполняет `INSERT INTO user_achievements (user_id, achievement_code) VALUES ($1, $2) ON CONFLICT DO NOTHING`.
     * Возвращает `true`, если запись была вставлена (новая ачивка), `false` — если уже существовала.

  3. **Метод `GetUserAchievements(ctx, userID) ([]models.Achievement, error)`:**
     * `SELECT * FROM user_achievements WHERE user_id = $1 ORDER BY unlocked_at ASC`.
     * Нужен для отображения разблокированных ачивок на дашборде.

* **Ограничения:**
  * Каждая проверка ачивки должна быть идемпотентной: повторный вызов не создаёт дубликат (обеспечивается `ON CONFLICT DO NOTHING` на уровне БД + `UNIQUE(user_id, achievement_code)`).
  * Метод `CheckAndUnlockAchievements` должен **никогда не роняться** — если одна проверка упала, залогировать ошибку и продолжить проверку остальных.
  * Возвращать только **вновь разблокированные** коды (те, которые `unlockAchievement` вернул `true`), чтобы фронтенд показал поп-ап ровно один раз.

---

### Задача 7: HTTP Handlers для Workouts и регистрация маршрутов

* **Файлы:**
  * `internal/handlers/workout_handler.go` **(NEW)**
  * `internal/handlers/router.go` **(MODIFY)**

* **Описание:**
  1. **Создать `WorkoutHandler` в `internal/handlers/workout_handler.go`:**

     **a) `HandleCreateWorkout` — `POST /api/challenges/:id/workouts`:**
     * Извлечь `X-User-Id` из заголовков.
     * Извлечь `challengeID` из URL-пути (`/api/challenges/{id}/workouts`).
       * Путь разбирается через `strings.Split(strings.Trim(r.URL.Path, "/"), "/")`. Ожидаемая структура: `["api", "challenges", "{id}", "workouts"]`. `challengeID` — элемент с индексом 2.
     * Десериализовать тело запроса в структуру с полями `WorkoutDate` и `Value`.
     * **Валидация:**
       * `value` > 0, иначе → `400 Bad Request` с сообщением `"Value must be greater than 0"`.
       * `workout_date` должна быть валидной датой, иначе → `400 Bad Request`.
     * Вызвать `db.CreateWorkout(ctx, userID, challengeID, &workout)`.
       * Если челлендж не найден или не active → `400 Bad Request` с сообщением `"Challenge not found or not active"`.
       * Если ошибка БД → `500 Internal Server Error`.
     * При успехе — вызвать `db.CheckAndUnlockAchievements(ctx, userID, challengeID, newProgress, targetValue)`.
     * Сформировать ответ:
       ```json
       {
         "success": true,
         "workout": { ... },
         "unlocked_achievements": ["first_step", "equator"]
       }
       ```
     * Вернуть HTTP `201 Created`.
     * **Важно:** Если `CheckAndUnlockAchievements` упал с ошибкой, тренировка всё равно считается успешно добавленной. Залогировать ошибку ачивок, вернуть `201` с пустым массивом `unlocked_achievements: []`.

     **b) `HandleDeleteWorkout` — `DELETE /api/workouts/:id`:**
     * Извлечь `X-User-Id` из заголовков.
     * Извлечь `workoutID` из URL-пути (`/api/workouts/{id}`). Ожидаемая структура: `["api", "workouts", "{id}"]`. `workoutID` — элемент с индексом 2.
     * Вызвать `db.DeleteWorkout(ctx, userID, workoutID)`.
       * Если тренировка не найдена → `404 Not Found`.
       * Если ошибка БД → `500 Internal Server Error`.
     * При успехе — вернуть обновлённый объект `Challenge` (из `DeleteWorkout`):
       ```json
       {
         "success": true,
         "challenge": { ... }
       }
       ```
     * Вернуть HTTP `200 OK`.

  2. **Обновить `internal/handlers/router.go`:**
     * Создать экземпляр `WorkoutHandler`:
       ```go
       workoutHandler := NewWorkoutHandler(db)
       ```
     * Зарегистрировать маршруты:
       * Путь `/api/challenges/` уже зарегистрирован для `GET /api/challenges/:id`. Его нужно **расширить**: если URL содержит `/workouts` в конце — делегировать в `workoutHandler.HandleCreateWorkout`. Иначе — как раньше, `challengeHandler.HandleGetByID`.
         ```go
         mux.HandleFunc("/api/challenges/", func(w http.ResponseWriter, r *http.Request) {
             path := strings.Trim(r.URL.Path, "/")
             // Check if this is a workout sub-route: api/challenges/{id}/workouts
             if strings.HasSuffix(path, "/workouts") && r.Method == http.MethodPost {
                 workoutHandler.HandleCreateWorkout(w, r)
                 return
             }
             // Otherwise, treat as challenge detail
             if r.Method == http.MethodGet {
                 challengeHandler.HandleGetByID(w, r)
             } else {
                 http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
             }
         })
         ```
       * Новый путь `/api/workouts/` для `DELETE`:
         ```go
         mux.HandleFunc("/api/workouts/", func(w http.ResponseWriter, r *http.Request) {
             if r.Method == http.MethodDelete {
                 workoutHandler.HandleDeleteWorkout(w, r)
             } else {
                 http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
             }
         })
         ```

* **Ограничения:**
  * Не ломать существующие маршруты (`GET /api/challenges/:id` должен работать как раньше).
  * Все ошибки логировать через `log.Printf`.
  * Заголовок ответа всегда `Content-Type: application/json`.
  * Массив `unlocked_achievements` в ответе **всегда** должен быть `[]`, а не `null` (инициализировать пустым слайсом).

---

### Задача 8: API для получения ачивок пользователя

* **Файлы:**
  * `internal/handlers/achievement_handler.go` **(NEW)**
  * `internal/handlers/router.go` **(MODIFY)**

* **Описание:**
  1. **Создать `AchievementHandler` в `internal/handlers/achievement_handler.go`:**

     **`HandleList` — `GET /api/achievements`:**
     * Извлечь `X-User-Id` из заголовков.
     * Вызвать `db.GetUserAchievements(ctx, userID)`.
     * Вернуть массив ачивок. Если пустой — вернуть `[]`, а не `null`.
     * HTTP `200 OK`.

  2. **Зарегистрировать в `router.go`:**
     ```go
     achievementHandler := NewAchievementHandler(db)
     mux.HandleFunc("/api/achievements", func(w http.ResponseWriter, r *http.Request) {
         if r.Method == http.MethodGet {
             achievementHandler.HandleList(w, r)
         } else {
             http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
         }
     })
     ```

* **Ограничения:**
  * Эндпоинт нужен для того, чтобы дашборд мог подсветить разблокированные ачивки при загрузке страницы.

---

## Epic: US-8 Удаление челленджа

**Цель:** Реализовать API для удаления челленджа пользователем.

### Задача 12: Метод DeleteChallenge (БД) и HandleDelete (Handler)

* **Файлы:** `internal/database/challenge.go`, `internal/handlers/challenge_handler.go`, `internal/handlers/router.go`
* **Описание:**
  1. В `internal/database/challenge.go` реализовать метод:
     * `DeleteChallenge(ctx, userID, id)`: удалить запись с `WHERE id = $1 AND user_id = $2`. Если `RowsAffected() == 0` — вернуть `pgx.ErrNoRows` (не найден или чужой).
     * Благодаря `ON DELETE CASCADE` в схеме все `workouts` удалятся автоматически.
  2. В `internal/handlers/challenge_handler.go` реализовать:
     * `HandleDelete(w, r)` — обрабатывает `DELETE /api/challenges/:id`.
     * Парсить `id` из URL. Если некорректный — `400 Bad Request`.
     * Вызвать `db.DeleteChallenge`. Если `pgx.ErrNoRows` — `404 Not Found`. Если другая ошибка — `500`.
     * При успехе — `200 OK` с `{ "success": true }`.
  3. В `internal/handlers/router.go` добавить `case http.MethodDelete: challengeHandler.HandleDelete(w, r)` в блок обработки `/api/challenges/:id`.

* **Ограничения:**
  * Реализация должна быть user-scoped: нельзя удалить чужой челлендж (проверка `user_id`).
  * Логировать все ошибки.

---

## Epic: US-9 Авторизация через Telegram Mini App

**Цель:** Реализовать Middleware для валидации Telegram `initData` и безопасной аутентификации пользователей.

### Задача 13: Создание TelegramAuthMiddleware
* **Файлы:** `internal/handlers/middleware.go` (NEW), `internal/handlers/router.go`
* **Описание:**
  1. Реализовать пакет/функцию `TelegramAuthMiddleware(next http.Handler, botToken string) http.Handler`.
  2. Middleware должен проверять наличие заголовка `Authorization: Bearer <initData>`.
  3. Если токен бота не задан или пуст (режим обхода / dev-режим), Middleware должен считывать заголовок `X-User-Id` и передавать управление дальше (fallback).
  4. Если токен задан, Middleware:
     * Парсит `initData`.
     * Проверяет подпись алгоритмом HMAC-SHA256 (сверка `hash` с расчетом по `WebAppData`).
     * Проверяет, что `auth_date` не старше 24 часов.
     * Если данные валидны, извлекает `user.id`, преобразует в строку и записывает в `X-User-Id` заголовка запроса или передает через `context`.
     * Если невалидны — возвращает HTTP `401 Unauthorized`.
  5. В `internal/handlers/router.go` обернуть все обработчики `/api/*` в этот Middleware.
* **Ограничения:**
  * Алгоритм валидации должен строго соответствовать документации Telegram.
  * Тщательно логировать причины 401 ошибок (без утечки самого токена).

---

## Epic: Архитектура и Безопасность

### Задача 14: Архитектурное и Code Ревью (Backend-14 / Architect-1)
* **Файлы:** `internal/handlers/*`, `internal/database/*`
* **Описание:** Провести полное код-ревью реализации MVP и интеграции с Telegram (US-9). Проверить безопасность валидации `initData`, отсутствие хардкода (особенно `default_user_1`), соблюдение принципов Loose Coupling. Составить список технического долга и исправить критические замечания.
* **Ограничения:** Уделить особое внимание SQL-транзакциям при обновлении статусов челленджей.

---

## Epic: US-13 Редактирование челленджа

**Цель:** Реализовать эндпоинт частичного обновления челленджа с проверкой бизнес-правил.

### Задача 15: Реализация PATCH /api/challenges/:id
* **Файлы:** `internal/handlers/challenge_handler.go`, `internal/database/challenge.go`, `internal/handlers/router.go`
* **Описание:**
  1. В `internal/database/challenge.go` добавить метод `UpdateChallenge(ctx, userID, id, updates)`.
  2. В `internal/handlers/challenge_handler.go` реализовать `HandleUpdate(w, r)`.
     * Валидировать, что челлендж активен (статус `active`).
     * Игнорировать поле `exercise_id`, если оно передано.
     * Валидировать даты: `end_date` не раньше текущей даты, `start_date` можно менять только если она ещё не наступила.
     * Валидировать `target_value` >= 1.
     * Если новое `target_value` <= `current_progress`, то в обновлении сразу менять `status = 'completed'`.
  3. В `internal/handlers/router.go` добавить обработку `PATCH` для `/api/challenges/:id`.
* **Ограничения:**
  * Убедиться, что обновляются только те поля, которые были переданы. Использовать `UPDATE ... COALESCE` или динамический SQL.

---

## Epic: US-14 Ачивки внутри челленджа

**Цель:** Реализовать логику выдачи и получения ачивок (8 штук) с привязкой к конкретному челленджу, а не глобально к пользователю.

### Задача 16: Миграция БД и обновление моделей для ачивок
* **Файлы:** Скрипт миграции (или ручное обновление схемы), `internal/models/achievement.go`
* **Описание:**
  1. Изменить таблицу `user_achievements`:
     * Добавить колонку `challenge_id` (foreign key на `challenges.id`, `ON DELETE CASCADE`).
     * Изменить уникальный индекс с `UNIQUE(user_id, achievement_code)` на `UNIQUE(user_id, challenge_id, achievement_code)`.
  2. Обновить структуру `Achievement` в `models`: добавить поле `ChallengeID`.

### Задача 17: Обновление Движка Ачивок (8 ачивок)
* **Файлы:** `internal/database/achievement.go`
* **Описание:**
  1. Изменить метод проверки ачивок на `CheckAndUnlockAchievements(ctx, userID, challengeID, newProgress, targetValue int) ([]string, error)`.
  2. Внутри метода реализовать проверку 8 ачивок:
     * `first_step` (существует тренировка для данного `challenge_id`)
     * `equator` (`newProgress >= targetValue / 2`)
     * `hero` (`newProgress >= targetValue` и `end_date` не прошла)
     * `stability` (тренировки 3 дня подряд в данном челлендже)
     * `power_start` (разовая добавленная тренировка $\ge$ 25% от цели)
     * `overachiever` (`newProgress >= targetValue * 1.2`)
     * `early_bird` (добавление тренировки с 5:00 до 8:59)
     * `final_spurt` (достижение 100% ровно в `end_date`)
  3. Обновить `unlockAchievement`, чтобы он принимал и записывал `challenge_id`.

### Задача 18: API для получения ачивок челленджа
* **Файлы:** `internal/handlers/achievement_handler.go`, `internal/handlers/router.go`
* **Описание:**
  1. Реализовать `GET /api/challenges/:id/achievements`.
  2. Получать ачивки конкретного пользователя для конкретного челленджа из `user_achievements`.
  3. Зарегистрировать роут в `router.go` (удалив старый глобальный роут `/api/achievements`, если он больше не нужен).
* **Ограничения:** Не забыть возвращать `[]` вместо `null`, если ачивок нет.
