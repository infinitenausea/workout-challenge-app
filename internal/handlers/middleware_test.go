package handlers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"workout-challenge-app/internal/config"
)

func TestTelegramAuthMiddleware(t *testing.T) {
	// TC-9.3: Успешная отработка Dev Bypass при APP_ENV="development"
	t.Run("Dev Bypass: APP_ENV=development", func(t *testing.T) {
		cfg := &config.Config{
			AppEnv: "development",
		}
		
		handlerCalled := false
		var parsedUserID string
		
		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handlerCalled = true
			parsedUserID = r.Header.Get("X-User-Id")
			w.WriteHeader(http.StatusOK)
		})
		
		middleware := TelegramAuthMiddleware(cfg, testHandler)
		
		// 1. With X-User-Id header
		req := httptest.NewRequest(http.MethodGet, "/api/exercises", nil)
		req.Header.Set("X-User-Id", "custom_user")
		rec := httptest.NewRecorder()
		
		middleware.ServeHTTP(rec, req)
		
		if rec.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", rec.Code)
		}
		if !handlerCalled {
			t.Error("Handler was not called")
		}
		if parsedUserID != "custom_user" {
			t.Errorf("Expected user ID custom_user, got %s", parsedUserID)
		}
		
		// 2. Without X-User-Id header (fallback to default_user_1)
		handlerCalled = false
		req = httptest.NewRequest(http.MethodGet, "/api/exercises", nil)
		rec = httptest.NewRecorder()
		
		middleware.ServeHTTP(rec, req)
		
		if rec.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", rec.Code)
		}
		if !handlerCalled {
			t.Error("Handler was not called")
		}
		if parsedUserID != "default_user_1" {
			t.Errorf("Expected user ID default_user_1, got %s", parsedUserID)
		}
	})

	// Setup for prod tests
	token := "123456789:ABCdefGhIJKlmNoPQRsTUVwxyZ"
	userJSON := `{"id":98765,"first_name":"Test"}`
	
	// Create the data_check_string using sorted order
	nowStr := fmt.Sprintf("%d", time.Now().Unix())
	dataCheckString := fmt.Sprintf("auth_date=%s\nuser=%s", nowStr, userJSON)
	
	// Compute valid hash
	secretMac := hmac.New(sha256.New, []byte("WebAppData"))
	secretMac.Write([]byte(token))
	secretKey := secretMac.Sum(nil)
	
	mac := hmac.New(sha256.New, secretKey)
	mac.Write([]byte(dataCheckString))
	validHash := hex.EncodeToString(mac.Sum(nil))
	
	// Format as URL query parameter string
	validInitData := fmt.Sprintf("auth_date=%s&user=%s&hash=%s", nowStr, userJSON, validHash)

	// TC-9.1: Успешная валидация initData по тестовому токену
	t.Run("Prod: Valid signature", func(t *testing.T) {
		cfg := &config.Config{
			AppEnv:           "production",
			TelegramBotToken: token,
		}
		
		handlerCalled := false
		var parsedUserID string
		
		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handlerCalled = true
			parsedUserID = r.Header.Get("X-User-Id")
			w.WriteHeader(http.StatusOK)
		})
		
		middleware := TelegramAuthMiddleware(cfg, testHandler)
		
		req := httptest.NewRequest(http.MethodGet, "/api/exercises", nil)
		req.Header.Set("Authorization", "Bearer "+validInitData)
		rec := httptest.NewRecorder()
		
		middleware.ServeHTTP(rec, req)
		
		if rec.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d. Body: %s", rec.Code, rec.Body.String())
		}
		if !handlerCalled {
			t.Error("Handler was not called")
		}
		if parsedUserID != "98765" {
			t.Errorf("Expected user ID 98765, got %s", parsedUserID)
		}
	})

	// TC-9.2: Проверка на HTTP 401 при неверном хэше
	t.Run("Prod: Invalid signature", func(t *testing.T) {
		cfg := &config.Config{
			AppEnv:           "production",
			TelegramBotToken: token,
		}
		
		handlerCalled := false
		
		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handlerCalled = true
			w.WriteHeader(http.StatusOK)
		})
		
		middleware := TelegramAuthMiddleware(cfg, testHandler)
		
		// Corrupt the hash
		invalidInitData := fmt.Sprintf("auth_date=%s&user=%s&hash=%s", nowStr, userJSON, "invalid_hash_value")
		
		req := httptest.NewRequest(http.MethodGet, "/api/exercises", nil)
		req.Header.Set("Authorization", "Bearer "+invalidInitData)
		rec := httptest.NewRecorder()
		
		middleware.ServeHTTP(rec, req)
		
		if rec.Code != http.StatusUnauthorized {
			t.Errorf("Expected status 401, got %d", rec.Code)
		}
		if handlerCalled {
			t.Error("Handler should not have been called")
		}
	})

	// TC: Missing or invalid Authorization header
	t.Run("Prod: Missing header", func(t *testing.T) {
		cfg := &config.Config{
			AppEnv:           "production",
			TelegramBotToken: token,
		}
		
		handlerCalled := false
		
		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handlerCalled = true
			w.WriteHeader(http.StatusOK)
		})
		
		middleware := TelegramAuthMiddleware(cfg, testHandler)
		
		req := httptest.NewRequest(http.MethodGet, "/api/exercises", nil)
		rec := httptest.NewRecorder()
		
		middleware.ServeHTTP(rec, req)
		
		if rec.Code != http.StatusUnauthorized {
			t.Errorf("Expected status 401, got %d", rec.Code)
		}
		if handlerCalled {
			t.Error("Handler should not have been called")
		}
	})
}
