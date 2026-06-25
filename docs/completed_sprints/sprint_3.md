# Kanban Board — Sprint 3 Archive

## 🚀 Спринт №3 (US-3: Логирование тренировки и Ачивки)
**Цель спринта:** Реализовать полный цикл добавления и удаления тренировок с транзакционным пересчётом прогресса, систему достижений (ачивок) с поздравительными поп-апами и страницу деталей челленджа.

---

### TO DO

---

### IN PROGRESS

---

### QA / REVIEW

---

### DONE

- [x] **QA-5: API Тестирование Workouts и Ачивок (cURL / Postman)**
  * **Описание:** Набор из 17 тест-кейсов для проверки API:
    * **Добавление тренировки:**
      * TC-3.1 (Positive): Успешное добавление, проверка `201`, `unlocked_achievements: ["first_step"]`, обновление `current_progress`.
      * TC-3.2 (Positive): Ачивка «Экватор» при достижении 50%.
      * TC-3.3 (Positive): Ачивка «Герой» при 100% до дедлайна, `status → completed`.
      * TC-3.4 (Positive): Ачивка «Стабильность» при 3 днях подряд.
      * TC-3.5 (Negative): Отрицательное `value` → 400.
      * TC-3.6 (Negative): `value = 0` → 400.
      * TC-3.7 (Negative): Несуществующий челлендж → 400/404.
      * TC-3.8 (Negative): Завершённый челлендж → 400.
      * TC-3.9 (Edge): Пустая дата → 400.
      * TC-3.10 (Edge): Ачивки не дублируются.
    * **Удаление тренировки:**
      * TC-3.11 (Positive): Успешное удаление, `current_progress` уменьшился.
      * TC-3.12 (Positive): Каскадный откат `completed → active`.
      * TC-3.13 (Negative): Несуществующая тренировка → 404.
      * TC-3.14 (Negative): Чужая тренировка → 404.
      * TC-3.15 (Edge): `current_progress` не уходит ниже 0.
    * **API Ачивок:**
      * TC-3.16 (Positive): Получение списка ачивок.
      * TC-3.17 (Positive): Пустой список → `[]`, не `null`.
  * **Зависимости:** Backend-4, Backend-5, Backend-6, Backend-7, Backend-8.

---

- [x] **QA-6: UI/UX Тестирование Workouts и Ачивок (Браузер)**
  * **Описание:** Набор из 13 тест-кейсов для UI:
    * **Навигация:**
      * TC-3.18: Переход на страницу деталей по клику на карточку.
      * TC-3.19: Кнопка «Назад» возвращает на дашборд.
    * **Добавление тренировки:**
      * TC-3.20: E2E флоу — прогресс-бар мгновенно обновляется без перезагрузки (AC-1).
      * TC-3.21: Поздравительный поп-ап при «Экватор» (AC-2).
      * TC-3.22: Валидация — пустое количество.
      * TC-3.23: Валидация — отрицательное количество.
    * **Удаление тренировки:**
      * TC-3.24: Удаление из списка, пересчёт прогресса.
      * TC-3.25: Откат `completed → active` после удаления (US-4 AC-1).
      * TC-3.26: Отмена удаления в `confirm()`.
    * **Ачивки и модалки:**
      * TC-3.27: Подсветка разблокированных ачивок на дашборде.
      * TC-3.28: Модалка закрывается по оверлею и Escape.
    * **Граничные случаи:**
      * TC-3.29: 100% прогресс и возврат на дашборд.
      * TC-3.30: Множественные ачивки за одну тренировку (3 последовательных поп-апа).
  * **Зависимости:** Frontend-6...Frontend-11, Backend-4...Backend-8 (полная интеграция).

---

