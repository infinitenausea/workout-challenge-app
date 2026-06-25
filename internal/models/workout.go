package models

import "time"

// Workout represents a workout session for a challenge
type Workout struct {
	ID          int       `json:"id"`
	UserID      string    `json:"user_id"`
	ChallengeID int       `json:"challenge_id"`
	WorkoutDate time.Time `json:"workout_date"`
	Value       int       `json:"value"`
	CreatedAt   time.Time `json:"created_at"`
}

// WorkoutResponse represents the response when adding a workout
type WorkoutResponse struct {
	Success              bool     `json:"success"`
	Workout              Workout  `json:"workout"`
	UnlockedAchievements []string `json:"unlocked_achievements"`
}
