package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"workout-challenge-app/internal/database"
)

// AchievementHandler handles HTTP requests related to achievements
type AchievementHandler struct {
	db *database.DBWrapper
}

// NewAchievementHandler creates a new AchievementHandler
func NewAchievementHandler(db *database.DBWrapper) *AchievementHandler {
	return &AchievementHandler{db: db}
}

func (h *AchievementHandler) getUserID(r *http.Request) string {
	userID := r.Header.Get("X-User-Id")
	if userID == "" {
		userID = "default_user_1"
	}
	return userID
}

// HandleList handles GET /api/achievements
func (h *AchievementHandler) HandleList(w http.ResponseWriter, r *http.Request) {
	userID := h.getUserID(r)

	achievements, err := h.db.GetUserAchievements(r.Context(), userID)
	if err != nil {
		log.Printf("HandleList achievements: DB error: %v\n", err)
		http.Error(w, "Failed to retrieve achievements", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(achievements); err != nil {
		log.Printf("HandleList achievements: failed to encode response: %v\n", err)
	}
}
