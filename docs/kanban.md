# Kanban Board

## 🚀 Спринт №8 (Планирование)
**Цель спринта:** Завершить интеграцию приложения с Telegram Mini Apps SDK, реализовав нативную навигацию, тактильный отклик (haptic feedback) и полную поддержку визуальных тем (светлая/темная) для бесшовного пользовательского опыта.

---

### TO DO

#### [Frontend] Задача 18: Адаптация CSS под темы Telegram (US-10)
* **Файлы:** `frontend/css/main.css`, `frontend/js/telegram.js`
* **Описание:**
  1. В `main.css` убедиться, что **все** цвета приложения используют CSS-переменные Telegram (`var(--tg-theme-bg-color)`, `var(--tg-theme-text-color)`, `var(--tg-theme-button-color)`, `var(--tg-theme-button-text-color)`, `var(--tg-theme-hint-color)` и т.д.).
  2. Добавить `fallback` значения для локальной разработки в секцию `:root`. Например: `--app-bg: var(--tg-theme-bg-color, #ffffff);` и использовать `--app-bg` в стилях.
  3. В `js/telegram.js` добавить слушатель события `themeChanged` (через `Telegram.WebApp.onEvent('themeChanged', callback)`), чтобы принудительно перерисовывать специфичные элементы, если это потребуется (например, графики, если они отрисованы на Canvas и не обновляются через CSS).
* **Ограничения:**
  * Избегать захардкоженных hex или rgb цветов в компонентах.
  * Тщательно проверить контрастность текстов на светлой и темной теме.

#### [Frontend] Задача 19: Интеграция Haptic Feedback через Telegram SDK
* **Файлы:** `frontend/js/telegram.js`, `frontend/js/components/ui/workout-modal.js`, `frontend/js/components/ui/achievement-popup.js`, `frontend/js/components/challenge/challenge-detail.js`
* **Описание:**
  1. В `js/telegram.js` реализовать обертки для вызовов `Telegram.WebApp.HapticFeedback`:
     * `triggerImpact(style)` — вызов `impactOccurred(style)`, где style может быть `'light'`, `'medium'`, `'heavy'`, `'rigid'`, `'soft'`.
     * `triggerNotification(type)` — вызов `notificationOccurred(type)`, где type — `'error'`, `'success'`, `'warning'`.
     * `triggerSelection()` — вызов `selectionChanged()`.
  2. В `workout-modal.js`: 
     * При успешном добавлении тренировки вызывать `telegram.triggerNotification('success')`.
     * При ошибке валидации (пустое поле, отрицательное число) или ошибке API вызывать `telegram.triggerNotification('error')`.
  3. В `achievement-popup.js`:
     * При показе поп-апа с ачивкой вызывать `telegram.triggerNotification('success')` (в идеале с небольшой задержкой для синхронизации с анимацией).
  4. В `challenge-detail.js`:
     * При нажатии на кнопку удаления тренировки и подтверждении удалять с `telegram.triggerImpact('medium')`.
     * При успешном удалении челленджа: `telegram.triggerNotification('success')`.
* **Ограничения:**
  * Обертки в `telegram.js` должны проверять доступность метода `Telegram.WebApp.HapticFeedback`, чтобы не ломать приложение при локальной разработке в браузере (через `if (window.Telegram?.WebApp?.HapticFeedback) { ... }`).
  * Не злоупотреблять вибрацией (например, не вешать на каждый клик мыши/тап).

