package database

import (
	"context"
	"fmt"
	"log"

	"workout-challenge-app/internal/models"
)

// CreateWorkout inserts a workout and updates the challenge progress transactionally.
func (db *DBWrapper) CreateWorkout(ctx context.Context, userID string, challengeID int, workout *models.Workout) (*models.Workout, int, int, error) {
	tx, err := db.Pool.Begin(ctx)
	if err != nil {
		log.Printf("CreateWorkout: failed to begin tx: %v\n", err)
		return nil, 0, 0, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Step 1: Check challenge
	var id int
	err = tx.QueryRow(ctx, "SELECT id FROM challenges WHERE id = $1 AND user_id = $2 AND status = 'active' FOR UPDATE", challengeID, userID).Scan(&id)
	if err != nil {
		log.Printf("CreateWorkout: challenge not found or not active: %v\n", err)
		return nil, 0, 0, fmt.Errorf("challenge not found or not active: %w", err)
	}

	// Step 2: Insert workout
	err = tx.QueryRow(ctx, `
		INSERT INTO workouts (user_id, challenge_id, workout_date, value)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at
	`, userID, challengeID, workout.WorkoutDate, workout.Value).Scan(&workout.ID, &workout.CreatedAt)
	if err != nil {
		log.Printf("CreateWorkout: failed to insert workout: %v\n", err)
		return nil, 0, 0, fmt.Errorf("failed to insert workout: %w", err)
	}
	workout.UserID = userID
	workout.ChallengeID = challengeID

	// Step 3: Update progress
	var newProgress, targetValue int
	var newStatus string
	err = tx.QueryRow(ctx, `
		UPDATE challenges
		SET current_progress = current_progress + $1,
		    status = CASE
		        WHEN current_progress + $1 >= target_value THEN 'completed'
		        ELSE status
		    END
		WHERE id = $2
		RETURNING current_progress, target_value, status
	`, workout.Value, challengeID).Scan(&newProgress, &targetValue, &newStatus)
	if err != nil {
		log.Printf("CreateWorkout: failed to update challenge: %v\n", err)
		return nil, 0, 0, fmt.Errorf("failed to update challenge: %w", err)
	}

	// Step 4: Commit
	if err := tx.Commit(ctx); err != nil {
		log.Printf("CreateWorkout: failed to commit tx: %v\n", err)
		return nil, 0, 0, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return workout, newProgress, targetValue, nil
}

// DeleteWorkout removes a workout and reverts the challenge progress and status.
func (db *DBWrapper) DeleteWorkout(ctx context.Context, userID string, workoutID int) (*models.Challenge, error) {
	tx, err := db.Pool.Begin(ctx)
	if err != nil {
		log.Printf("DeleteWorkout: failed to begin tx: %v\n", err)
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Step 1: Get workout
	var challengeID, value int
	err = tx.QueryRow(ctx, "SELECT challenge_id, value FROM workouts WHERE id = $1 AND user_id = $2", workoutID, userID).Scan(&challengeID, &value)
	if err != nil {
		log.Printf("DeleteWorkout: workout not found: %v\n", err)
		return nil, fmt.Errorf("workout not found: %w", err)
	}

	// Step 2: Delete workout
	_, err = tx.Exec(ctx, "DELETE FROM workouts WHERE id = $1 AND user_id = $2", workoutID, userID)
	if err != nil {
		log.Printf("DeleteWorkout: failed to delete workout: %v\n", err)
		return nil, fmt.Errorf("failed to delete workout: %w", err)
	}

	// Step 3: Update challenge
	var challenge models.Challenge
	err = tx.QueryRow(ctx, `
		UPDATE challenges
		SET current_progress = GREATEST(current_progress - $1, 0),
		    status = CASE
		        WHEN status = 'completed' AND (current_progress - $2) < target_value THEN 'active'
		        ELSE status
		    END
		WHERE id = $3
		RETURNING id, user_id, name, exercise_id, target_value, current_progress, start_date, end_date, status
	`, value, value, challengeID).Scan(
		&challenge.ID,
		&challenge.UserID,
		&challenge.Name,
		&challenge.ExerciseID,
		&challenge.TargetValue,
		&challenge.CurrentProgress,
		&challenge.StartDate,
		&challenge.EndDate,
		&challenge.Status,
	)
	if err != nil {
		log.Printf("DeleteWorkout: failed to update challenge: %v\n", err)
		return nil, fmt.Errorf("failed to update challenge: %w", err)
	}

	// Step 4: Commit
	if err := tx.Commit(ctx); err != nil {
		log.Printf("DeleteWorkout: failed to commit tx: %v\n", err)
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &challenge, nil
}

// GetWorkoutsByChallenge returns all workouts for a challenge, sorted by date.
func (db *DBWrapper) GetWorkoutsByChallenge(ctx context.Context, userID string, challengeID int) ([]models.Workout, error) {
	rows, err := db.Pool.Query(ctx, `
		SELECT id, user_id, challenge_id, workout_date, value, created_at
		FROM workouts
		WHERE challenge_id = $1 AND user_id = $2
		ORDER BY workout_date DESC, created_at DESC
	`, challengeID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query workouts: %w", err)
	}
	defer rows.Close()

	workouts := []models.Workout{} // must not be nil
	for rows.Next() {
		var w models.Workout
		if err := rows.Scan(&w.ID, &w.UserID, &w.ChallengeID, &w.WorkoutDate, &w.Value, &w.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan workout: %w", err)
		}
		workouts = append(workouts, w)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return workouts, nil
}
