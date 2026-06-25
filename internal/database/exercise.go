package database

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"workout-challenge-app/internal/models"
)

var ErrDuplicateExercise = errors.New("exercise already exists for this user")

// CreateExercise inserts a new custom exercise for a user
func (db *DBWrapper) CreateExercise(ctx context.Context, userID, name string) (*models.Exercise, error) {
	query := `
		INSERT INTO exercises (user_id, name, is_custom)
		VALUES ($1, $2, true)
		RETURNING id, user_id, name, is_custom
	`

	var exercise models.Exercise
	err := db.Pool.QueryRow(ctx, query, userID, name).Scan(
		&exercise.ID,
		&exercise.UserID,
		&exercise.Name,
		&exercise.IsCustom,
	)

	if err != nil {
		// pgx does not have a specific error for unique constraint violation out-of-the-box
		// without checking the pgconn.PgError code (23505).
		// For simplicity without importing pgconn, we check the error string.
		if strings.Contains(err.Error(), "SQLSTATE 23505") || strings.Contains(err.Error(), "duplicate key value") {
			return nil, ErrDuplicateExercise
		}
		return nil, fmt.Errorf("failed to create exercise: %w", err)
	}

	return &exercise, nil
}

// GetExercises retrieves all exercises available to a user
// (assuming global exercises have a specific user_id like 'system' or we just return all for the user)
// Based on the spec, user can have their own custom exercises. If there are default ones, we would
// also fetch where user_id = 'default' OR user_id = $1. For now, we fetch all where user_id = $1.
func (db *DBWrapper) GetExercises(ctx context.Context, userID string) ([]models.Exercise, error) {
	query := `
		SELECT id, user_id, name, is_custom 
		FROM exercises 
		WHERE user_id = $1 OR user_id = 'system'
		ORDER BY id ASC
	`
	rows, err := db.Pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	var exercises []models.Exercise
	for rows.Next() {
		var e models.Exercise
		if err := rows.Scan(&e.ID, &e.UserID, &e.Name, &e.IsCustom); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		exercises = append(exercises, e)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	// Return empty slice instead of nil for better JSON serialization
	if exercises == nil {
		exercises = []models.Exercise{}
	}

	return exercises, nil
}
