package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"workout-challenge-app/internal/models"
)

// CreateChallenge inserts a new challenge into the database
func (db *DBWrapper) CreateChallenge(ctx context.Context, userID string, challenge *models.Challenge) error {
	query := `
		INSERT INTO challenges (user_id, name, exercise_id, target_value, start_date, end_date, status, current_progress)
		VALUES ($1, $2, $3, $4, $5, $6, 'active', 0)
		RETURNING id, status, current_progress
	`

	err := db.Pool.QueryRow(
		ctx,
		query,
		userID,
		challenge.Name,
		challenge.ExerciseID,
		challenge.TargetValue,
		challenge.StartDate,
		challenge.EndDate,
	).Scan(&challenge.ID, &challenge.Status, &challenge.CurrentProgress)

	if err != nil {
		return fmt.Errorf("failed to create challenge: %w", err)
	}

	challenge.UserID = userID
	return nil
}

// GetChallenges retrieves all challenges for a specific user
func (db *DBWrapper) GetChallenges(ctx context.Context, userID string) ([]models.Challenge, error) {
	query := `
		SELECT id, user_id, name, exercise_id, target_value, current_progress, start_date, end_date, status
		FROM challenges
		WHERE user_id = $1
		ORDER BY start_date DESC
	`

	rows, err := db.Pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query challenges: %w", err)
	}
	defer rows.Close()

	var challenges []models.Challenge
	for rows.Next() {
		var c models.Challenge
		err := rows.Scan(
			&c.ID,
			&c.UserID,
			&c.Name,
			&c.ExerciseID,
			&c.TargetValue,
			&c.CurrentProgress,
			&c.StartDate,
			&c.EndDate,
			&c.Status,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan challenge: %w", err)
		}
		challenges = append(challenges, c)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	// Return empty slice instead of nil to play nice with JSON encoding
	if challenges == nil {
		challenges = []models.Challenge{}
	}

	return challenges, nil
}

// GetChallengeByID retrieves a specific challenge by its ID for a user
func (db *DBWrapper) GetChallengeByID(ctx context.Context, userID string, id int) (*models.Challenge, error) {
	query := `
		SELECT id, user_id, name, exercise_id, target_value, current_progress, start_date, end_date, status
		FROM challenges
		WHERE id = $1 AND user_id = $2
	`

	var c models.Challenge
	err := db.Pool.QueryRow(ctx, query, id, userID).Scan(
		&c.ID,
		&c.UserID,
		&c.Name,
		&c.ExerciseID,
		&c.TargetValue,
		&c.CurrentProgress,
		&c.StartDate,
		&c.EndDate,
		&c.Status,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to query challenge by ID: %w", err)
	}

	return &c, nil
}

// DeleteChallenge deletes a challenge from the database.
// If the challenge does not exist or is not owned by the user, returns pgx.ErrNoRows.
func (db *DBWrapper) DeleteChallenge(ctx context.Context, userID string, id int) error {
	tag, err := db.Pool.Exec(ctx, `
		DELETE FROM challenges
		WHERE id = $1 AND user_id = $2
	`, id, userID)
	if err != nil {
		return fmt.Errorf("failed to delete challenge: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

// UpdateChallenge updates specific fields of a challenge.
func (db *DBWrapper) UpdateChallenge(ctx context.Context, userID string, id int, name *string, targetValue *int, startDate *time.Time, endDate *time.Time, status *string) error {
	query := `
		UPDATE challenges
		SET name = COALESCE($3, name),
		    target_value = COALESCE($4, target_value),
		    start_date = COALESCE($5, start_date),
		    end_date = COALESCE($6, end_date),
		    status = COALESCE($7, status)
		WHERE id = $1 AND user_id = $2
	`

	tag, err := db.Pool.Exec(ctx, query, id, userID, name, targetValue, startDate, endDate, status)
	if err != nil {
		return fmt.Errorf("failed to update challenge: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}
