package models

import "time"

// Challenge represents a user's workout challenge
type Challenge struct {
	ID              int       `json:"id"`
	UserID          string    `json:"user_id"`
	Name            string    `json:"name"`
	ExerciseID      int       `json:"exercise_id"`
	TargetValue     int       `json:"target_value"`
	CurrentProgress int       `json:"current_progress"`
	StartDate       time.Time `json:"start_date"`
	EndDate         time.Time `json:"end_date"`
	Status          string    `json:"status"`
}
