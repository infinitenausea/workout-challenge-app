# Frontend Tasks

## Epic: US-1 Создание упражнения

**Цель:** Реализовать UI для добавления кастомных упражнений в форме создания челленджа.

### Задача 1: API Client
* **Файл:** `js/api.js`
* **Описание:**
  1. Добавить метод `getExercises()`, делающий `GET /api/exercises`.
  2. Добавить метод `createExercise(name)`, делающий `POST /api/exercises` с телом `{ "name": "..." }`.
* **Ограничения:**
  * Оба метода должны автоматически прикреплять заголовки `X-User-Id: default_user_1` и `Content-Type: application/json`.

### Задача 2: State Management (Store)
* **Файл:** `js/store.js`
* **Описание:**
  1. Добавить в начальное состояние массив `exercises: []`.
  2. При инициализации приложения вызывать `api.getExercises()` и обновлять Store.

### Задача 3: UI Component (Challenge Form)
* **Файл:** `js/components/challenge/challenge-form.js`, `challenge-form.css`
* **Описание:**
  1. Отрендерить выпадающий список (`<select>`) с упражнениями из Store.
  2. В конец списка добавить опцию `value="custom"` с текстом «Добавить свое...».
  3. Логика переключения: повесить слушатель `change` на селект. Если выбрано «Добавить свое», показывать скрытое ранее текстовое поле `<input type="text" id="custom-exercise-name">`. Иначе — скрывать.
  4. Логика сохранения (при сабмите формы создания челленджа):
     * Если селект в режиме "custom", сначала вызвать `api.createExercise(customName)`.
     * Обработать ответ: если API вернул ошибку 400 (пустое имя) или 409 (уже существует), показать пользователю уведомление (alert или текст под инпутом) и прервать отправку.
     * Если успешно: обновить Store новым списком упражнений, выбрать созданное упражнение в селекте и продолжить логику создания челленджа.
* **Ограничения:**
  * Никаких фреймворков. Использовать Vanilla JS и `addEventListener`.
  * Стилизовать инпуты с использованием CSS Variables.

---

## Epic: US-2, US-5 Создание челленджа и Дашборд

**Цель:** Реализовать интеграцию создания челленджей на клиенте и отображение их списка на дашборде.

### Задача 4: API Client & Store for Challenges
* **Файлы:** `frontend/js/api.js`, `frontend/js/store.js`
* **Описание:**
  1. В `api.js` добавить методы:
     * `getChallenges()`: делает `GET /api/challenges`.
     * `createChallenge(payload)`: делает `POST /api/challenges` с телом челленджа.
  2. В `store.js` добавить:
     * Поле `challenges: []` в начальное состояние.
     * Метод `setChallenges(challenges)` для обновления списка в стейте.
     * Метод `addChallenge(challenge)` для реактивного добавления нового челленджа в список.

### Задача 5: Интеграция Формы и Отображение на Дашборде
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

---

## Epic: US-3, US-4, US-6, US-7 Логирование тренировки, Детали, Удаление и Ачивки

**Цель:** Реализовать UI для добавления тренировок, мгновенного пересчёта прогресса, удаления ошибочных записей и отображения поздравительных поп-апов при получении ачивок.

### Задача 6: API Client — Workout & Achievement Methods

* **Файл:** `frontend/js/api.js` **(MODIFY)**

* **Описание:**
  1. Добавить метод `createWorkout(challengeId, payload)`:
     * `POST /api/challenges/${challengeId}/workouts`
     * Тело: `{ "workout_date": "YYYY-MM-DD", "value": number }`
     * Ответ содержит `{ success, workout, unlocked_achievements }` — метод возвращает весь JSON-ответ целиком.
  2. Добавить метод `deleteWorkout(workoutId)`:
     * `DELETE /api/workouts/${workoutId}`
     * Ответ содержит `{ success, challenge }` — обновлённый объект челленджа.
  3. Добавить метод `getChallengeDetail(challengeId)`:
     * `GET /api/challenges/${challengeId}`
     * Возвращает детальную информацию о челлендже (включая список тренировок, если бэкенд их присоединяет, или без них — тогда нужен отдельный вызов).
  4. Добавить метод `getAchievements()`:
     * `GET /api/achievements`
     * Возвращает массив разблокированных ачивок пользователя.