- [x] **Frontend-6: API Client — Workout & Achievement Methods**
  * **Описание:**
    1. Добавить метод `createWorkout(challengeId, payload)`:
       * `POST /api/challenges/${challengeId}/workouts`
       * Тело: `{ "workout_date": "YYYY-MM-DD", "value": number }`
       * Возвращает весь JSON-ответ: `{ success, workout, unlocked_achievements }`.
    2. Добавить метод `deleteWorkout(workoutId)`:
       * `DELETE /api/workouts/${workoutId}`
       * Возвращает: `{ success, challenge }`.
    3. Добавить метод `getChallengeDetail(challengeId)`:
       * `GET /api/challenges/${challengeId}`
       * Возвращает детальную информацию о челлендже.
    4. Добавить метод `getAchievements()`:
       * `GET /api/achievements`
       * Возвращает массив разблокированных ачивок.

---

- [x] **Frontend-7: Store — Состояние для Workouts и Achievements**
  * **Описание:**
    1. Добавить новые поля в `this.state`:
       * `currentChallenge: null` — детали текущего просматриваемого челленджа.
       * `workouts: []` — список тренировок текущего челленджа.
       * `achievements: []` — разблокированные ачивки пользователя.
    2. Добавить методы:
       * `setCurrentChallenge(challenge)` — устанавливает текущий челлендж.
       * `setWorkouts(workouts)` — устанавливает список тренировок.
       * `addWorkout(workout)` — добавляет тренировку в **начало** массива (сортировка DESC).
       * `removeWorkout(workoutId)` — удаляет тренировку по `id`.
       * `updateChallengeProgress(challengeId, newProgress, newStatus)` — обновляет `current_progress` и `status` в массиве `challenges` **И** в `currentChallenge` (если совпадает `id`).
       * `setAchievements(achievements)` — устанавливает массив ачивок.
       * `addAchievements(newCodes)` — добавляет новые коды ачивок.

---

- [x] **Frontend-8: Компонент «Страница деталей челленджа» (Challenge Detail)**
  * **Описание:**
    1. **Создать класс `ChallengeDetail`** с методами `constructor(container)`, `mount()`, `unmount()`, `render()`.
    2. **При монтировании:**
       * Получить `currentChallengeId` из `store.getState()`.
       * Вызвать `api.getChallengeDetail(id)` → `store.setCurrentChallenge(challenge)`.
       * Подписаться на изменения стора.
    3. **Рендеринг — что отображать:**
       * **Заголовок:** Кнопка «← Назад» (→ `store.navigate('dashboard')`), название, статус-бейдж.
       * **Прогресс-бар (крупный):** Процент, текст `"X / Y"`, визуальная полоса (высота 12px).
       * **Таймер обратного отсчёта:** Разница `end_date` и сегодня.
       * **Кнопка «Добавить тренировку»:** Открывает модалку (Frontend-9).
       * **История тренировок:** Список карточек (дата DD.MM.YYYY, value, иконка 🗑️).
       * **Обработка удаления:** `click` на 🗑️ → `confirm()` → `api.deleteWorkout(id)` → `store.removeWorkout(id)` + `store.updateChallengeProgress(...)` + `store.setCurrentChallenge(...)`.

---

- [x] **Frontend-9: Модальное окно «Добавить тренировку» (Workout Modal)**
  * **Описание:**
    1. **Создать класс `WorkoutModal`** с методами `constructor()`, `open(challengeId)`, `close()`.
    2. **Разметка:** Оверлей с формой: поле даты (`type="date"`, дефолт — сегодня), поле количества (`type="number"`, `min="1"`), кнопка «Сохранить».
    3. **Логика отправки:**
       * Клиентская валидация: `value > 0`, `workout_date` заполнено.
       * Блокировка кнопки на время запроса.
       * `api.createWorkout(challengeId, { workout_date, value })`.
       * **При успехе:**
         1. `store.addWorkout(response.workout)`.
         2. `store.updateChallengeProgress(challengeId, newProgress, newStatus)`.
         3. Закрыть модалку.
         4. Тост `"Тренировка добавлена! +{value} повторений"`.
         5. Если `unlocked_achievements.length > 0` → показать Achievement Popup (Frontend-10).

