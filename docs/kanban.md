# Kanban Board

## Спринт №2: Планирование (US-2: Создание челленджа)
**Цель спринта:** Реализовать сценарий создания челленджей (US-2), включая сохранение в БД, валидацию дат и отображение на клиенте.

### TO DO
- [ ] **Backend-3: API Создания и Получения Челленджей**
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
- [ ] **Frontend-4: API Client & Store for Challenges**
  * **Файлы:** `frontend/js/api.js`, `frontend/js/store.js`
  * **Описание:**
    1. В `api.js` добавить методы:
       * `getChallenges()`: делает `GET /api/challenges`.
       * `createChallenge(payload)`: делает `POST /api/challenges` с телом челленджа.
    2. В `store.js` добавить:
       * Поле `challenges: []` в начальное состояние.
       * Метод `setChallenges(challenges)` для обновления списка в стейте.
       * Метод `addChallenge(challenge)` для реактивного добавления нового челленджа в список.
- [ ] **Frontend-5: Интеграция Формы и Отображение на Дашборде**
  * **Файлы:** `frontend/js/components/challenge/challenge-form.js`, `frontend/js/components/dashboard/dashboard.js`, `frontend/js/app.js`
  * **Описание:**
    1. В `challenge-form.js` заменить демо-заглушку создания челленджа на реальный вызов `api.createChallenge(challengePayload)`. При успешном сохранении добавлять челлендж в стор через `store.addChallenge` и перенаправлять на `'dashboard'`.
    2. В `dashboard.js` реализовать отображение списка челленджей из стора:
       * Если список пуст — показывать заглушку «У вас пока нет активных челленджей...».
       * Если не пуст — рендерить список карточек. Каждая карточка должна отображать: Название челленджа, Дату начала и окончания, Прогресс-бар (актуальный прогресс от цели).
    3. В `app.js` при загрузке страницы вызывать `api.getChallenges()` параллельно с упражнениями и обновлять стор.
  * **Ограничения:**
    * Использовать Vanilla JS для динамической вставки элементов.
    * Стилизовать карточки челленджей с помощью HSL палитры Telegram в `main.css`.
- [ ] **QA-3: API Тестирование Челленджей (cURL / Postman)**
  * **Описание:**
    1. **TC-2.1 (Positive): Успешное создание челленджа**
       * *Steps:* `POST /api/challenges` с телом `{"name": "3000 отжиманий", "exercise_id": 1, "target_value": 3000, "start_date": "2026-06-01", "end_date": "2026-06-30"}`.
       * *Expected:* HTTP 201 Created. В базе данных создана запись со статусом `active` и `current_progress = 0`.
    2. **TC-2.2 (Negative): Дедлайн раньше старта (AC-2)**
       * *Steps:* Отправить запрос с `start_date = "2026-06-30"` и `end_date = "2026-06-01"`.
       * *Expected:* HTTP 400 Bad Request. В БД запись не добавляется.
    3. **TC-2.3 (Negative): Целевое количество <= 0**
       * *Steps:* Отправить запрос с `target_value = 0` или `target_value = -100`.
       * *Expected:* HTTP 400 Bad Request.
- [ ] **QA-4: UI/UX Тестирование Челленджей (Браузер)**
  * **Описание:**
    1. **TC-2.4 (Positive): Успешное создание и показ на Дашборде (AC-1)**
       * *Steps:* Заполнить форму (выбрать упражнение, ввести цель 100, даты), нажать «Создать».
       * *Expected:* Форма отправляется, показывается тост об успехе, происходит редирект на Дашборд. На дашборде отображается карточка нового челленджа с нулевым прогрессом и правильными датами.
    2. **TC-2.5 (Negative): Блокировка отправки при некорректных датах (AC-2)**
       * *Steps:* Выбрать дату окончания раньше даты начала.
       * *Expected:* Кнопка отправки заблокирована, или при клике выводится тост с предупреждением об ошибке валидации, форма не уходит в API.

### IN PROGRESS

### QA / REVIEW

### DONE

