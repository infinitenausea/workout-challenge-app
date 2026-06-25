package models

import "time"

// Achievement represents a user's unlocked achievement
type Achievement struct {
	ID              int       `json:"id"`
	UserID          string    `json:"user_id"`
	AchievementCode string    `json:"achievement_code"`
	UnlockedAt      time.Time `json:"unlocked_at"`
}

// AchievementDefinition represents the static metadata for an achievement
type AchievementDefinition struct {
	Code        string `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
}