* **Ограничения:**
  * Все методы используют существующий `_request()` — не дублировать логику.
  * Заголовок `X-User-Id` прикрепляется автоматически через базовый `_request`.

---

### Задача 7: Store — Состояние для Workouts и Achievements

* **Файл:** `frontend/js/store.js` **(MODIFY)**

* **Описание:**
  1. Добавить новые поля в начальное состояние `this.state`:
     ```js
     currentChallenge: null,  // Детали текущего просматриваемого челленджа
     workouts: [],            // Список тренировок текущего челленджа
     achievements: [],        // Разблокированные ачивки пользователя
     ```
  2. Добавить методы:
     * `setCurrentChallenge(challenge)` — устанавливает текущий просматриваемый челлендж.
     * `setWorkouts(workouts)` — устанавливает список тренировок.
     * `addWorkout(workout)` — добавляет одну тренировку в начало массива `workouts` (т.к. сортировка DESC по дате).
     * `removeWorkout(workoutId)` — удаляет тренировку из массива `workouts` по `id`.
     * `updateChallengeProgress(challengeId, newProgress, newStatus)` — обновляет `current_progress` и `status` конкретного челленджа в массиве `challenges` **и** в `currentChallenge` (если совпадает `id`). Это нужно для мгновенного пересчёта прогресс-бара без перезагрузки.
     * `setAchievements(achievements)` — устанавливает массив ачивок.
     * `addAchievements(newCodes)` — добавляет новые коды ачивок в массив (для мгновенного обновления UI после получения ачивки, без перезагрузки).

* **Ограничения:**
  * Все методы должны вызывать `this.setState(...)` / `this.notify()` для реактивного обновления UI.
  * Метод `updateChallengeProgress` обновляет **и** массив `challenges` (для дашборда), **и** `currentChallenge` (для страницы деталей). Иначе при возврате на дашборд прогресс-бар будет устаревшим.

---

### Задача 8: Компонент «Страница деталей челленджа» (Challenge Detail)

* **Файлы:**
  * `frontend/js/components/challenge/challenge-detail.js` **(NEW)**
  * `frontend/css/main.css` **(MODIFY — добавить стили)**

* **Описание:**
  1. **Создать класс `ChallengeDetail`** с методами `constructor(container)`, `mount()`, `unmount()`, `render()`.
  2. **При монтировании (`mount`):**
     * Получить `currentChallengeId` из `store.getState()`.
     * Вызвать `api.getChallengeDetail(id)` для получения данных челленджа с бэкенда.
     * Сохранить в стор: `store.setCurrentChallenge(challenge)`.
     * Запросить тренировки: если бэкенд не возвращает их в составе challenge, нужен отдельный запрос (определяется контрактом API). **Рекомендация:** добавить на бэкенде в ответ `GET /api/challenges/:id` массив `workouts` — это избавит от лишнего запроса. Если нет — запросить отдельно.
     * Подписаться на изменения стора.
  3. **Рендеринг (`render`) — что отображать:**

     **a) Заголовок:**
     * Кнопка «← Назад» для навигации на дашборд: `store.navigate('dashboard')`.
     * Название челленджа (`challenge.name`).
     * Статус-бейдж (`active` / `completed` / `failed`).

     **b) Прогресс-бар (крупный):**
     * Процент: `Math.min(100, Math.round((current_progress / target_value) * 100))`.
     * Текст: `"${current_progress} / ${target_value}"`.
     * Визуальный прогресс-бар (как на дашборде, но крупнее — высота 12px).

     **c) Таймер обратного отсчёта:**
     * Рассчитать разницу между `end_date` и текущей датой.
     * Формат: `"Осталось: X дней"` (если > 0) или `"Дедлайн истёк"` (если <= 0).
     * Если статус `completed` — показать текст `"Челлендж завершён! 🎉"`.

     **d) Кнопка «Добавить тренировку»:**
     * Открывает модальное окно (see Задача 9).
     * Кнопка **заблокирована**, если статус челленджа `completed` или `failed`.

     **e) История тренировок:**
     * Список карточек тренировок, отсортированных по `workout_date` DESC.
     * Каждая карточка показывает:
       * Дата тренировки (формат `DD.MM.YYYY`).
       * Количество повторений (`value`).
       * Иконка 🗑️ (корзина) для удаления.
     * Пустое состояние: `"Тренировок пока нет. Начните прямо сейчас!"`.

     **f) Обработка удаления тренировки:**
     * Повесить `click` слушатель на иконку 🗑️ каждой тренировки.
     * При клике:
       1. Показать `confirm('Удалить эту тренировку?')`.
       2. Если подтверждено — вызвать `api.deleteWorkout(workoutId)`.
       3. При успехе:
          * `store.removeWorkout(workoutId)`.
          * Из ответа API извлечь обновлённый `challenge` и вызвать `store.updateChallengeProgress(challenge.id, challenge.current_progress, challenge.status)`.
          * Также обновить `store.setCurrentChallenge(challenge)`.
       4. При ошибке — показать тост с сообщением.

  4. **CSS стили в `main.css`:**
     * `.challenge-detail` — основной контейнер.
     * `.challenge-detail-header` — flex-контейнер с кнопкой «Назад» и названием.
     * `.countdown-timer` — стилизация таймера (например, цвет `--danger-color` если < 3 дней).
     * `.workout-list-item` — карточка тренировки с flex-layout (дата + значение + иконка удаления).
     * `.add-workout-btn` — крупная кнопка для добавления тренировки.

