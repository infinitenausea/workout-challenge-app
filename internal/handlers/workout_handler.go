package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"workout-challenge-app/internal/database"
	"workout-challenge-app/internal/models"
)

// WorkoutHandler handles HTTP requests related to workouts
type WorkoutHandler struct {
	db *database.DBWrapper
}

// NewWorkoutHandler creates a new WorkoutHandler
func NewWorkoutHandler(db *database.DBWrapper) *WorkoutHandler {
	return &WorkoutHandler{db: db}
}

func (h *WorkoutHandler) getUserID(r *http.Request) string {
	userID := r.Header.Get("X-User-Id")
	if userID == "" {
		userID = "default_user_1"
	}
	return userID
}

// createWorkoutRequest represents the JSON request payload for adding a workout
type createWorkoutRequest struct {
	WorkoutDate string `json:"workout_date"`
	Value       int    `json:"value"`
}

// HandleCreateWorkout handles POST /api/challenges/:id/workouts
func (h *WorkoutHandler) HandleCreateWorkout(w http.ResponseWriter, r *http.Request) {
	userID := h.getUserID(r)

	// Extract challenge ID from URL path (e.g. /api/challenges/123/workouts)
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 3 {
		log.Printf("HandleCreateWorkout: invalid path structure: %s\n", r.URL.Path)
		http.Error(w, "Invalid URL path", http.StatusBadRequest)
		return
	}

	challengeID, err := strconv.Atoi(parts[2])
	if err != nil {
		log.Printf("HandleCreateWorkout: invalid challenge ID: %v\n", err)
		http.Error(w, "Invalid challenge ID", http.StatusBadRequest)
		return
	}

	var req createWorkoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("HandleCreateWorkout: failed to decode JSON: %v\n", err)
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	// Validation
	if req.Value <= 0 {
		log.Printf("HandleCreateWorkout: invalid value: %d\n", req.Value)
		http.Error(w, "Workout value must be greater than 0", http.StatusBadRequest)
		return
	}

	parsedDate, err := time.Parse("2006-01-02", req.WorkoutDate)
	if err != nil {
		log.Printf("HandleCreateWorkout: invalid date format: %v\n", err)
		http.Error(w, "Invalid date format. Expected YYYY-MM-DD", http.StatusBadRequest)
		return
	}

	workout := &models.Workout{
		WorkoutDate: parsedDate,
		Value:       req.Value,
	}

	// Call DB to create workout inside a transaction
	newWorkout, newProgress, targetValue, err := h.db.CreateWorkout(r.Context(), userID, challengeID, workout)
	if err != nil {
		log.Printf("HandleCreateWorkout: DB error: %v\n", err)
		if strings.Contains(err.Error(), "challenge not found or not active") {
			http.Error(w, "Challenge not found or not active", http.StatusBadRequest)
		} else {
			http.Error(w, "Failed to create workout", http.StatusInternalServerError)
		}
		return
	}

	// Check and unlock achievements
	unlocked, err := h.db.CheckAndUnlockAchievements(r.Context(), userID, challengeID, newProgress, targetValue)
	if err != nil {
		log.Printf("HandleCreateWorkout: failed to check achievements: %v\n", err)
		unlocked = []string{} // ensure it's not nil
	}
	if unlocked == nil {
		unlocked = []string{}
	}

	resp := models.WorkoutResponse{
		Success:              true,
		Workout:              *newWorkout,
		UnlockedAchievements: unlocked,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("HandleCreateWorkout: failed to encode response: %v\n", err)
	}
}

// deleteWorkoutResponse represents the response structure for workout deletion
type deleteWorkoutResponse struct {
	Success   bool              `json:"success"`
	Challenge *models.Challenge `json:"challenge"`
}

// HandleDeleteWorkout handles DELETE /api/workouts/:id
func (h *WorkoutHandler) HandleDeleteWorkout(w http.ResponseWriter, r *http.Request) {
	userID := h.getUserID(r)

	// Extract workout ID from URL path (e.g. /api/workouts/123)
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 3 {
		log.Printf("HandleDeleteWorkout: invalid path structure: %s\n", r.URL.Path)
		http.Error(w, "Invalid URL path", http.StatusBadRequest)
		return
	}

	workoutID, err := strconv.Atoi(parts[2])
	if err != nil {
		log.Printf("HandleDeleteWorkout: invalid workout ID: %v\n", err)
		http.Error(w, "Invalid workout ID", http.StatusBadRequest)
		return
	}

	challenge, err := h.db.DeleteWorkout(r.Context(), userID, workoutID)
	if err != nil {
		log.Printf("HandleDeleteWorkout: DB error: %v\n", err)
		// Return 404 if workout is not found/not authorized, otherwise 500
		if strings.Contains(err.Error(), "no rows") || strings.Contains(err.Error(), "not found") {
			http.Error(w, "Workout not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to delete workout", http.StatusInternalServerError)
		}
		return
	}

	resp := deleteWorkoutResponse{
		Success:   true,
		Challenge: challenge,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("HandleDeleteWorkout: failed to encode response: %v\n", err)
	}
}
