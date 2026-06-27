package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// DBWrapper wraps the pgxpool connection
type DBWrapper struct {
	Pool *pgxpool.Pool
}

// Connect establishing a connection to the PostgreSQL database with retries
func Connect(dsn string) (*DBWrapper, error) {
	var pool *pgxpool.Pool
	var err error

	maxRetries := 3
	for i := 0; i < maxRetries; i++ {
		log.Printf("Attempting to connect to database (attempt %d/%d)...\n", i+1, maxRetries)
		pool, err = pgxpool.New(context.Background(), dsn)
		if err == nil {
			err = pool.Ping(context.Background())
			if err == nil {
				log.Println("Successfully connected to the database!")
				return &DBWrapper{Pool: pool}, nil
			}
		}
		log.Printf("Database connection failed: %v\n", err)
		time.Sleep(2 * time.Second)
	}

	return nil, fmt.Errorf("could not connect to database after %d attempts: %w", maxRetries, err)
}

// RunMigrations executes the initial DDL script to create tables and indexes
func (db *DBWrapper) RunMigrations() error {
	log.Println("Running database migrations...")

	ddl := `
	-- Таблица упражнений
	CREATE TABLE IF NOT EXISTS exercises (
		id SERIAL PRIMARY KEY,
		user_id VARCHAR(100) NOT NULL, 
		name VARCHAR(100) NOT NULL,
		is_custom BOOLEAN DEFAULT FALSE,
		CONSTRAINT unique_user_exercise UNIQUE(user_id, name)
	);

	-- Таблица челленджей
	CREATE TABLE IF NOT EXISTS challenges (
		id SERIAL PRIMARY KEY,
		user_id VARCHAR(100) NOT NULL,
		name VARCHAR(150) NOT NULL,
		exercise_id INT NOT NULL,
		target_value INT NOT NULL CHECK (target_value > 0),
		current_progress INT DEFAULT 0,
		start_date DATE NOT NULL,
		end_date DATE NOT NULL,
		status VARCHAR(20) DEFAULT 'active' CHECK (status IN ('active', 'completed', 'failed')),
		CONSTRAINT fk_exercise FOREIGN KEY (exercise_id) REFERENCES exercises(id) ON DELETE CASCADE,
		CONSTRAINT check_dates CHECK (end_date >= start_date)
	);

	-- Таблица тренировок
	CREATE TABLE IF NOT EXISTS workouts (
		id SERIAL PRIMARY KEY,
		user_id VARCHAR(100) NOT NULL,
		challenge_id INT NOT NULL,
		workout_date DATE NOT NULL,
		value INT NOT NULL CHECK (value > 0),
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		CONSTRAINT fk_challenge FOREIGN KEY (challenge_id) REFERENCES challenges(id) ON DELETE CASCADE
	);

	-- DROP таблицу достижений для накатывания новой схемы (временно для MVP/спринта)
	DROP TABLE IF EXISTS user_achievements;

	-- Таблица достижений пользователей
	CREATE TABLE IF NOT EXISTS user_achievements (
		id SERIAL PRIMARY KEY,
		user_id VARCHAR(100) NOT NULL,
		challenge_id INT NOT NULL,
		achievement_code VARCHAR(50) NOT NULL,
		unlocked_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		CONSTRAINT fk_achievements_challenge FOREIGN KEY (challenge_id) REFERENCES challenges(id) ON DELETE CASCADE,
		CONSTRAINT unique_user_challenge_achievement UNIQUE(user_id, challenge_id, achievement_code)
	);

	-- Индексы для оптимизации
	CREATE INDEX IF NOT EXISTS idx_challenges_user ON challenges(user_id);
	CREATE INDEX IF NOT EXISTS idx_workouts_challenge ON workouts(challenge_id);
	CREATE INDEX IF NOT EXISTS idx_achievements_user ON user_achievements(user_id);
	`

	_, err := db.Pool.Exec(context.Background(), ddl)
	if err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	log.Println("Database migrations completed successfully.")
	return nil
}

// Close closes the database pool
func (db *DBWrapper) Close() {
	if db.Pool != nil {
		db.Pool.Close()
	}
}