* **Ограничения:**
  * Vanilla JS, никаких фреймворков.
  * При рендеринге НЕ перерисовывать весь контейнер, если изменились только `current_progress` или `workouts` — минимизировать мерцание. **Допустимый упрощённый подход:** перерисовка через `innerHTML` с сохранением scroll-позиции (для MVP это приемлемо).
  * Прогресс-бар должен обновляться **мгновенно** (AC-1 спецификации) — без перезагрузки страницы, через реактивность стора.

---

### Задача 9: Модальное окно «Добавить тренировку» (Workout Modal)

* **Файлы:**
  * `frontend/js/components/ui/workout-modal.js` **(NEW)**
  * `frontend/css/main.css` **(MODIFY — добавить стили модалки)**

* **Описание:**
  1. **Создать класс `WorkoutModal`** с методами:
     * `constructor()` — не привязан к конкретному контейнеру, будет рендериться в `body`.
     * `open(challengeId)` — показывает модальное окно.
     * `close()` — скрывает и очищает модалку.

  2. **HTML-разметка модалки:**
     ```html
     <div class="modal-overlay" id="workout-modal-overlay">
       <div class="modal-content">
         <div class="modal-header">
           <h3>Добавить тренировку</h3>
           <button class="modal-close-btn">&times;</button>
         </div>
         <form id="workout-form">
           <div class="form-group">
             <label for="workout_date">Дата тренировки</label>
             <input type="date" id="workout_date" name="workout_date" value="[TODAY]" required>
           </div>
           <div class="form-group">
             <label for="workout_value">Количество повторений</label>
             <input type="number" id="workout_value" name="value" min="1" placeholder="50" required>
           </div>
           <button type="submit">Сохранить</button>
         </form>
       </div>
     </div>
     ```

  3. **Логика отправки формы (`submit`):**
     * Извлечь `workout_date` и `value` из формы.
     * **Клиентская валидация:**
       * `value` > 0 — иначе тост `"Количество должно быть больше 0"`.
       * `workout_date` заполнено — иначе тост `"Укажите дату тренировки"`.
     * Заблокировать кнопку «Сохранить» на время запроса (`disabled = true`).
     * Вызвать `api.createWorkout(challengeId, { workout_date, value })`.
     * **При успехе:**
       1. Извлечь из ответа `workout` и `unlocked_achievements`.
       2. `store.addWorkout(response.workout)`.
       3. Обновить прогресс челленджа в сторе: нужно пересчитать `current_progress` (старый `current_progress` + `value`). **Лучше:** получить актуальные данные из ответа API — для этого бэкенд должен возвращать обновлённый `current_progress` и `status`. Вызвать `store.updateChallengeProgress(challengeId, newProgress, newStatus)`.
       4. Закрыть модалку: `this.close()`.
       5. Показать тост `"Тренировка добавлена! +{value} повторений"`.
       6. **Если `unlocked_achievements.length > 0`** — вызвать показ поздравительного поп-апа (см. Задача 10).
     * **При ошибке:**
       * Показать тост с текстом ошибки от сервера.
       * Разблокировать кнопку.

  4. **Закрытие модалки:**
     * Клик по оверлею (`.modal-overlay`) — закрыть.
     * Клик по кнопке `×` — закрыть.
     * Нажатие `Escape` — закрыть.

  5. **CSS стили:**
     * `.modal-overlay` — `position: fixed; top: 0; left: 0; width: 100%; height: 100%; background: rgba(0,0,0,0.6); z-index: 999; display: flex; align-items: center; justify-content: center;`
     * `.modal-content` — `background: var(--card-bg); border-radius: var(--border-radius); padding: 24px; max-width: 400px; width: 90%;`
     * `.modal-header` — flex-контейнер с заголовком и крестиком.
     * `.modal-close-btn` — стилизованная кнопка-крестик (background: none, большой font-size).
     * Анимация появления: `animation: fadeIn 0.2s ease-out`.

