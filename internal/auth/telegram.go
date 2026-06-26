package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

// ValidateTelegramData validates the initData string against the bot token.
func ValidateTelegramData(initData, token string) (bool, error) {
	parsed, err := url.ParseQuery(initData)
	if err != nil {
		return false, err
	}

	hash := parsed.Get("hash")
	if hash == "" {
		return false, fmt.Errorf("missing hash")
	}

	var dataCheckArr []string
	for k, v := range parsed {
		if k == "hash" {
			continue
		}
		dataCheckArr = append(dataCheckArr, fmt.Sprintf("%s=%s", k, v[0]))
	}

	sort.Strings(dataCheckArr)
	dataCheckString := strings.Join(dataCheckArr, "\n")

	secretKeyMac := hmac.New(sha256.New, []byte("WebAppData"))
	secretKeyMac.Write([]byte(token))
	secretKey := secretKeyMac.Sum(nil)

	mac := hmac.New(sha256.New, secretKey)
	mac.Write([]byte(dataCheckString))
	calculatedHash := hex.EncodeToString(mac.Sum(nil))

	if calculatedHash != hash {
		return false, nil
	}

	authDateStr := parsed.Get("auth_date")
	if authDateStr == "" {
		return false, fmt.Errorf("missing auth_date")
	}

	authDateInt, err := strconv.ParseInt(authDateStr, 10, 64)
	if err != nil {
		return false, fmt.Errorf("invalid auth_date")
	}

	authDate := time.Unix(authDateInt, 0)
	if time.Since(authDate) > 24*time.Hour {
		return false, fmt.Errorf("initData expired")
	}

	return true, nil
}

// GetUserIDFromInitData parses the initData and extracts the user ID.
func GetUserIDFromInitData(initData string) (string, error) {
	parsed, err := url.ParseQuery(initData)
	if err != nil {
		return "", err
	}
	userStr := parsed.Get("user")
	if userStr == "" {
		return "", fmt.Errorf("missing user field")
	}
	var user map[string]interface{}
	if err := json.Unmarshal([]byte(userStr), &user); err != nil {
		return "", err
	}
	idFloat, ok := user["id"].(float64)
	if !ok {
		return "", fmt.Errorf("invalid id type")
	}
	return fmt.Sprintf("%.0f", idFloat), nil
}
