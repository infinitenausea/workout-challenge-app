package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"

	"github.com/jackc/pgx/v5"
)

const (
	baseURL = "http://localhost:8080"
	defaultDSN = "postgres://postgres:postgres_secure_pass@localhost:5433/workout_tracker?sslmode=disable"
)

func getDBConn(t *testing.T) *pgx.Conn {
	dsn := os.Getenv("DATABASE_DSN")
	if dsn == "" {
		dsn = defaultDSN
	}
	conn, err := pgx.Connect(context.Background(), dsn)
	if err != nil {
		t.Fatalf("Failed to connect to DB: %v", err)
	}
	return conn
}

func resetDatabase(t *testing.T, conn *pgx.Conn) {
	ctx := context.Background()
	_, err := conn.Exec(ctx, "TRUNCATE TABLE workouts, challenges, user_achievements, exercises RESTART IDENTITY CASCADE;")
	if err != nil {
		t.Fatalf("Failed to truncate tables: %v", err)
	}

	// Insert default exercises (let database assign IDs automatically)
	_, err = conn.Exec(ctx, "INSERT INTO exercises (user_id, name, is_custom) VALUES ('default_user_1', 'Pushups', false);")
	if err != nil {
		t.Fatalf("Failed to insert default exercise: %v", err)
	}
	_, err = conn.Exec(ctx, "INSERT INTO exercises (user_id, name, is_custom) VALUES ('default_user_1', 'Squats', false);")
	if err != nil {
		t.Fatalf("Failed to insert default exercise 2: %v", err)
	}

	// Insert challenge 1 for default_user_1
	_, err = conn.Exec(ctx, `
		INSERT INTO challenges (user_id, name, exercise_id, target_value, current_progress, start_date, end_date, status)
		VALUES ('default_user_1', 'Challenge 1', 1, 100, 0, '2026-06-01', '2026-07-01', 'active');
	`)
	if err != nil {
		t.Fatalf("Failed to insert challenge 1: %v", err)
	}
}

func sendRequest(t *testing.T, method, path string, userID string, body interface{}) (*http.Response, []byte) {
	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("Failed to marshal request body: %v", err)
		}
		bodyReader = bytes.NewReader(b)
	}

	req, err := http.NewRequest(method, baseURL+path, bodyReader)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if userID != "" {
		req.Header.Set("X-User-Id", userID)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	return resp, respBody
}