* **Ограничения:**
  * Модалка рендерится в `document.body`, а не в `#app`, чтобы не конфликтовать с роутером.
  * Не забывать удалять DOM-элемент модалки при `close()` и снимать `keydown` листенер на `Escape`.
  * Дата по умолчанию — сегодня (`new Date().toISOString().split('T')[0]`).

---

### Задача 10: Поздравительный поп-ап при получении ачивки (Achievement Popup)

* **Файлы:**
  * `frontend/js/components/ui/achievement-popup.js` **(NEW)**
  * `frontend/css/main.css` **(MODIFY — добавить стили)**

* **Описание:**
  1. **Создать функцию (или класс) `showAchievementPopup(achievementCodes)`:**
     * Принимает массив кодов ачивок: `["first_step", "equator"]`.
     * Для каждого кода определяет название и иконку из **локального маппинга**:
       ```js
       const ACHIEVEMENTS = {
         first_step: { name: 'Первый шаг',  icon: '🌱', description: 'Внесена первая тренировка' },
         equator:    { name: 'Экватор',      icon: '📈', description: 'Прогресс достиг 50%' },
         hero:       { name: 'Герой',        icon: '⚡', description: 'Челлендж завершён до дедлайна!' },
         stability:  { name: 'Стабильность', icon: '🔥', description: 'Тренировки 3 дня подряд' }
       };
       ```
     * Рендерит по центру экрана оверлей с **анимированной** карточкой:
       ```
       ┌───────────────────┐
       │    🌱              │
       │  Первый шаг!       │
       │  Внесена первая    │
       │  тренировка        │
       │                    │
       │  [ Отлично! ]      │
       └───────────────────┘
       ```
     * Если `achievementCodes` содержит несколько ачивок — показывать их **последовательно** (закрытие первого показывает следующий).

  2. **Анимация:**
     * Карточка появляется с эффектом `scale(0.8) → scale(1)` + `opacity 0 → 1` (300ms).
     * Иконка ачивки пульсирует (`@keyframes pulse { ... }`).
     * Фон — оверлей с `backdrop-filter: blur(4px)`.

  3. **Закрытие:**
     * Кнопка «Отлично!» закрывает поп-ап.
     * Клик по оверлею — закрывает.
     * Автозакрытие через 5 секунд (с прогресс-баром внизу карточки).

  4. **CSS стили:**
     * `.achievement-popup-overlay` — фиксированный оверлей с блюром.
     * `.achievement-popup-card` — карточка с gradient-бордером (для "премиум" ощущения):
       ```css
       .achievement-popup-card {
         background: var(--card-bg);
         border: 2px solid transparent;
         background-clip: padding-box;
         position: relative;
         border-radius: 16px;
         padding: 32px;
         text-align: center;
       }
       .achievement-popup-card::before {
         content: '';
         position: absolute;
         inset: -2px;
         border-radius: 18px;
         background: linear-gradient(135deg, #FFD700, #FFA500, #FF6347);
         z-index: -1;
       }
       ```
     * `.achievement-icon` — большая иконка (font-size: 64px) с `animation: pulse 1s ease infinite`.
     * `.achievement-auto-close-bar` — полоска автозакрытия внизу (анимация `width: 100% → 0%` за 5 секунд).

* **Ограничения:**
  * Поп-ап рендерится в `document.body` (как и workout-modal).
  * Очищать DOM и снимать таймеры при закрытии.
  * Если пользователь получил 2+ ачивки одновременно — не показывать их все разом. Очередь: закрыл один → показался следующий.

---

### Задача 11: Регистрация маршрута Challenge Detail и обновление дашборда