---

- [x] **Frontend-10: Поздравительный поп-ап при получении ачивки (Achievement Popup)**
  * **Описание:**
    1. **Создать функцию `showAchievementPopup(achievementCodes)`:**
       * Принимает массив кодов: `["first_step", "equator"]`.
       * Локальный маппинг: `first_step` 🌱, `equator` 📈, `hero` ⚡, `stability` 🔥.
       * Рендерит оверлей с **анимированной** карточкой: иконка, название, описание, кнопка «Отлично!».
       * Несколько ачивок → показ **последовательно**.

---

- [x] **Frontend-11: Регистрация маршрута Challenge Detail и обновление дашборда**
  * **Описание:**
    1. **В `app.js`:**
       * Импортировать `ChallengeDetail` из `./components/challenge/challenge-detail.js`.
       * Добавить маршрут: `'challenge-detail': ChallengeDetail`.
       * При инициализации загрузить ачивки: `api.getAchievements()` → `store.setAchievements(...)`.
    2. **В `dashboard.js`:**
       * **Кликабельные карточки:** Обработчик `click` на `.challenge-card` → `store.navigate('challenge-detail', { currentChallengeId: id })`.
       * **Динамические ачивки:** Заменить захардкоженный блок ачивок на динамический.

---

- [x] **Backend-4: Модель Workout и структуры ответа**
  * **Описание:**
    1. Создать файл `internal/models/workout.go` со структурой `Workout`.
    2. Создать файл `internal/models/achievement.go` со структурами `Achievement` и `AchievementDefinition`.
    3. Создать структуру ответа `WorkoutResponse`.

---

- [x] **Backend-5: Database Layer — CRUD Workouts с транзакциями**
  * **Описание:**
    1. **Метод `CreateWorkout(ctx, userID, challengeID, workout) (*models.Workout, error)`:**
       * Работает **внутри SQL-транзакции** (`db.Pool.Begin(ctx)`).
       * Проверяет существование/статус челленджа, вставляет запись в `workouts`, обновляет `current_progress` и `status` в `challenges`.
    2. **Метод `DeleteWorkout(ctx, userID, workoutID) (*models.Challenge, error)`:**
       * Работает **внутри SQL-транзакции**.
       * Удаляет тренировку, уменьшает `current_progress` в `challenges` с каскадным откатом статуса.
    3. **Метод `GetWorkoutsByChallenge(ctx, userID, challengeID) ([]models.Workout, error)`:**
       * Выбирает тренировки по челленджу, отсортированные DESC.

---

- [x] **Backend-6: Движок ачивок (Achievement Engine)**
  * **Описание:**
    1. **Метод `CheckAndUnlockAchievements(ctx, userID, challengeID, newProgress, targetValue int) ([]string, error)`:**
       * Проверяет условия для 4 ачивок (`first_step`, `equator`, `hero`, `stability`) и возвращает коды новых.
    2. **Вспомогательный метод `unlockAchievement(ctx, userID, code string) (bool, error)`:**
       * `INSERT ... ON CONFLICT DO NOTHING`.
    3. **Метод `GetUserAchievements(ctx, userID) ([]models.Achievement, error)`:**
       * Возвращает список ачивок пользователя.

---

- [x] **Backend-7: HTTP Handlers для Workouts и регистрация маршрутов**
  * **Описание:**
    1. **Создать `WorkoutHandler` в `workout_handler.go`:**
       * `HandleCreateWorkout` — `POST /api/challenges/:id/workouts`
       * `HandleDeleteWorkout` — `DELETE /api/workouts/:id`
    2. **Обновить `router.go`:**
       * Зарегистрировать новые маршруты.

---

- [x] **Backend-8: API для получения ачивок пользователя**
  * **Описание:**
    1. **Создать `AchievementHandler` в `achievement_handler.go`:**
       * `HandleList` — `GET /api/achievements`
    2. **Зарегистрировать в `router.go`.**
