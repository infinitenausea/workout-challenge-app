# Completed Sprint №2: Создание челленджа (US-2)

**Цель спринта:** Реализовать сценарий создания челленджей (US-2), включая сохранение в БД, валидацию дат и отображение на клиенте.

## Статус: Успешно завершён ✅

---

### Выполненные задачи:

1. **Backend-3: API Создания и Получения Челленджей**
   * **Статус:** DONE ✅
   * **Файлы:** `internal/models/challenge.go`, `internal/database/challenge.go`, `internal/handlers/challenge_handler.go`, `internal/handlers/router.go`
   * **Описание:** Создана модель `Challenge`, реализованы методы БД (`CreateChallenge`, `GetChallenges`, `GetChallengeByID`) и эндпоинты `POST /api/challenges`, `GET /api/challenges`, `GET /api/challenges/:id` с валидацией бизнес-правил (непустое имя, `target_value > 0`, `end_date >= start_date`).

2. **Frontend-4: API Client & Store for Challenges**
   * **Статус:** DONE ✅
   * **Файлы:** `frontend/js/api.js`, `frontend/js/store.js`
   * **Описание:** В `api.js` добавлены методы `getChallenges()` и `createChallenge(payload)`. В `store.js` добавлены методы `setChallenges(challenges)` и `addChallenge(challenge)`.

3. **QA-3: API Тестирование Челленджей (cURL / Postman)**
   * **Статус:** DONE ✅
   * **Описание:** Все 6 тест-кейсов пройдены. TC-2.1 (создание), TC-2.2 (дедлайн < старт), TC-2.3 (target <= 0), TC-2.4 (пустое имя), TC-2.5 (список), TC-2.6 (по ID) — все PASS.

4. **Frontend-5: Интеграция Формы и Отображение на Дашборде**
   * **Статус:** DONE ✅
   * **Файлы:** `frontend/js/components/challenge/challenge-form.js`, `frontend/js/components/dashboard/dashboard.js`, `frontend/js/app.js`, `frontend/css/main.css`
   * **Описание:** Форма создания челленджа подключена к реальному API. Дашборд отображает список карточек с прогресс-барами и датами. При пустом списке — заглушка. При загрузке приложения данные подтягиваются из API.

5. **QA-4: UI/UX Тестирование Челленджей (Браузер)**
   * **Статус:** DONE ✅
   * **Описание:** TC-2.4 (создание челленджа, HTTP 201, `current_progress=0`) и TC-2.5 (валидация дат, HTTP 400) — оба PASS. Bugs Found: None.

---

### Итоги спринта

| Метрика | Значение |
|---|---|
| Задач в спринте | 5 |
| Завершено | 5 |
| Открытых багов | 0 |
| Покрытых Acceptance Criteria | AC-1 (создание), AC-2 (валидация дат) |
