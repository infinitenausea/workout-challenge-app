package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
	"workout-challenge-app/internal/database"
	"workout-challenge-app/internal/models"
)

// ChallengeHandler handles HTTP requests related to challenges
type ChallengeHandler struct {
	db *database.DBWrapper
}

// NewChallengeHandler creates a new ChallengeHandler
func NewChallengeHandler(db *database.DBWrapper) *ChallengeHandler {
	return &ChallengeHandler{db: db}
}

// getUserID extracts the user ID from the headers, falling back to a default
func (h *ChallengeHandler) getUserID(r *http.Request) string {
	userID := r.Header.Get("X-User-Id")
	if userID == "" {
		userID = "default_user_1"
	}
	return userID
}

// HandleCreate handles the POST /api/challenges endpoint
func (h *ChallengeHandler) HandleCreate(w http.ResponseWriter, r *http.Request) {
	userID := h.getUserID(r)

	var challenge models.Challenge
	if err := json.NewDecoder(r.Body).Decode(&challenge); err != nil {
		log.Printf("Invalid request body for creating challenge: %v\n", err)
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	// Validation
	if strings.TrimSpace(challenge.Name) == "" {
		log.Printf("Create challenge validation failed: empty name")
		http.Error(w, "Challenge name cannot be empty", http.StatusBadRequest)
		return
	}
	if challenge.TargetValue <= 0 {
		log.Printf("Create challenge validation failed: target_value <= 0 (%d)", challenge.TargetValue)
		http.Error(w, "Target value must be greater than 0", http.StatusBadRequest)
		return
	}
	
	// Truncate times to dates for comparison, as DB uses DATE type
	startDate := challenge.StartDate.Truncate(24 * 60 * 60 * 1000 * 1000 * 1000) // Truncate to days
	endDate := challenge.EndDate.Truncate(24 * 60 * 60 * 1000 * 1000 * 1000)
	
	// Simplest comparison: start date shouldn't be after end date. 
	// To be safe with any timezone/truncation issues, we compare simply using After.
	if startDate.After(endDate) {
		log.Printf("Create challenge validation failed: end_date before start_date")
		http.Error(w, "End date cannot be before start date", http.StatusBadRequest)
		return
	}

	err := h.db.CreateChallenge(r.Context(), userID, &challenge)
	if err != nil {
		log.Printf("Database error creating challenge: %v\n", err)
		http.Error(w, "Failed to create challenge", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(challenge); err != nil {
		log.Printf("Error encoding response for created challenge: %v\n", err)
	}
}

// HandleList handles the GET /api/challenges endpoint
func (h *ChallengeHandler) HandleList(w http.ResponseWriter, r *http.Request) {
	userID := h.getUserID(r)

	challenges, err := h.db.GetChallenges(r.Context(), userID)
	if err != nil {
		log.Printf("Database error querying challenges: %v\n", err)
		http.Error(w, "Failed to retrieve challenges", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(challenges); err != nil {
		log.Printf("Error encoding response for challenges list: %v\n", err)
	}
}

// HandleGetByID handles the GET /api/challenges/:id endpoint
func (h *ChallengeHandler) HandleGetByID(w http.ResponseWriter, r *http.Request) {
	userID := h.getUserID(r)

	// Extract ID from URL path
	// Assuming path is like /api/challenges/{id}
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 3 {
		http.Error(w, "Invalid challenge ID", http.StatusBadRequest)
		return
	}

	idStr := pathParts[2]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Printf("Invalid challenge ID format: %s\n", idStr)
		http.Error(w, "Invalid challenge ID format", http.StatusBadRequest)
		return
	}

	challenge, err := h.db.GetChallengeByID(r.Context(), userID, id)
	if err != nil {
		log.Printf("Challenge not found or db error: %v\n", err)
		http.Error(w, "Challenge not found", http.StatusNotFound)
		return
	}

	workouts, err := h.db.GetWorkoutsByChallenge(r.Context(), userID, id)
	if err != nil {
		log.Printf("Failed to get workouts for challenge detail: %v\n", err)
		workouts = []models.Workout{} // fallback to empty slice
	}

	resp := struct {
		*models.Challenge
		Workouts []models.Workout `json:"workouts"`
	}{
		Challenge: challenge,
		Workouts:  workouts,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("Error encoding response for challenge details: %v\n", err)
	}
}

// HandleDelete handles the DELETE /api/challenges/:id endpoint
func (h *ChallengeHandler) HandleDelete(w http.ResponseWriter, r *http.Request) {
	userID := h.getUserID(r)

	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 3 {
		http.Error(w, "Invalid challenge ID", http.StatusBadRequest)
		return
	}

	idStr := pathParts[2]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Printf("Invalid challenge ID format: %s\n", idStr)
		http.Error(w, "Invalid challenge ID format", http.StatusBadRequest)
		return
	}

	err = h.db.DeleteChallenge(r.Context(), userID, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			log.Printf("Challenge not found for deletion: id %d\n", id)
			http.Error(w, "Challenge not found", http.StatusNotFound)
			return
		}
		log.Printf("Database error deleting challenge: %v\n", err)
		http.Error(w, "Failed to delete challenge", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}
