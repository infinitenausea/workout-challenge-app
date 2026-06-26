package handlers

import (
	"log"
	"net/http"
	"strings"

	"workout-challenge-app/internal/auth"
	"workout-challenge-app/internal/config"
)

// TelegramAuthMiddleware intercepts requests and validates the Telegram initData if in production.
func TelegramAuthMiddleware(cfg *config.Config, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if cfg.AppEnv == "development" {
			userID := r.Header.Get("X-User-Id")
			if userID == "" {
				userID = "default_user_1"
			}
			r.Header.Set("X-User-Id", userID)
			next.ServeHTTP(w, r)
			return
		}

		authHeader := r.Header.Get("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			log.Printf("[WARN] Unauthorized access attempt from IP: %s (missing or invalid Authorization header)", r.RemoteAddr)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		initData := strings.TrimPrefix(authHeader, "Bearer ")

		isValid, err := auth.ValidateTelegramData(initData, cfg.TelegramBotToken)
		if err != nil || !isValid {
			log.Printf("[WARN] Invalid telegram signature from IP: %s", r.RemoteAddr)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		userID, err := auth.GetUserIDFromInitData(initData)
		if err != nil || userID == "" {
			log.Printf("[WARN] Failed to parse user id from initData from IP: %s", r.RemoteAddr)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Inject the validated user ID so that handlers can extract it easily.
		r.Header.Set("X-User-Id", userID)
		next.ServeHTTP(w, r)
	}
}
