package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
)

const userIDKey = "user_id"

type tokenCacheEntry struct {
	sub     string
	expires time.Time
}

var (
	tokenCache   = map[string]tokenCacheEntry{}
	tokenCacheMu sync.RWMutex
)

// AuthMiddleware verifies Google ID tokens sent as "Authorization: Bearer <token>".
func AuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		authHeader := c.Request().Header.Get("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "missing authorization header"})
		}
		token := strings.TrimPrefix(authHeader, "Bearer ")
		sub, err := verifyGoogleToken(c.Request().Context(), token)
		if err != nil {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid token"})
		}
		c.Set(userIDKey, sub)
		return next(c)
	}
}

// userID returns the authenticated user's Google subject claim from the context.
func userID(c echo.Context) string {
	v, _ := c.Get(userIDKey).(string)
	return v
}

// verifyGoogleToken calls Google's tokeninfo endpoint and caches valid results for 5 minutes.
func verifyGoogleToken(ctx context.Context, token string) (string, error) {
	tokenCacheMu.RLock()
	if entry, ok := tokenCache[token]; ok && time.Now().Before(entry.expires) {
		tokenCacheMu.RUnlock()
		return entry.sub, nil
	}
	tokenCacheMu.RUnlock()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		"https://oauth2.googleapis.com/tokeninfo?id_token="+url.QueryEscape(token), nil)
	if err != nil {
		return "", fmt.Errorf("failed to build tokeninfo request: %w", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("tokeninfo request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("invalid token (status %d)", resp.StatusCode)
	}

	var payload struct {
		Sub string `json:"sub"`
		Aud string `json:"aud"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return "", fmt.Errorf("failed to decode tokeninfo response: %w", err)
	}

	clientID := os.Getenv("GOOGLE_CLIENT_ID")
	if clientID == "" {
		return "", fmt.Errorf("GOOGLE_CLIENT_ID env var not set")
	}
	if payload.Aud != clientID {
		return "", fmt.Errorf("token audience mismatch")
	}
	if payload.Sub == "" {
		return "", fmt.Errorf("token missing subject claim")
	}

	tokenCacheMu.Lock()
	tokenCache[token] = tokenCacheEntry{sub: payload.Sub, expires: time.Now().Add(5 * time.Minute)}
	tokenCacheMu.Unlock()

	return payload.Sub, nil
}