#### [Frontend] Задача 20: Нативная навигация (BackButton) и защита от закрытия (ClosingConfirmation)
* **Файлы:** `frontend/js/telegram.js`, `frontend/js/router.js`, `frontend/js/app.js`, `frontend/js/components/challenge/challenge-form.js`
* **Описание:**
  1. **BackButton:**
     * В `js/telegram.js` создать методы `showBackButton(onClick)` and `hideBackButton()`.
     * Метод `showBackButton` должен вызывать `Telegram.WebApp.BackButton.show()` и назначать обработчик `Telegram.WebApp.BackButton.onClick(onClick)`.
     * Метод `hideBackButton` вызывает `Telegram.WebApp.BackButton.hide()` и снимает обработчик через `offClick`.
  2. **Интеграция с роутером (`js/router.js` или логика навигации):**
     * При переходе на любой экран, кроме `dashboard` (например, `challenge-detail` или `challenge-form`), вызывать `telegram.showBackButton(() => store.navigate('dashboard'))`.
     * При возврате на `dashboard` обязательно вызывать `telegram.hideBackButton()`.
  3. **Защита от случайного закрытия (Closing Confirmation):**
     * В `js/telegram.js` добавить методы `enableClosingConfirmation()` и `disableClosingConfirmation()` (обертки над `Telegram.WebApp.enableClosingConfirmation()`).
     * В `challenge-form.js` повесить слушатель `input` на форму. Как только пользователь ввел любой текст в поля, вызывать `telegram.enableClosingConfirmation()`.
     * При успешном сохранении челленджа (перед переходом на дашборд) или при нажатии кнопки отмены вызывать `telegram.disableClosingConfirmation()`.
* **Ограничения:**
  * Безопасные вызовы: все методы `telegram.js` должны проверять инициализацию SDK.
  * Убедиться, что при нажатии на BackButton не только происходит переход на дашборд, но и очищается состояние формы, а также отключается подтверждение закрытия (`disableClosingConfirmation`).

#### [QA] Epic: US-11 Тактильный отклик (Haptic Feedback)
**Цель:** Проверить, что приложение вызывает методы Haptic Feedback при правильных сценариях.

##### UI/UX Тестирование (Telegram Emulator / Mock)
1. **TC-11.1 (Positive): Вибрация при успешном добавлении тренировки**
   * *Steps:* Добавить тренировку, перехватив вызовы `Telegram.WebApp.HapticFeedback.notificationOccurred`.
   * *Expected:* Метод вызывается с аргументом `'success'`.
2. **TC-11.2 (Negative): Вибрация при ошибке валидации**
   * *Steps:* Отправить форму с отрицательным количеством, перехватить вызовы HapticFeedback.
   * *Expected:* Метод вызывается с аргументом `'error'`.
3. **TC-11.3 (Positive): Вибрация при удалении (Impact)**
   * *Steps:* Подтвердить удаление тренировки в диалоге.
   * *Expected:* Вызывается `impactOccurred('medium')`.
4. **TC-11.4 (Edge): Отсутствие крашей вне Telegram**
   * *Steps:* Запустить приложение в обычном браузере, выполнить действия (добавление, удаление).
   * *Expected:* Приложение работает штатно, ошибки о `Telegram.WebApp.HapticFeedback is undefined` в консоли нет.

#### [QA] Epic: US-12 Нативная навигация и закрытие приложения
**Цель:** Проверить интеграцию BackButton и ClosingConfirmation.

##### UI/UX Тестирование (Telegram Emulator / Mock)
1. **TC-12.1 (Positive): Отображение BackButton вне Дашборда**
   * *Steps:* Перейти на страницу деталей челленджа.
   * *Expected:* Вызывается `Telegram.WebApp.BackButton.show()`.
2. **TC-12.2 (Positive): Скрытие BackButton на Дашборде**
   * *Steps:* Вернуться с деталей на дашборд.
   * *Expected:* Вызывается `Telegram.WebApp.BackButton.hide()`.
3. **TC-12.3 (Positive): Включение защиты от закрытия при вводе**
   * *Steps:* Открыть форму создания челленджа, начать вводить текст.
   * *Expected:* Вызывается `Telegram.WebApp.enableClosingConfirmation()`.
4. **TC-12.4 (Positive): Отключение защиты после сохранения/отмены**
   * *Steps:* Успешно сохранить челлендж или вернуться на дашборд.
   * *Expected:* Вызывается `Telegram.WebApp.disableClosingConfirmation()`.

---

### IN PROGRESS

---

### QA / REVIEW

---

### DONE

---
