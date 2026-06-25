# Technical Architecture: Workout Challenge Tracker

## 1. System Architecture Overview (Обзор архитектуры системы)

Система построена на принципах слабой связанности (Loose Coupling) между клиентом и сервером. Взаимодействие происходит по протоколу HTTP через REST API с обменом данными в формате JSON.

```mermaid
graph LR
    subgraph Frontend [Клиентский слой]
        UI[Vanilla HTML/CSS/JS] --> Client[API Client]
        Store[Pub-Sub Store] --> UI
    end
    
    subgraph Backend [Серверный слой]
        API[Go API Server] --> Handlers[Handlers / Routing]
        Handlers --> DB_Client[Database Layer / pgx]
        Worker[Go Cron Worker] --> DB_Client
    end
    
    subgraph Storage [Слой хранения]
        DB[(PostgreSQL)]
    end
    
    Client -->|HTTP REST / JSON / X-User-Id| API
    DB_Client -->|SQL Queries| DB


---

## 2. Directory Tree (Дерево директорий)

workout-challenge-app/
├── .agents/                     # Инструкции и промпты для AI-агентов команды разработки
│   ├── analyst.md
│   ├── architect.md
│   ├── backend_dev.md
│   ├── devops.md
│   ├── frontend_dev.md
│   ├── qa_engineer.md
│   └── scrum_master.md
├── docs/
│   ├── spec.md                  # Функциональная спецификация
│   ├── architecture.md          # Техническая архитектура (этот файл)
│   ├── kanban.md                # Канбан-доска проекта
│   ├── tasks_backend.md         # Задачи для бэкенд-разработки
│   ├── tasks_frontend.md        # Задачи для фронтенд-разработки
│   └── tasks_qa.md              # Задачи и тест-кейсы для тестирования (QA)
├── docker-compose.yml           # Конфигурация для запуска PostgreSQL
├── README.md                    # Краткое описание проекта и стек технологий
├── backend/
│   ├── main.go                  # Точка входа приложения Go
│   ├── go.mod                   
│   ├── go.sum
│   ├── internal/
│   │   ├── config/              # Конфигурация приложения
│   │   ├── database/            # Подключение к БД, транзакции
│   │   ├── models/              # Структуры данных (Go structs)
│   │   ├── handlers/            # Обработчики API запросов (эндпоинты)
│   │   └── workers/             # Фоновые процессы (пересчет failed статусов)
├── frontend/
│   ├── index.html               # Единственная HTML страница (SPA)
│   ├── css/
│   │   └── main.css             # Глобальные стили и CSS-переменные Telegram
│   ├── js/
│   │   ├── app.js               # Инициализация приложения
│   │   ├── router.js            # SPA-роутер
│   │   ├── store.js             # Pub-Sub стейт-менеджер
│   │   ├── api.js               # API Клиент 
│   │   └── components/          # Компоненты UI 
│   │       ├── dashboard/
│   │       ├── challenge/
│   │       └── ui/              # Общие UI элементы (модалки, тосты)


---

## 3. Database Schema (DDL) (Схема базы данных)

При запуске бэкенда автоматически выполняются SQL-запросы для создания таблиц.

-- Таблица упражнений
CREATE TABLE IF NOT EXISTS exercises (
    id SERIAL PRIMARY KEY,
    user_id VARCHAR(100) NOT NULL, 
    name VARCHAR(100) NOT NULL,
    is_custom BOOLEAN DEFAULT FALSE,
    CONSTRAINT unique_user_exercise UNIQUE(user_id, name)
);

-- Таблица челленджей (с денормализованным полем current_progress)
CREATE TABLE IF NOT EXISTS challenges (
    id SERIAL PRIMARY KEY,
    user_id VARCHAR(100) NOT NULL,
    name VARCHAR(150) NOT NULL,
    exercise_id INT NOT NULL,
    target_value INT NOT NULL CHECK (target_value > 0),
    current_progress INT DEFAULT 0, -- Кэшированный прогресс
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    status VARCHAR(20) DEFAULT 'active' CHECK (status IN ('active', 'completed', 'failed')),
    CONSTRAINT fk_exercise FOREIGN KEY (exercise_id) REFERENCES exercises(id) ON DELETE CASCADE,
    CONSTRAINT check_dates CHECK (end_date >= start_date)
);

-- Таблица тренировок (логов выполнения)
CREATE TABLE IF NOT EXISTS workouts (
    id SERIAL PRIMARY KEY,
    user_id VARCHAR(100) NOT NULL,
    challenge_id INT NOT NULL,
    workout_date DATE NOT NULL,
    value INT NOT NULL CHECK (value > 0),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_challenge FOREIGN KEY (challenge_id) REFERENCES challenges(id) ON DELETE CASCADE
);

-- Таблица достижений пользователей
CREATE TABLE IF NOT EXISTS user_achievements (
    id SERIAL PRIMARY KEY,
    user_id VARCHAR(100) NOT NULL,
    achievement_code VARCHAR(50) NOT NULL, -- 'first_step', 'equator', 'hero', 'stability'
    unlocked_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT unique_user_achievement UNIQUE(user_id, achievement_code)
);

-- Индексы для оптимизации
CREATE INDEX IF NOT EXISTS idx_challenges_user ON challenges(user_id);
CREATE INDEX IF NOT EXISTS idx_workouts_challenge ON workouts(challenge_id);
CREATE INDEX IF NOT EXISTS idx_achievements_user ON user_achievements(user_id);

---

## 4. Backend Architecture (Архитектура бэкенда)

### Целостность данных (Data Integrity)

* **Транзакции:** Любой `POST` или `DELETE` запрос в `/api/workouts/` использует SQL-транзакции. Внутри транзакции добавляется запись в `workouts` и одновременно обновляется поле `current_progress` в таблице `challenges`. Если новая сумма `current_progress` >= `target_value`, `status` меняется на `completed`.
* **Фоновый Воркер (Cron Job):** В пакете `internal/workers/` запускается горутина (через библиотеку `robfig/cron`), которая раз в час выполняет запрос:
`UPDATE challenges SET status = 'failed' WHERE status = 'active' AND end_date < CURRENT_DATE AND current_progress < target_value;`

### Система Ачивок

При добавлении тренировки бэкенд проверяет триггеры ачивок. Если условия выполнены (например, прогресс достиг 50%), бэкенд делает `INSERT` в таблицу `user_achievements` (если такой записи еще нет). Коды новых ачивок добавляются в массив `unlocked_achievements` в JSON-ответе.

---

## 5. Frontend Architecture (Архитектура фронтенда)

### Управление состоянием (Store + Pub-Sub)

Для предотвращения прямой связи между компонентами используется централизованный объект `Store`. Компоненты подписываются на изменения, а при действиях пользователя отправляют события, которые обновляют Store и уведомляют подписчиков.

### Стилизация и Telegram Mini App

* Использование исключительно Vanilla CSS с упором на нативные CSS Variables (например, `var(--tg-theme-bg-color)`).
* Это обеспечит бесшовную интеграцию с темами Telegram (Light/Dark mode) без необходимости переписывать классы или бороться со сторонними UI-библиотеками.

---

## 6. Security & Integration Points (Безопасность)

### MVP-авторизация (Передача User ID)

Все HTTP-запросы от фронтенда включают заголовок `X-User-Id: default_user_1`. Бэкенд считывает его и использует для фильтрации данных во всех SQL-запросах.

### Переход на Telegram (Будущая интеграция)

1. Фронтенд извлекает строку инициализации `initData` через Telegram WebApp API.
2. Передает ее в заголовке `Authorization: Bearer <initData>`.
3. Бэкенд (Go) валидирует подпись с помощью секрета бота (HMAC-SHA256) и извлекает реальный Telegram ID для выполнения запросов к БД.
