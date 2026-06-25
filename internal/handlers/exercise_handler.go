package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"

	"workout-challenge-app/internal/database"
)

// ExerciseHandler handles HTTP requests for exercises
type ExerciseHandler struct {
	DB *database.DBWrapper
}

// NewExerciseHandler creates a new ExerciseHandler
func NewExerciseHandler(db *database.DBWrapper) *ExerciseHandler {
	return &ExerciseHandler{DB: db}
}

// CreateExerciseRequest represents the JSON body for creating an exercise
type CreateExerciseRequest struct {
	Name string `json:"name"`
}

// HandleCreate handles POST /api/exercises
func (h *ExerciseHandler) HandleCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := r.Header.Get("X-User-Id")
	if userID == "" {
		userID = "default_user_1"
	}

	var req CreateExerciseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error decoding request body: %v", err)
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		log.Printf("Validation failed: empty name provided by user %s", userID)
		http.Error(w, "Exercise name cannot be empty", http.StatusBadRequest)
		return
	}

	exercise, err := h.DB.CreateExercise(r.Context(), userID, name)
	if err != nil {
		if errors.Is(err, database.ErrDuplicateExercise) {
			log.Printf("Conflict: exercise '%s' already exists for user %s", name, userID)
			http.Error(w, "Exercise with this name already exists", http.StatusConflict)
			return
		}
		log.Printf("Server error while creating exercise: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(exercise); err != nil {
		log.Printf("Error encoding response: %v", err)
	}
}

// HandleList handles GET /api/exercises
func (h *ExerciseHandler) HandleList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := r.Header.Get("X-User-Id")
	if userID == "" {
		userID = "default_user_1"
	}

	exercises, err := h.DB.GetExercises(r.Context(), userID)
	if err != nil {
		log.Printf("Server error while fetching exercises for user %s: %v", userID, err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(exercises); err != nil {
		log.Printf("Error encoding response: %v", err)
	}
}

