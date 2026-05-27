package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

const userIDKey = "user_id"

// ── JWT helpers ───────────────────────────────────────────────────────────────

func jwtSecret() []byte {
	s := os.Getenv("JWT_SECRET")
	if s == "" {
		s = "bajet-secret-change-in-production"
	}
	return []byte(s)
}

type appClaims struct {
	Name    string `json:"name"`
	Email   string `json:"email"`
	Picture string `json:"picture"`
	jwt.RegisteredClaims
}

const tokenTTL = 30 * 24 * time.Hour
const renewWindow = 7 * 24 * time.Hour

func issueToken(sub, name, email, picture string) (string, error) {
	claims := appClaims{
		Name:    name,
		Email:   email,
		Picture: picture,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   sub,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(tokenTTL)),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(jwtSecret())
}

func parseToken(tokenStr string) (*appClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &appClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return jwtSecret(), nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*appClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	return claims, nil
}

// ── Middleware ────────────────────────────────────────────────────────────────

// AuthMiddleware verifies the custom JWT and silently renews it when it is
// within renewWindow of expiry (sliding session).
func AuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		authHeader := c.Request().Header.Get("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "missing authorization header"})
		}
		claims, err := parseToken(strings.TrimPrefix(authHeader, "Bearer "))
		if err != nil {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid token"})
		}
		c.Set(userIDKey, claims.Subject)

		// Sliding expiry: renew if token expires within renewWindow.
		if time.Until(claims.ExpiresAt.Time) < renewWindow {
			if fresh, err := issueToken(claims.Subject, claims.Name, claims.Email, claims.Picture); err == nil {
				c.Response().Header().Set("X-Refresh-Token", fresh)
			}
		}

		return next(c)
	}
}

func userID(c echo.Context) string {
	v, _ := c.Get(userIDKey).(string)
	return v
}

// ── Google sign-in exchange ───────────────────────────────────────────────────

// GoogleSignIn verifies a Google ID token and returns a 30-day custom JWT.
func GoogleSignIn(c echo.Context) error {
	var body struct {
		Credential string `json:"credential"`
	}
	if err := c.Bind(&body); err != nil || body.Credential == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "credential required"})
	}
	info, err := verifyGoogleToken(c.Request().Context(), body.Credential)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid Google credential"})
	}
	token, err := issueToken(info.Sub, info.Name, info.Email, info.Picture)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to issue token"})
	}
	return c.JSON(http.StatusOK, map[string]string{
		"token":   token,
		"name":    info.Name,
		"email":   info.Email,
		"picture": info.Picture,
	})
}

// ── Google tokeninfo verification ────────────────────────────────────────────

type googleInfo struct {
	Sub     string `json:"sub"`
	Aud     string `json:"aud"`
	Name    string `json:"name"`
	Email   string `json:"email"`
	Picture string `json:"picture"`
}

func verifyGoogleToken(ctx context.Context, token string) (*googleInfo, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		"https://oauth2.googleapis.com/tokeninfo?id_token="+url.QueryEscape(token), nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("invalid Google token (status %d)", resp.StatusCode)
	}
	var info googleInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, err
	}
	clientID := os.Getenv("GOOGLE_CLIENT_ID")
	if clientID != "" && info.Aud != clientID {
		return nil, fmt.Errorf("token audience mismatch")
	}
	if info.Sub == "" {
		return nil, fmt.Errorf("token missing subject")
	}
	return &info, nil
}
