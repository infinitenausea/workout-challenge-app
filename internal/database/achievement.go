package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"workout-challenge-app/internal/models"
)

// CheckAndUnlockAchievements checks all achievements and unlocks any new ones.
// It will not return an error if a single check fails; it logs the error and continues.
func (db *DBWrapper) CheckAndUnlockAchievements(ctx context.Context, userID string, challengeID int, newProgress, targetValue int) ([]string, error) {
	newlyUnlocked := []string{}

	// Helper function to unlock
	tryUnlock := func(code string) {
		unlocked, err := db.unlockAchievement(ctx, userID, challengeID, code)
		if err != nil {
			log.Printf("CheckAndUnlockAchievements: failed to unlock %s: %v\n", code, err)
		} else if unlocked {
			newlyUnlocked = append(newlyUnlocked, code)
		}
	}

	// 1. first_step: Внесена первая тренировка в этот челлендж.
	firstStepEligible := false
	err := db.Pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM workouts WHERE user_id = $1 AND challenge_id = $2 LIMIT 1)", userID, challengeID).Scan(&firstStepEligible)
	if err != nil {
		log.Printf("CheckAndUnlockAchievements: failed to check first_step eligibility: %v\n", err)
	} else if firstStepEligible {
		tryUnlock("first_step")
	}

	// 2. equator: newProgress * 2 >= targetValue
	if newProgress*2 >= targetValue {
		tryUnlock("equator")
	}

	// Fetch challenge end_date for hero and final_spurt
	var endDate time.Time
	err = db.Pool.QueryRow(ctx, "SELECT end_date FROM challenges WHERE id = $1", challengeID).Scan(&endDate)
	if err != nil {
		log.Printf("CheckAndUnlockAchievements: failed to get challenge end_date: %v\n", err)
	} else {
		now := time.Now()
		today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		endDateNorm := time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 0, 0, 0, 0, now.Location())

		// 3. hero: 100% and end_date not passed yet
		if newProgress >= targetValue && !endDateNorm.Before(today) {
			tryUnlock("hero")
		}

		// 8. final_spurt: 100% exactly on end_date
		if newProgress >= targetValue && endDateNorm.Equal(today) {
			tryUnlock("final_spurt")
		}
	}

	// 4. stability: тренировки вносились 3 дня подряд в этом челлендже.
	rows, err := db.Pool.Query(ctx, "SELECT DISTINCT workout_date FROM workouts WHERE user_id = $1 AND challenge_id = $2 ORDER BY workout_date DESC LIMIT 10", userID, challengeID)
	if err != nil {
		log.Printf("CheckAndUnlockAchievements: failed to query workout dates for stability check: %v\n", err)
	} else {
		var dates []time.Time
		for rows.Next() {
			var d time.Time
			if err := rows.Scan(&d); err == nil {
				dates = append(dates, d)
			}
		}
		rows.Close()

		stabilityEligible := false
		for i := 0; i <= len(dates)-3; i++ {
			d1 := dates[i]
			d2 := dates[i+1]
			d3 := dates[i+2]

			utc1 := time.Date(d1.Year(), d1.Month(), d1.Day(), 0, 0, 0, 0, time.UTC)
			utc2 := time.Date(d2.Year(), d2.Month(), d2.Day(), 0, 0, 0, 0, time.UTC)
			utc3 := time.Date(d3.Year(), d3.Month(), d3.Day(), 0, 0, 0, 0, time.UTC)

			if utc1.Sub(utc2) == 24*time.Hour && utc2.Sub(utc3) == 24*time.Hour {
				stabilityEligible = true
				break
			}
		}

		if stabilityEligible {
			tryUnlock("stability")
		}
	}

	// Get latest workout for power_start and early_bird
	var lastWorkoutValue int
	var lastWorkoutCreatedAt time.Time
	err = db.Pool.QueryRow(ctx, "SELECT value, created_at FROM workouts WHERE user_id = $1 AND challenge_id = $2 ORDER BY created_at DESC LIMIT 1", userID, challengeID).Scan(&lastWorkoutValue, &lastWorkoutCreatedAt)
	if err != nil {
		log.Printf("CheckAndUnlockAchievements: failed to get last workout: %v\n", err)
	} else {
		// 5. power_start: >= 25% of targetValue
		if float64(lastWorkoutValue) >= float64(targetValue)*0.25 {
			tryUnlock("power_start")
		}

		// 7. early_bird: added between 5:00 and 8:59
		hour := lastWorkoutCreatedAt.Hour()
		if hour >= 5 && hour < 9 {
			tryUnlock("early_bird")
		}
	}

	// 6. overachiever: >= 120% of targetValue
	if float64(newProgress) >= float64(targetValue)*1.2 {
		tryUnlock("overachiever")
	}

	return newlyUnlocked, nil
}

// unlockAchievement inserts an achievement record and returns true if it was newly created.
func (db *DBWrapper) unlockAchievement(ctx context.Context, userID string, challengeID int, code string) (bool, error) {
	tag, err := db.Pool.Exec(ctx, `
		INSERT INTO user_achievements (user_id, challenge_id, achievement_code)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id, challenge_id, achievement_code) DO NOTHING
	`, userID, challengeID, code)
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() == 1, nil
}

// GetChallengeAchievements returns all achievements unlocked by a user for a specific challenge, sorted by unlock date.
func (db *DBWrapper) GetChallengeAchievements(ctx context.Context, userID string, challengeID int) ([]models.Achievement, error) {
	rows, err := db.Pool.Query(ctx, `
		SELECT id, user_id, challenge_id, achievement_code, unlocked_at
		FROM user_achievements
		WHERE user_id = $1 AND challenge_id = $2
		ORDER BY unlocked_at ASC
	`, userID, challengeID)
	if err != nil {
		return nil, fmt.Errorf("failed to query achievements: %w", err)
	}
	defer rows.Close()

	achievements := []models.Achievement{} // must be empty slice, not nil
	for rows.Next() {
		var a models.Achievement
		if err := rows.Scan(&a.ID, &a.UserID, &a.ChallengeID, &a.AchievementCode, &a.UnlockedAt); err != nil {
			return nil, fmt.Errorf("failed to scan achievement: %w", err)
		}
		achievements = append(achievements, a)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return achievements, nil
}
