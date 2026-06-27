# Kanban Board

## 🚀 Спринт №10: Редактирование челленджей (US-13)
**Цель:** Реализовать возможность редактировать созданные челленджи через UI с обновлением на бэкенде.

---

### TO DO

---

### IN PROGRESS

---

### QA / REVIEW

---

### DONE

- [x] **Backend-15: Реализация PATCH /api/challenges/:id**
  * `internal/database/challenge.go` обновлен с функцией UpdateChallenge (COALESCE).
  * `internal/handlers/challenge_handler.go` обновлен (валидация active статуса, логика дат, target_value).
  * `internal/handlers/router.go` обновлен (добавлен PATCH).
- [x] **Frontend-21: Метод API и обновление Store**
  * В `api.js` добавлен метод `updateChallenge(id, payload)`.
  * В `store.js` добавлен метод `updateChallengeInList(updatedChallenge)`.
- [x] **Frontend-22: UI Формы редактирования**
  * Кнопка редактирования добавлена на `challenge-detail.js`.
  * Form `challenge-form.js` адаптирована для `editMode` (презаполнение, блокировки, отправка PATCH, роутинг).
- [x] **QA: Тестирование US-13 (TC-13.1 - TC-13.6)**
  * Успешно проведены и пройдены все позитивные, негативные и граничные E2E-тесты в Playwright.
