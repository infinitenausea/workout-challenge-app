package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

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

// HandleList handles GET /api/challenges/:id/achievements
func (h *AchievementHandler) HandleList(w http.ResponseWriter, r *http.Request) {
	userID := h.getUserID(r)

	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 4 || pathParts[len(pathParts)-1] != "achievements" {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	challengeID, err := strconv.Atoi(pathParts[2])
	if err != nil {
		http.Error(w, "Invalid challenge ID", http.StatusBadRequest)
		return
	}

	achievements, err := h.db.GetChallengeAchievements(r.Context(), userID, challengeID)
	if err != nil {
		log.Printf("HandleList achievements: DB error: %v\n", err)
		http.Error(w, "Failed to retrieve achievements", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
	if err := json.NewEncoder(w).Encode(achievements); err != nil {
		log.Printf("HandleList achievements: failed to encode response: %v\n", err)
	}
}
