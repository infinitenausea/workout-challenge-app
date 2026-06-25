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

	// 1. first_step: Внесена первая тренировка.
	// SQL: SELECT EXISTS(SELECT 1 FROM workouts WHERE user_id = $1 LIMIT 1)
	firstStepEligible := false
	err := db.Pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM workouts WHERE user_id = $1 LIMIT 1)", userID).Scan(&firstStepEligible)
	if err != nil {
		log.Printf("CheckAndUnlockAchievements: failed to check first_step eligibility: %v\n", err)
	} else if firstStepEligible {
		unlocked, err := db.unlockAchievement(ctx, userID, "first_step")
		if err != nil {
			log.Printf("CheckAndUnlockAchievements: failed to unlock first_step: %v\n", err)
		} else if unlocked {
			newlyUnlocked = append(newlyUnlocked, "first_step")
		}
	}

	// 2. equator: newProgress * 2 >= targetValue
	if newProgress*2 >= targetValue {
		unlocked, err := db.unlockAchievement(ctx, userID, "equator")
		if err != nil {
			log.Printf("CheckAndUnlockAchievements: failed to unlock equator: %v\n", err)
		} else if unlocked {
			newlyUnlocked = append(newlyUnlocked, "equator")
		}
	}

	// 3. hero: newProgress >= targetValue && end_date >= CURRENT_DATE
	if newProgress >= targetValue {
		var endDate time.Time
		err := db.Pool.QueryRow(ctx, "SELECT end_date FROM challenges WHERE id = $1", challengeID).Scan(&endDate)
		if err != nil {
			log.Printf("CheckAndUnlockAchievements: failed to get challenge end_date for hero check: %v\n", err)
		} else {
			// Compare endDate (only date part) with current time in local/UTC.
			// The DB DATE is parsed as time.Time. We check if it is before today.
			// A clean way is comparing year, month, day.
			now := time.Now()
			today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
			
			// If endDate is today or in the future
			if !endDate.Before(today) {
				unlocked, err := db.unlockAchievement(ctx, userID, "hero")
				if err != nil {
					log.Printf("CheckAndUnlockAchievements: failed to unlock hero: %v\n", err)
				} else if unlocked {
					newlyUnlocked = append(newlyUnlocked, "hero")
				}
			}
		}
	}

	// 4. stability: тренировки вносились 3 дня подряд.
	// SQL: SELECT DISTINCT workout_date FROM workouts WHERE user_id = $1 ORDER BY workout_date DESC LIMIT 10
	rows, err := db.Pool.Query(ctx, "SELECT DISTINCT workout_date FROM workouts WHERE user_id = $1 ORDER BY workout_date DESC LIMIT 10", userID)
	if err != nil {
		log.Printf("CheckAndUnlockAchievements: failed to query workout dates for stability check: %v\n", err)
	} else {
		defer rows.Close()
		var dates []time.Time
		for rows.Next() {
			var d time.Time
			if err := rows.Scan(&d); err == nil {
				dates = append(dates, d)
			}
		}
		
		// Check for 3 consecutive days in the retrieved dates
		// dates are sorted DESC, e.g. [D_today, D_yesterday, D_day_before, ...]
		stabilityEligible := false
		for i := 0; i <= len(dates)-3; i++ {
			d1 := dates[i]
			d2 := dates[i+1]
			d3 := dates[i+2]
			
			// Normalize to UTC to avoid DST/timezone issues when comparing differences
			utc1 := time.Date(d1.Year(), d1.Month(), d1.Day(), 0, 0, 0, 0, time.UTC)
			utc2 := time.Date(d2.Year(), d2.Month(), d2.Day(), 0, 0, 0, 0, time.UTC)
			utc3 := time.Date(d3.Year(), d3.Month(), d3.Day(), 0, 0, 0, 0, time.UTC)
			
			if utc1.Sub(utc2) == 24*time.Hour && utc2.Sub(utc3) == 24*time.Hour {
				stabilityEligible = true
				break
			}
		}

		if stabilityEligible {
			unlocked, err := db.unlockAchievement(ctx, userID, "stability")
			if err != nil {
				log.Printf("CheckAndUnlockAchievements: failed to unlock stability: %v\n", err)
			} else if unlocked {
				newlyUnlocked = append(newlyUnlocked, "stability")
			}
		}
	}

	return newlyUnlocked, nil
}

// unlockAchievement inserts an achievement record and returns true if it was newly created.
func (db *DBWrapper) unlockAchievement(ctx context.Context, userID string, code string) (bool, error) {
	tag, err := db.Pool.Exec(ctx, `
		INSERT INTO user_achievements (user_id, achievement_code)
		VALUES ($1, $2)
		ON CONFLICT (user_id, achievement_code) DO NOTHING
	`, userID, code)
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() == 1, nil
}

// GetUserAchievements returns all achievements unlocked by a user, sorted by unlock date.
func (db *DBWrapper) GetUserAchievements(ctx context.Context, userID string) ([]models.Achievement, error) {
	rows, err := db.Pool.Query(ctx, `
		SELECT id, user_id, achievement_code, unlocked_at
		FROM user_achievements
		WHERE user_id = $1
		ORDER BY unlocked_at ASC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query achievements: %w", err)
	}
	defer rows.Close()

	achievements := []models.Achievement{} // must be empty slice, not nil
	for rows.Next() {
		var a models.Achievement
		if err := rows.Scan(&a.ID, &a.UserID, &a.AchievementCode, &a.UnlockedAt); err != nil {
			return nil, fmt.Errorf("failed to scan achievement: %w", err)
		}
		achievements = append(achievements, a)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return achievements, nil
}
