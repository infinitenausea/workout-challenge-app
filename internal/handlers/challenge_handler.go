package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

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

// ChallengeUpdatePayload represents the fields that can be updated
type ChallengeUpdatePayload struct {
	Name        *string    `json:"name,omitempty"`
	TargetValue *int       `json:"target_value,omitempty"`
	StartDate   *time.Time `json:"start_date,omitempty"`
	EndDate     *time.Time `json:"end_date,omitempty"`
}

// HandleUpdate handles the PATCH /api/challenges/:id endpoint
func (h *ChallengeHandler) HandleUpdate(w http.ResponseWriter, r *http.Request) {
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

	var payload ChallengeUpdatePayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		log.Printf("Invalid request body for updating challenge: %v\n", err)
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	// Fetch current challenge
	challenge, err := h.db.GetChallengeByID(r.Context(), userID, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "Challenge not found", http.StatusNotFound)
			return
		}
		log.Printf("Database error fetching challenge for update: %v\n", err)
		http.Error(w, "Failed to retrieve challenge", http.StatusInternalServerError)
		return
	}

	// Validate status active
	if challenge.Status != "active" {
		http.Error(w, "Cannot update a non-active challenge", http.StatusBadRequest)
		return
	}

	// Validate target_value >= 1
	if payload.TargetValue != nil && *payload.TargetValue < 1 {
		http.Error(w, "Target value must be at least 1", http.StatusBadRequest)
		return
	}

	now := time.Now().Truncate(24 * time.Hour)
	currentStart := challenge.StartDate.Truncate(24 * time.Hour)

	// Validate start_date
	if payload.StartDate != nil {
		newStart := payload.StartDate.Truncate(24 * time.Hour)
		if !newStart.Equal(currentStart) {
			if currentStart.Before(now) || currentStart.Equal(now) {
				http.Error(w, "Cannot change start_date after the challenge has started", http.StatusBadRequest)
				return
			}
		}
	}

	// Validate end_date
	if payload.EndDate != nil {
		newEnd := payload.EndDate.Truncate(24 * time.Hour)
		if newEnd.Before(now) {
			http.Error(w, "End date cannot be in the past", http.StatusBadRequest)
			return
		}
		
		effectiveStart := currentStart
		if payload.StartDate != nil {
			effectiveStart = payload.StartDate.Truncate(24 * time.Hour)
		}
		if effectiveStart.After(newEnd) {
			http.Error(w, "End date cannot be before start date", http.StatusBadRequest)
			return
		}
	}

	// Determine new status if target_value is updated
	var status *string
	if payload.TargetValue != nil && *payload.TargetValue <= challenge.CurrentProgress {
		completed := "completed"
		status = &completed
	}

	// Update in database
	err = h.db.UpdateChallenge(
		r.Context(),
		userID,
		id,
		payload.Name,
		payload.TargetValue,
		payload.StartDate,
		payload.EndDate,
		status,
	)

	if err != nil {
		log.Printf("Database error updating challenge: %v\n", err)
		http.Error(w, "Failed to update challenge", http.StatusInternalServerError)
		return
	}

	// Fetch updated challenge to return
	updatedChallenge, err := h.db.GetChallengeByID(r.Context(), userID, id)
	if err != nil {
		log.Printf("Error fetching updated challenge: %v\n", err)
		http.Error(w, "Failed to retrieve updated challenge", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(updatedChallenge); err != nil {
		log.Printf("Error encoding response for updated challenge: %v\n", err)
	}
}
