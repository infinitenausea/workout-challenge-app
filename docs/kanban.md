# Kanban Board

## 🏃 Спринт №7: Редактирование челленджей (US-13)
**Цель:** Реализовать возможность редактировать созданные челленджи через UI с обновлением на бэкенде.

---

### TO DO

**Backend-15: Реализация PATCH /api/challenges/:id**
1. В `internal/database/challenge.go` добавить метод `UpdateChallenge(ctx, userID, id, updates)`.
2. В `internal/handlers/challenge_handler.go` реализовать `HandleUpdate(w, r)`.
   * Валидировать, что челлендж активен (статус `active`).
   * Игнорировать поле `exercise_id`, если оно передано.
   * Валидировать даты: `end_date` не раньше текущей даты, `start_date` можно менять только если она ещё не наступила.
   * Валидировать `target_value` >= 1.
   * Если новое `target_value` <= `current_progress`, то в обновлении сразу менять `status = 'completed'`.
3. В `internal/handlers/router.go` добавить обработку `PATCH` для `/api/challenges/:id`.
Ограничения: Убедиться, что обновляются только те поля, которые были переданы. Использовать `UPDATE ... COALESCE` или динамический SQL.

**Frontend-21: Метод API и обновление Store**
1. В `api.js` добавить метод `updateChallenge(id, payload)` -> `PATCH /api/challenges/:id`.
2. В `store.js` добавить метод `updateChallengeInList(updatedChallenge)` для обновления нужного челленджа в массиве `challenges` и `currentChallenge`.

**Frontend-22: UI Формы редактирования**
1. На страницу `challenge-detail.js` добавить кнопку «Редактировать» (только если челлендж в статусе `active`). Кнопка должна переводить пользователя на форму редактирования `challenge-form.js` (например, через `store.navigate('challenge-form', { editMode: true, challengeId: c.id })`).
2. Адаптировать `challenge-form.js`:
   * Если `editMode`, предзаполнять поля данными текущего челленджа.
   * Заблокировать (disable) селект выбора упражнения.
   * Заблокировать поле `start_date`, если оригинальная дата уже наступила или прошла.
   * При сабмите вызывать `api.updateChallenge`, после чего возвращать пользователя на детали челленджа.
Ограничения: Переиспользовать существующий компонент формы, чтобы избежать дублирования кода.

**QA: Тестирование US-13 (TC-13.1 - TC-13.6)**
Проверить API и UI частичного обновления челленджей:
* TC-13.1 (Positive): Успешное обновление названия и даты
* TC-13.2 (Positive): Снижение цели и автозавершение
* TC-13.3 (Negative): Попытка изменить exercise_id
* TC-13.4 (Negative): Некорректные даты
* TC-13.5 (Positive): Предзаполнение формы
* TC-13.6 (Positive): Сохранение изменений

---

### IN PROGRESS

---

### QA / REVIEW

---

### DONE
