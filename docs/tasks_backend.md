# Backend Tasks

## Epic: US-1 Создание упражнения

**Цель:** Реализовать API для получения списка упражнений и добавления новых кастомных упражнений пользователем.

### Задача 1: Data Model & Database Layer
* **Файл(ы):** `internal/models/exercise.go`, `internal/database/exercise_db.go`
* **Описание:** 
  1. Создать структуру `Exercise` (id, user_id, name, is_custom).
  2. Написать функцию `CreateExercise(ctx, pool, exercise)` с использованием чистого SQL через `pgx`.
  3. Написать функцию `GetExercises(ctx, pool, userID)` для получения списка (должны возвращаться системные упражнения `user_id = 'system'` и упражнения текущего пользователя).
* **Ограничения:** 
  * Обработать ошибку нарушения уникального индекса `unique_user_exercise` и возвращать понятную ошибку уровня приложения.

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
