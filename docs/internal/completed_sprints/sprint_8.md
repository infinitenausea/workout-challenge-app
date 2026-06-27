# Kanban Board

## 🚀 Спринт №8 (Завершен)
**Цель спринта:** Завершить интеграцию приложения с Telegram Mini Apps SDK, реализовав нативную навигацию, тактильный отклик (haptic feedback) и полную поддержку визуальных тем (светлая/темная) для бесшовного пользовательского опыта.

---

### TO DO

---

### IN PROGRESS

---

### QA / REVIEW

---

### DONE

- [x] **[QA] Epic: US-11 Тактильный отклик (Haptic Feedback)**
  * **TC-11.1 (Positive):** Вибрация при успешном добавлении тренировки (`notificationOccurred('success')`).
  * **TC-11.2 (Negative):** Вибрация при ошибке валидации (`notificationOccurred('error')`).
  * **TC-11.3 (Positive):** Вибрация при удалении тренировки (`impactOccurred('medium')`).
  * **TC-11.4 (Edge):** Отсутствие крашей вне Telegram.

- [x] **[QA] Epic: US-12 Нативная навигация и закрытие приложения**
  * **TC-12.1 (Positive):** Отображение BackButton вне Дашборда.
  * **TC-12.2 (Positive):** Скрытие BackButton на Дашборде.
  * **TC-12.3 (Positive):** Включение защиты от закрытия при вводе в форму челленджа.
  * **TC-12.4 (Positive):** Отключение защиты после сохранения/отмены.

- [x] **[Frontend] Задача 19: Интеграция Haptic Feedback через Telegram SDK**
  * **Файлы:** `frontend/js/telegram.js`, `frontend/js/components/ui/workout-modal.js`, `frontend/js/components/ui/achievement-popup.js`, `frontend/js/components/challenge/challenge-detail.js`
  * **Описание:**
    1. В `js/telegram.js` реализованы безопасные обертки для вызовов `Telegram.WebApp.HapticFeedback`.
    2. В `workout-modal.js` добавлен вызов `telegram.triggerNotification('success')` при успешном добавлении тренировки и `telegram.triggerNotification('error')` при ошибке валидации/API.
    3. В `achievement-popup.js` при показе поп-апа с ачивкой вызывается `telegram.triggerNotification('success')`.
    4. В `challenge-detail.js` удаление тренировки подтверждается с вызовом `telegram.triggerImpact('medium')`, а успешное удаление челленджа возвращает `telegram.triggerNotification('success')`.

- [x] **[Frontend] Задача 20: Нативная навигация (BackButton) и защита от закрытия (ClosingConfirmation)**
  * **Файлы:** `frontend/js/telegram.js`, `frontend/js/router.js`, `frontend/js/components/challenge/challenge-form.js`, `frontend/js/components/ui/workout-modal.js`
  * **Описание:**
    1. В `js/telegram.js` реализованы безопасные обертки для BackButton и ClosingConfirmation.
    2. В `router.js` добавлено управление видимостью BackButton: кнопка показывается на страницах деталей и формы и корректно отключается (со сбросом обработчиков через `offClick`) при переходе на дашборд.
    3. Добавлена защита от закрытия Mini App (`enableClosingConfirmation`) при редактировании форм в `challenge-form.js` и `workout-modal.js`, которая отключается после успешного сохранения или отмены.
