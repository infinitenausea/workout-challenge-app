# Архив: Спринт №4 — US-8 Удаление челленджа

**Статус:** ✅ ЗАКРЫТ  
**Дата закрытия:** 2026-06-26  
**Цель спринта:** Реализовать функцию удаления челленджа — полный цикл от API до UI с интеграционным и Playwright тестированием.

---

## Итоговая Канбан-доска

### TO DO
*(пусто)*

---

### IN PROGRESS
*(пусто)*

---

### QA / REVIEW
*(пусто)*

---

### DONE

- **[BE-12] DB: Метод `DeleteChallenge`** — `internal/database/challenge.go`
- **[BE-12] Handler: `HandleDelete`** — `internal/handlers/challenge_handler.go`
- **[BE-12] Router: DELETE `/api/challenges/:id`** — `internal/handlers/router.go`
- **[FE-14] API Client: `deleteChallenge()`** — `frontend/js/api.js`
- **[FE-15] Store: `removeChallenge()`** — `frontend/js/store.js`
- **[FE-16] CSS: стиль `button.danger`** — `frontend/css/main.css`
- **[FE-16] UI: кнопка "Удалить челлендж"** — `frontend/js/components/challenge/challenge-detail.js`
- **[QA] TC-3.18 Positive: Integration test** ✅ PASS
- **[QA] TC-3.18 Negative: Integration test** ✅ PASS
- **[QA] TC-4.1 Playwright: Подтверждение удаления → редирект, челлендж исчез** ✅ PASS
- **[QA] TC-4.2 Playwright: Отмена удаления → остаётся на странице** ✅ PASS
- **[QA] TC-4.3 Playwright: Кнопка "Удалить челлендж" видима** ✅ PASS
- **[DOCS] spec.md: US-8, DELETE `/api/challenges/:id`**
- **[DOCS] tasks_backend.md: Epic US-8, Задача 12**
- **[DOCS] tasks_frontend.md: Epic US-8, Задачи 14–16**
- **[DOCS] tasks_qa.md: Epic US-8, TC-3.18, TC-4.1–4.3**

---

## Метрики спринта

| Метрика | Значение |
|---------|----------|
| Запланировано задач | 8 (BE: 3, FE: 4, DOCS: 1) |
| Завершено задач | 8 / 8 (100%) |
| Тест-кейсов написано | 5 (TC-3.18×2, TC-4.1, TC-4.2, TC-4.3) |
| Тест-кейсов прошло | 5 / 5 (100%) ✅ |
| Инструменты тестирования | Go integration tests + Playwright MCP |
| User Story закрыта | US-8 ✅ |
| Дефектов найдено | 0 |

---

## Sprint Review Notes

- **US-8 реализована полностью** — бэкенд (DB + Handler + Router), фронтенд (API client + Store + CSS + UI компонент).
- **Каскадное удаление** работает корректно через `ON DELETE CASCADE` — тренировки удаляются автоматически вместе с челленджем.
- **Защита от случайного удаления** реализована через `confirm()` диалог с предупреждением.
- **User-scoped удаление** — нельзя удалить чужой челлендж (проверка `user_id`), возвращает `404`.
- **Playwright-тестирование** впервые применено в проекте через MCP-инструменты (без отдельного test runner).