func TestAPI_Sprint3(t *testing.T) {
	conn := getDBConn(t)
	defer conn.Close(context.Background())

	t.Run("Reset DB", func(t *testing.T) {
		resetDatabase(t, conn)
	})

	// TC-3.1 (Positive): Успешное добавление, проверка 201, unlocked_achievements: ["first_step"], обновление current_progress
	var workout1ID int
	t.Run("TC-3.1 Positive: Add workout", func(t *testing.T) {
		payload := map[string]interface{}{
			"workout_date": "2026-06-25",
			"value":        30,
		}
		resp, body := sendRequest(t, "POST", "/api/challenges/1/workouts", "default_user_1", payload)
		if resp.StatusCode != http.StatusCreated {
			t.Errorf("Expected status 201, got %d. Body: %s", resp.StatusCode, string(body))
		}

		var respData map[string]interface{}
		if err := json.Unmarshal(body, &respData); err != nil {
			t.Fatalf("Failed to unmarshal body: %v", err)
		}

		if respData["success"] != true {
			t.Errorf("Expected success to be true, got %v", respData["success"])
		}

		unlocked, ok := respData["unlocked_achievements"].([]interface{})
		foundFirstStep := false
		if ok {
			for _, ach := range unlocked {
				if ach == "first_step" {
					foundFirstStep = true
				}
			}
		}
		if !foundFirstStep {
			t.Errorf("Expected unlocked_achievements to contain 'first_step', got %v", respData["unlocked_achievements"])
		}

		workout, ok := respData["workout"].(map[string]interface{})
		if !ok {
			t.Fatalf("Expected workout in response, got nil")
		}
		workout1ID = int(workout["id"].(float64))
		_ = workout1ID

		// Check DB progress
		var progress int
		err := conn.QueryRow(context.Background(), "SELECT current_progress FROM challenges WHERE id = 1").Scan(&progress)
		if err != nil {
			t.Fatalf("Failed to query challenge: %v", err)
		}
		if progress != 30 {
			t.Errorf("Expected current_progress to be 30, got %d", progress)
		}
	})

	// TC-3.2 (Positive): Ачивка «Экватор» при достижении 50%
	var workout2ID int
	t.Run("TC-3.2 Positive: Equator achievement", func(t *testing.T) {
		payload := map[string]interface{}{
			"workout_date": "2026-06-26",
			"value":        20,
		}
		resp, body := sendRequest(t, "POST", "/api/challenges/1/workouts", "default_user_1", payload)
		if resp.StatusCode != http.StatusCreated {
			t.Errorf("Expected status 201, got %d. Body: %s", resp.StatusCode, string(body))
		}

		var respData map[string]interface{}
		if err := json.Unmarshal(body, &respData); err != nil {
			t.Fatalf("Failed to unmarshal body: %v", err)
		}

		unlocked, ok := respData["unlocked_achievements"].([]interface{})
		if !ok || len(unlocked) != 1 || unlocked[0] != "equator" {
			t.Errorf("Expected unlocked_achievements to contain only 'equator', got %v", respData["unlocked_achievements"])
		}

		workout := respData["workout"].(map[string]interface{})
		workout2ID = int(workout["id"].(float64))
		_ = workout2ID

		// Check DB
		var progress int
		err := conn.QueryRow(context.Background(), "SELECT current_progress FROM challenges WHERE id = 1").Scan(&progress)
		if err != nil {
			t.Fatalf("Failed to query challenge: %v", err)
		}
		if progress != 50 {
			t.Errorf("Expected current_progress to be 50, got %d", progress)
		}
	})

	// TC-3.3 (Positive): Ачивка «Герой» при 100% до дедлайна, status → completed
	var workout3ID int
	t.Run("TC-3.3 Positive: Hero achievement", func(t *testing.T) {
		// Use June 29 to avoid triggering stability (since consecutive would be 25, 26, 27)
		payload := map[string]interface{}{
			"workout_date": "2026-06-29",
			"value":        50,
		}
		resp, body := sendRequest(t, "POST", "/api/challenges/1/workouts", "default_user_1", payload)
		if resp.StatusCode != http.StatusCreated {
			t.Errorf("Expected status 201, got %d. Body: %s", resp.StatusCode, string(body))
		}

		var respData map[string]interface{}
		if err := json.Unmarshal(body, &respData); err != nil {
			t.Fatalf("Failed to unmarshal body: %v", err)
		}

		unlocked, ok := respData["unlocked_achievements"].([]interface{})
		if !ok || len(unlocked) != 1 || unlocked[0] != "hero" {
			t.Errorf("Expected unlocked_achievements to contain only 'hero', got %v", respData["unlocked_achievements"])
		}

		workout := respData["workout"].(map[string]interface{})
		workout3ID = int(workout["id"].(float64))

		// Check DB
		var progress int
		var status string
		err := conn.QueryRow(context.Background(), "SELECT current_progress, status FROM challenges WHERE id = 1").Scan(&progress, &status)
		if err != nil {
			t.Fatalf("Failed to query challenge: %v", err)
		}
		if progress != 100 {
			t.Errorf("Expected current_progress to be 100, got %d", progress)
		}
		if status != "completed" {
			t.Errorf("Expected status to be 'completed', got %s", status)
		}
	})

	// TC-3.4 (Positive): Ачивка «Стабильность» при 3 днях подряд
	t.Run("TC-3.4 Positive: Stability achievement", func(t *testing.T) {
		// Create Challenge 2
		c2Payload := map[string]interface{}{
			"name":        "Challenge 2",
			"exercise_id": 1,
			"target_value": 200,
			"start_date":  "2026-06-01T00:00:00Z",
			"end_date":    "2026-07-01T00:00:00Z",
		}
		resp, body := sendRequest(t, "POST", "/api/challenges", "default_user_1", c2Payload)
		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("Failed to create Challenge 2, status %d. Body: %s", resp.StatusCode, string(body))
		}
		var c2Data map[string]interface{}
		if err := json.Unmarshal(body, &c2Data); err != nil {
			t.Fatalf("Failed to unmarshal Challenge 2: %v", err)
		}
		c2ID := int(c2Data["id"].(float64))

		// Add workout on June 25
		w1Payload := map[string]interface{}{
			"workout_date": "2026-06-25",
			"value":        10,
		}
		resp, body = sendRequest(t, "POST", fmt.Sprintf("/api/challenges/%d/workouts", c2ID), "default_user_1", w1Payload)
		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("Failed to add workout 1: %s", string(body))
		}

		// Add workout on June 26
		w2Payload := map[string]interface{}{
			"workout_date": "2026-06-26",
			"value":        10,
		}
		resp, body = sendRequest(t, "POST", fmt.Sprintf("/api/challenges/%d/workouts", c2ID), "default_user_1", w2Payload)
		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("Failed to add workout 2: %s", string(body))
		}

		// Add workout on June 27 (triggers stability!)
		w3Payload := map[string]interface{}{
			"workout_date": "2026-06-27",
			"value":        10,
		}
		resp, body = sendRequest(t, "POST", fmt.Sprintf("/api/challenges/%d/workouts", c2ID), "default_user_1", w3Payload)
		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("Failed to add workout 3: %s", string(body))
		}

		var respData map[string]interface{}
		json.Unmarshal(body, &respData)
		
		unlocked, ok := respData["unlocked_achievements"].([]interface{})
		found := false
		if ok {
			for _, val := range unlocked {
				if val == "stability" {
					found = true
				}
			}
		}
		if !found {
			t.Errorf("Expected stability to be unlocked, got %v", respData["unlocked_achievements"])
		}

		// Add workout on June 28 (extra to preserve stability when one is deleted later)
		w4Payload := map[string]interface{}{
			"workout_date": "2026-06-28",
			"value":        10,
		}
		resp, body = sendRequest(t, "POST", fmt.Sprintf("/api/challenges/%d/workouts", c2ID), "default_user_1", w4Payload)
		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("Failed to add workout 4: %s", string(body))
		}
	})

	// TC-3.5 (Negative): Отрицательное value → 400
	t.Run("TC-3.5 Negative: Negative value", func(t *testing.T) {
		payload := map[string]interface{}{
			"workout_date": "2026-06-25",
			"value":        -10,
		}
		resp, body := sendRequest(t, "POST", "/api/challenges/2/workouts", "default_user_1", payload)
		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d. Body: %s", resp.StatusCode, string(body))
		}
	})

	// TC-3.6 (Negative): value = 0 → 400
	t.Run("TC-3.6 Negative: Zero value", func(t *testing.T) {
		payload := map[string]interface{}{
			"workout_date": "2026-06-25",
			"value":        0,
		}
		resp, body := sendRequest(t, "POST", "/api/challenges/2/workouts", "default_user_1", payload)
		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d. Body: %s", resp.StatusCode, string(body))
		}
	})

	// TC-3.7 (Negative): Несуществующий челлендж → 400/404
	t.Run("TC-3.7 Negative: Non-existent challenge", func(t *testing.T) {
		payload := map[string]interface{}{
			"workout_date": "2026-06-25",
			"value":        10,
		}
		resp, body := sendRequest(t, "POST", "/api/challenges/99999/workouts", "default_user_1", payload)
		if resp.StatusCode != http.StatusBadRequest && resp.StatusCode != http.StatusNotFound {
			t.Errorf("Expected status 400 or 404, got %d. Body: %s", resp.StatusCode, string(body))
		}
	})

	// TC-3.8 (Negative): Завершённый челлендж → 400
	t.Run("TC-3.8 Negative: Completed challenge", func(t *testing.T) {
		payload := map[string]interface{}{
			"workout_date": "2026-06-25",
			"value":        10,
		}
		resp, body := sendRequest(t, "POST", "/api/challenges/1/workouts", "default_user_1", payload)
		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d. Body: %s", resp.StatusCode, string(body))
		}
	})

	// TC-3.9 (Edge): Пустая дата → 400
	t.Run("TC-3.9 Edge: Empty date", func(t *testing.T) {
		payload := map[string]interface{}{
			"value": 10,
		}
		resp, body := sendRequest(t, "POST", "/api/challenges/2/workouts", "default_user_1", payload)
		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d. Body: %s", resp.StatusCode, string(body))
		}
	})

	// TC-3.10 (Edge): Ачивки не дублируются
	t.Run("TC-3.10 Edge: Achievements don't duplicate", func(t *testing.T) {
		payload := map[string]interface{}{
			"workout_date": "2026-06-30",
			"value":        10,
		}
		resp, body := sendRequest(t, "POST", "/api/challenges/2/workouts", "default_user_1", payload)
		if resp.StatusCode != http.StatusCreated {
			t.Errorf("Expected status 201, got %d. Body: %s", resp.StatusCode, string(body))
		}

		var respData map[string]interface{}
		if err := json.Unmarshal(body, &respData); err == nil {
			if unlocked, ok := respData["unlocked_achievements"].([]interface{}); ok {
				if len(unlocked) != 0 {
					t.Errorf("Expected no new achievements unlocked, got %v", unlocked)
				}
			}
		}
	})

	// TC-3.11 (Positive): Успешное удаление, current_progress уменьшился
	t.Run("TC-3.11 Positive: Delete workout", func(t *testing.T) {
		var progressBefore int
		err := conn.QueryRow(context.Background(), "SELECT current_progress FROM challenges WHERE id = 2").Scan(&progressBefore)
		if err != nil {
			t.Fatalf("Failed to query challenge 2 progress: %v", err)
		}

		var wID int
		var wVal int
		err = conn.QueryRow(context.Background(), "SELECT id, value FROM workouts WHERE challenge_id = 2 LIMIT 1").Scan(&wID, &wVal)
		if err != nil {
			t.Fatalf("Failed to query workout from challenge 2: %v", err)
		}

		resp, body := sendRequest(t, "DELETE", fmt.Sprintf("/api/workouts/%d", wID), "default_user_1", nil)
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d. Body: %s", resp.StatusCode, string(body))
		}

		var respData map[string]interface{}
		json.Unmarshal(body, &respData)

		if respData["success"] != true {
			t.Errorf("Expected success to be true, got %v", respData["success"])
		}

		var progressAfter int
		err = conn.QueryRow(context.Background(), "SELECT current_progress FROM challenges WHERE id = 2").Scan(&progressAfter)
		if err != nil {
			t.Fatalf("Failed to query challenge 2 progress after delete: %v", err)
		}

		if progressAfter != progressBefore-wVal {
			t.Errorf("Expected progress to decrease by %d to %d, got %d", wVal, progressBefore-wVal, progressAfter)
		}
	})

	// TC-3.12 (Positive): Каскадный откат completed → active
	t.Run("TC-3.12 Positive: Cascade rollback completed -> active", func(t *testing.T) {
		var statusBefore string
		err := conn.QueryRow(context.Background(), "SELECT status FROM challenges WHERE id = 1").Scan(&statusBefore)
		if err != nil {
			t.Fatalf("Failed to query challenge 1: %v", err)
		}
		if statusBefore != "completed" {
			t.Fatalf("Challenge 1 is not in 'completed' state, current state is: %s", statusBefore)
		}

		resp, body := sendRequest(t, "DELETE", fmt.Sprintf("/api/workouts/%d", workout3ID), "default_user_1", nil)
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d. Body: %s", resp.StatusCode, string(body))
		}

		var statusAfter string
		var progressAfter int
		err = conn.QueryRow(context.Background(), "SELECT status, current_progress FROM challenges WHERE id = 1").Scan(&statusAfter, &progressAfter)
		if err != nil {
			t.Fatalf("Failed to query challenge 1: %v", err)
		}

		if statusAfter != "active" {
			t.Errorf("Expected status to be 'active', got %s", statusAfter)
		}
		if progressAfter != 50 {
			t.Errorf("Expected progress to be 50, got %d", progressAfter)
		}
	})

	// TC-3.13 (Negative): Несуществующая тренировка → 404
	t.Run("TC-3.13 Negative: Delete non-existent workout", func(t *testing.T) {
		resp, body := sendRequest(t, "DELETE", "/api/workouts/99999", "default_user_1", nil)
		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("Expected status 404, got %d. Body: %s", resp.StatusCode, string(body))
		}
	})

	// TC-3.14 (Negative): Чужая тренировка → 404
	t.Run("TC-3.14 Negative: Delete other user's workout", func(t *testing.T) {
		resp, body := sendRequest(t, "DELETE", fmt.Sprintf("/api/workouts/%d", workout2ID), "user_b", nil)
		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("Expected status 404, got %d. Body: %s", resp.StatusCode, string(body))
		}
	})

	// TC-3.15 (Edge): current_progress не уходит ниже 0
	t.Run("TC-3.15 Edge: progress never below 0", func(t *testing.T) {
		c3Payload := map[string]interface{}{
			"name":        "Challenge 3",
			"exercise_id": 1,
			"target_value": 10,
			"start_date":  "2026-06-01T00:00:00Z",
			"end_date":    "2026-07-01T00:00:00Z",
		}
		resp, body := sendRequest(t, "POST", "/api/challenges", "default_user_1", c3Payload)
		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("Failed to create Challenge 3: %s", string(body))
		}

		payload := map[string]interface{}{
			"workout_date": "2026-06-25",
			"value":        10,
		}
		resp, body = sendRequest(t, "POST", "/api/challenges/3/workouts", "default_user_1", payload)
		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("Failed to add workout: %s", string(body))
		}

		var respData map[string]interface{}
		json.Unmarshal(body, &respData)
		workout := respData["workout"].(map[string]interface{})
		wID := int(workout["id"].(float64))

		_, err := conn.Exec(context.Background(), "UPDATE challenges SET current_progress = 5 WHERE id = 3")
		if err != nil {
			t.Fatalf("Failed to set progress to 5 in DB: %v", err)
		}

		resp, body = sendRequest(t, "DELETE", fmt.Sprintf("/api/workouts/%d", wID), "default_user_1", nil)
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d. Body: %s", resp.StatusCode, string(body))
		}

		var progress int
		err = conn.QueryRow(context.Background(), "SELECT current_progress FROM challenges WHERE id = 3").Scan(&progress)
		if err != nil {
			t.Fatalf("Failed to query challenge 3 progress: %v", err)
		}
		if progress != 0 {
			t.Errorf("Expected current_progress to be 0, got %d", progress)
		}
	})

	// TC-3.16 (Positive): Получение списка ачивок челленджа
	t.Run("TC-3.16 Positive: List user achievements", func(t *testing.T) {
		resp, body := sendRequest(t, "GET", "/api/challenges/2/achievements", "default_user_1", nil)
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d. Body: %s", resp.StatusCode, string(body))
		}

		var respData []map[string]interface{}
		if err := json.Unmarshal(body, &respData); err != nil {
			t.Fatalf("Failed to unmarshal body: %v", err)
		}

		if len(respData) == 0 {
			t.Errorf("Expected to get achievements list, got empty")
		}

		codes := map[string]bool{}
		for _, ach := range respData {
			codes[ach["achievement_code"].(string)] = true
		}

		expectedCodes := []string{"first_step", "stability"}
		for _, c := range expectedCodes {
			if !codes[c] {
				t.Errorf("Expected achievement %s to be in list, but it wasn't", c)
			}
		}
	})

	// TC-3.17 (Positive): Пустой список → [], не null
	t.Run("TC-3.17 Positive: Empty achievements list", func(t *testing.T) {
		// Create Challenge 3 (new challenge, no workouts added)
		c3Payload := map[string]interface{}{
			"name":         "Challenge 3",
			"exercise_id":  1,
			"target_value": 100,
			"start_date":   "2026-06-01T00:00:00Z",
			"end_date":     "2026-07-01T00:00:00Z",
		}
		resp, body := sendRequest(t, "POST", "/api/challenges", "default_user_1", c3Payload)
		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("Failed to create Challenge 3, status %d. Body: %s", resp.StatusCode, string(body))
		}
		var c3Data map[string]interface{}
		json.Unmarshal(body, &c3Data)
		c3ID := int(c3Data["id"].(float64))

		resp, body = sendRequest(t, "GET", fmt.Sprintf("/api/challenges/%d/achievements", c3ID), "default_user_1", nil)
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d. Body: %s", resp.StatusCode, string(body))
		}

		if string(body) != "[]\n" && string(body) != "[]" {
			t.Errorf("Expected empty JSON array '[]', got '%s'", string(body))
		}
	})

	// TC-3.18 (Positive): Удаление существующего челленджа
	t.Run("TC-3.18 Positive: Delete challenge removes it and its workouts", func(t *testing.T) {
		// Create a challenge to delete (use default_user_1 which has seeded exercises)
		challengePayload := map[string]interface{}{
			"name":         "Challenge to Delete",
			"exercise_id":  1,
			"target_value": 100,
			"start_date":   "2025-01-01T00:00:00Z",
			"end_date":     "2025-03-01T00:00:00Z",
		}
		resp, body := sendRequest(t, "POST", "/api/challenges", "default_user_1", challengePayload)
		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("Setup: failed to create challenge. Status: %d, body: %s", resp.StatusCode, string(body))
		}

		var created map[string]interface{}
		if err := json.Unmarshal(body, &created); err != nil {
			t.Fatalf("Setup: failed to unmarshal created challenge: %v", err)
		}
		challengeID := int(created["id"].(float64))

		// Delete the challenge
		resp, body = sendRequest(t, "DELETE", fmt.Sprintf("/api/challenges/%d", challengeID), "default_user_1", nil)
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected 200 on DELETE, got %d. Body: %s", resp.StatusCode, string(body))
		}

		// Verify it's gone: GET should return 404
		resp, body = sendRequest(t, "GET", fmt.Sprintf("/api/challenges/%d", challengeID), "default_user_1", nil)
		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("Expected 404 after deletion, got %d. Body: %s", resp.StatusCode, string(body))
		}
	})

	// TC-3.18 (Negative): Удаление несуществующего / чужого челленджа
	t.Run("TC-3.18 Negative: Delete non-existent challenge returns 404", func(t *testing.T) {
		resp, body := sendRequest(t, "DELETE", "/api/challenges/99999999", "some_other_user", nil)
		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("Expected 404 for non-existent challenge, got %d. Body: %s", resp.StatusCode, string(body))
		}
	})

	// TC-3.19 (Positive): Удаление тренировки откатывает ачивки
	t.Run("TC-3.19 Positive: Delete workout revokes achievements", func(t *testing.T) {
		// Challenge 4
		c4Payload := map[string]interface{}{
			"name":         "Challenge 4",
			"exercise_id":  1,
			"target_value": 100,
			"start_date":   "2026-06-01T00:00:00Z",
			"end_date":     "2026-07-01T00:00:00Z",
		}
		resp, body := sendRequest(t, "POST", "/api/challenges", "default_user_1", c4Payload)
		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("Failed to create Challenge 4: %s", string(body))
		}
		var c4Data map[string]interface{}
		json.Unmarshal(body, &c4Data)
		c4ID := int(c4Data["id"].(float64))

		// Add workout that triggers equator and first_step
		wPayload := map[string]interface{}{
			"workout_date": "2026-06-25",
			"value":        60,
		}
		resp, body = sendRequest(t, "POST", fmt.Sprintf("/api/challenges/%d/workouts", c4ID), "default_user_1", wPayload)
		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("Failed to add workout: %s", string(body))
		}
		var wData map[string]interface{}
		json.Unmarshal(body, &wData)
		w := wData["workout"].(map[string]interface{})
		wID := int(w["id"].(float64))

		// Verify equator is unlocked
		resp, body = sendRequest(t, "GET", fmt.Sprintf("/api/challenges/%d/achievements", c4ID), "default_user_1", nil)
		var achList []map[string]interface{}
		json.Unmarshal(body, &achList)
		foundEquator := false
		for _, ach := range achList {
			if ach["achievement_code"] == "equator" {
				foundEquator = true
			}
		}
		if !foundEquator {
			t.Fatalf("Equator not unlocked")
		}

		// Delete workout
		resp, _ = sendRequest(t, "DELETE", fmt.Sprintf("/api/workouts/%d", wID), "default_user_1", nil)
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Failed to delete workout")
		}

		// Verify equator and first_step are revoked
		resp, body = sendRequest(t, "GET", fmt.Sprintf("/api/challenges/%d/achievements", c4ID), "default_user_1", nil)
		json.Unmarshal(body, &achList)
		for _, ach := range achList {
			if ach["achievement_code"] == "equator" {
				t.Errorf("Equator achievement was not revoked")
			}
			if ach["achievement_code"] == "first_step" {
				t.Errorf("First step achievement was not revoked")
			}
		}
	})
}


