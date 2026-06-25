package main

import (
	"log"
	"net/http"

	"workout-challenge-app/internal/config"
	"workout-challenge-app/internal/database"
	"workout-challenge-app/internal/handlers"
)

func main() {
	log.Println("Starting Workout Challenge Tracker Backend...")

	// 1. Load configuration
	cfg := config.LoadConfig()
	log.Printf("Server will start on port %s", cfg.ServerPort)

	// 2. Connect to the database with retries
	db, err := database.Connect(cfg.DatabaseDSN)
	if err != nil {
		log.Fatalf("Fatal: Failed to connect to the database after retries: %v", err)
	}
	defer db.Close()

	// 3. Run database migrations
	err = db.RunMigrations()
	if err != nil {
		log.Fatalf("Fatal: Database migrations failed: %v", err)
	}

	// 4. Start the HTTP server and setup routes
	mux := http.NewServeMux()
	
	// Add a simple healthcheck endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Setup API routes
	handlers.SetupRoutes(mux, db)

	// Serve frontend static files
	fs := http.FileServer(http.Dir("./frontend"))
	mux.Handle("/", fs)

	serverAddr := ":" + cfg.ServerPort
	log.Printf("Server listening on %s", serverAddr)
	if err := http.ListenAndServe(serverAddr, mux); err != nil {
		log.Fatalf("Fatal: HTTP server error: %v", err)
	}
}