* **Файлы:**
  * `frontend/js/app.js` **(MODIFY)**
  * `frontend/js/components/dashboard/dashboard.js` **(MODIFY)**

* **Описание:**
  1. **В `app.js`:**
     * Импортировать `ChallengeDetail`:
       ```js
       import { ChallengeDetail } from './components/challenge/challenge-detail.js';
       ```
     * Добавить маршрут в объект `routes`:
       ```js
       const routes = {
         'dashboard': Dashboard,
         'challenge-form': ChallengeForm,
         'challenge-detail': ChallengeDetail
       };
       ```
     * При инициализации загрузить ачивки:
       ```js
       try {
         const achievements = await api.getAchievements();
         store.setAchievements(achievements);
       } catch (error) {
         console.error('Failed to load achievements:', error);
         store.setAchievements([]);
       }
       ```

  2. **В `dashboard.js`:**
     * **Сделать карточки челленджей кликабельными.** Добавить обработчик `click` на каждую `.challenge-card`:
       ```js
       const cards = this.container.querySelectorAll('.challenge-card');
       cards.forEach(card => {
         card.style.cursor = 'pointer';
         card.addEventListener('click', () => {
           const challengeId = parseInt(card.dataset.id);
           store.navigate('challenge-detail', { currentChallengeId: challengeId });
         });
       });
       ```
     * **Обновить секцию «Мои Достижения»:** вместо хардкода все 4 ачивки серыми — динамически подсвечивать те, которые есть в `state.achievements`:
       ```js
       const achievementDefs = [
         { code: 'first_step', icon: '🌱', name: 'Первый шаг' },
         { code: 'equator',    icon: '📈', name: 'Экватор' },
         { code: 'hero',       icon: '⚡', name: 'Герой' },
         { code: 'stability',  icon: '🔥', name: 'Стабильность' }
       ];

       const unlockedCodes = (state.achievements || []).map(a => a.achievement_code);

       const achievementsHTML = achievementDefs.map(def => {
         const isUnlocked = unlockedCodes.includes(def.code);
         return `
           <div style="opacity: ${isUnlocked ? '1' : '0.3'}; filter: ${isUnlocked ? 'none' : 'grayscale(100%)'}; transition: all 0.3s ease;">
             <div style="font-size: 24px;">${def.icon}</div>
             <div style="font-size: 10px; color: var(--tg-theme-hint-color);">${def.name}</div>
           </div>
         `;
       }).join('');
       ```
     * Вставить `achievementsHTML` вместо текущего захардкоженного блока ачивок.

* **Ограничения:**
  * Не ломать существующую навигацию между `dashboard` и `challenge-form`.
  * Карточки челленджей на дашборде должны получить визуальный фидбек при наведении (`cursor: pointer` + существующие hover-стили из CSS).

---

## Epic: US-8 Удаление челленджа

**Цель:** Реализовать UI для удаления челленджа на экране деталей.

### Задача 14: API Client

* **Файл:** `js/api.js`
* **Описание:** Добавить метод `deleteChallenge(challengeId)`, делающий `DELETE /api/challenges/:id`.

### Задача 15: State Management (Store)

* **Файл:** `js/store.js`
* **Описание:** Добавить метод `removeChallenge(challengeId)`, удаляющий челлендж из `state.challenges` и сбрасывающий `currentChallenge` и `currentChallengeId` если они совпадают.

### Задача 16: UI Component & CSS

* **Файл:** `js/components/challenge/challenge-detail.js`, `css/main.css`
* **Описание:**
  1. В `css/main.css` добавить стиль `button.danger` (прозрачный фон, цвет `--danger-color`, красная рамка, hover-эффект).
  2. В `challenge-detail.js` добавить метод `handleDeleteChallenge(challengeId)`:
     * Показать `confirm()` диалог с предупреждением.
     * При подтверждении вызвать `api.deleteChallenge(challengeId)`.
     * Вызвать `store.removeChallenge(challengeId)` и `store.navigate('dashboard')`.
     * Показать success-тост.
     * Обработать ошибку с error-тостом.
  3. В разметке `render()` добавить кнопку `<button id="delete-challenge-btn" class="danger">🗑️ Удалить челлендж</button>` под списком тренировок.
  4. В событийном блоке подключить обработчик клика на кнопку.

* **Ограничения:**
  * Кнопка должна быть стилизована как danger — красная, чтобы отличаться от остальных действий.