func TestAPI_Sprint6(t *testing.T) {
	conn := getDBConn(t)
	defer conn.Close(context.Background())

	t.Run("Reset DB", func(t *testing.T) {
		resetDatabase(t, conn)
	})

	// Epic: US-1 Создание упражнения
	// TC-1.1 (Positive): Успешное создание
	t.Run("TC-1.1 Positive: Create custom exercise", func(t *testing.T) {
		payload := map[string]interface{}{
			"name": "Бег",
		}
		resp, body := sendRequest(t, "POST", "/api/exercises", "test_user", payload)
		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("Expected status 201, got %d. Body: %s", resp.StatusCode, string(body))
		}

		var respData map[string]interface{}
		if err := json.Unmarshal(body, &respData); err != nil {
			t.Fatalf("Failed to unmarshal body: %v", err)
		}

		if respData["name"] != "Бег" {
			t.Errorf("Expected exercise name 'Бег', got %v", respData["name"])
		}
		if respData["is_custom"] != true {
			t.Errorf("Expected is_custom to be true, got %v", respData["is_custom"])
		}
		if respData["user_id"] != "test_user" {
			t.Errorf("Expected user_id 'test_user', got %v", respData["user_id"])
		}

		// Verify in DB
		var name string
		var isCustom bool
		var userID string
		err := conn.QueryRow(context.Background(), "SELECT name, is_custom, user_id FROM exercises WHERE name = 'Бег' AND user_id = 'test_user'").Scan(&name, &isCustom, &userID)
		if err != nil {
			t.Fatalf("Failed to query exercise from DB: %v", err)
		}
		if name != "Бег" || !isCustom || userID != "test_user" {
			t.Errorf("DB values mismatch: name=%s, isCustom=%t, userID=%s", name, isCustom, userID)
		}
	})

	// TC-1.2 (Negative): Пустое имя
	t.Run("TC-1.2 Negative: Empty and whitespace name", func(t *testing.T) {
		payloadEmpty := map[string]interface{}{
			"name": "",
		}
		resp, body := sendRequest(t, "POST", "/api/exercises", "test_user", payloadEmpty)
		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected status 400 for empty name, got %d. Body: %s", resp.StatusCode, string(body))
		}

		payloadSpace := map[string]interface{}{
			"name": "   ",
		}
		resp, body = sendRequest(t, "POST", "/api/exercises", "test_user", payloadSpace)
		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected status 400 for whitespace name, got %d. Body: %s", resp.StatusCode, string(body))
		}
	})

	// TC-1.3 (Negative): Дубликат имени
	t.Run("TC-1.3 Negative: Duplicate exercise name", func(t *testing.T) {
		payload := map[string]interface{}{
			"name": "Бег",
		}
		resp, body := sendRequest(t, "POST", "/api/exercises", "test_user", payload)
		if resp.StatusCode != http.StatusConflict && resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected status 409 (Conflict) or 400, got %d. Body: %s", resp.StatusCode, string(body))
		}
	})

	// Epic: US-2, US-5 Создание челленджа и Дашборд
	// TC-2.1 (Positive): Успешное создание челленджа
	t.Run("TC-2.1 Positive: Create challenge successfully", func(t *testing.T) {
		payload := map[string]interface{}{
			"name":         "3000 отжиманий",
			"exercise_id":  1,
			"target_value": 3000,
			"start_date":   "2026-06-01T00:00:00Z",
			"end_date":     "2026-06-30T00:00:00Z",
		}
		resp, body := sendRequest(t, "POST", "/api/challenges", "default_user_1", payload)
		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("Expected status 201, got %d. Body: %s", resp.StatusCode, string(body))
		}

		var respData map[string]interface{}
		if err := json.Unmarshal(body, &respData); err != nil {
			t.Fatalf("Failed to unmarshal body: %v", err)
		}

		if respData["name"] != "3000 отжиманий" {
			t.Errorf("Expected challenge name '3000 отжиманий', got %v", respData["name"])
		}
		if respData["status"] != "active" {
			t.Errorf("Expected status to be 'active', got %v", respData["status"])
		}
		if respData["current_progress"] != float64(0) {
			t.Errorf("Expected current_progress to be 0, got %v", respData["current_progress"])
		}

		// Verify in DB
		var dbStatus string
		var dbProgress int
		err := conn.QueryRow(context.Background(), "SELECT status, current_progress FROM challenges WHERE name = '3000 отжиманий'").Scan(&dbStatus, &dbProgress)
		if err != nil {
			t.Fatalf("Failed to query challenge from DB: %v", err)
		}
		if dbStatus != "active" || dbProgress != 0 {
			t.Errorf("DB values mismatch: status=%s, progress=%d", dbStatus, dbProgress)
		}
	})

	// TC-2.2 (Negative): Дедлайн раньше старта
	t.Run("TC-2.2 Negative: Deadline earlier than start date", func(t *testing.T) {
		payload := map[string]interface{}{
			"name":         "Invalid Date Challenge",
			"exercise_id":  1,
			"target_value": 100,
			"start_date":   "2026-06-30T00:00:00Z",
			"end_date":     "2026-06-01T00:00:00Z",
		}
		resp, body := sendRequest(t, "POST", "/api/challenges", "default_user_1", payload)
		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d. Body: %s", resp.StatusCode, string(body))
		}
	})

	// TC-2.3 (Negative): Целевое количество <= 0
	t.Run("TC-2.3 Negative: Target value <= 0", func(t *testing.T) {
		payloadZero := map[string]interface{}{
			"name":         "Zero Target Challenge",
			"exercise_id":  1,
			"target_value": 0,
			"start_date":   "2026-06-01T00:00:00Z",
			"end_date":     "2026-06-30T00:00:00Z",
		}
		resp, body := sendRequest(t, "POST", "/api/challenges", "default_user_1", payloadZero)
		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected status 400 for target_value = 0, got %d. Body: %s", resp.StatusCode, string(body))
		}

		payloadNegative := map[string]interface{}{
			"name":         "Negative Target Challenge",
			"exercise_id":  1,
			"target_value": -100,
			"start_date":   "2026-06-01T00:00:00Z",
			"end_date":     "2026-06-30T00:00:00Z",
		}
		resp, body = sendRequest(t, "POST", "/api/challenges", "default_user_1", payloadNegative)
		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected status 400 for target_value = -100, got %d. Body: %s", resp.StatusCode, string(body))
		}
	})
}

