package handlers

import (
	"net/http"

	"workout-challenge-app/internal/database"
)

// SetupRoutes registers all the API endpoints to the provided multiplexer
func SetupRoutes(mux *http.ServeMux, db *database.DBWrapper) {
	exerciseHandler := NewExerciseHandler(db)

	// Since we are using standard library net/http without a 3rd party router (like chi or gorilla),
	// we will map paths and then handle methods inside the handler.
	// For Go 1.22+, we could use method-based routing like `POST /api/exercises`,
	// but to be safe and compatible with older Go versions, we just map the path.
	mux.HandleFunc("/api/exercises", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			exerciseHandler.HandleList(w, r)
		case http.MethodPost:
			exerciseHandler.HandleCreate(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	challengeHandler := NewChallengeHandler(db)

	mux.HandleFunc("/api/challenges", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			challengeHandler.HandleList(w, r)
		case http.MethodPost:
			challengeHandler.HandleCreate(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/challenges/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			challengeHandler.HandleGetByID(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
}
