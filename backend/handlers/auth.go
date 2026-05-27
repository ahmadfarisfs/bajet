package handlers

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
)

const userIDKey = "user_id"

// ── Custom JWT (HS256) ────────────────────────────────────────────────────────

func jwtSecret() []byte {
	s := os.Getenv("JWT_SECRET")
	if s == "" {
		s = "bajet-secret-change-in-production"
	}
	return []byte(s)
}

func b64url(b []byte) string {
	return base64.RawURLEncoding.EncodeToString(b)
}

type tokenClaims struct {
	Sub     string `json:"sub"`
	Name    string `json:"name"`
	Email   string `json:"email"`
	Picture string `json:"picture"`
	Exp     int64  `json:"exp"`
}

func issueToken(sub, name, email, picture string) (string, error) {
	header, _ := json.Marshal(map[string]string{"alg": "HS256", "typ": "JWT"})
	claims, _ := json.Marshal(tokenClaims{
		Sub:     sub,
		Name:    name,
		Email:   email,
		Picture: picture,
		Exp:     time.Now().Add(30 * 24 * time.Hour).Unix(),
	})
	unsigned := b64url(header) + "." + b64url(claims)
	mac := hmac.New(sha256.New, jwtSecret())
	mac.Write([]byte(unsigned))
	return unsigned + "." + b64url(mac.Sum(nil)), nil
}

func parseToken(token string) (*tokenClaims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid token format")
	}
	unsigned := parts[0] + "." + parts[1]
	mac := hmac.New(sha256.New, jwtSecret())
	mac.Write([]byte(unsigned))
	if !hmac.Equal([]byte(b64url(mac.Sum(nil))), []byte(parts[2])) {
		return nil, fmt.Errorf("invalid signature")
	}
	raw, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("invalid payload")
	}
	var c tokenClaims
	if err := json.Unmarshal(raw, &c); err != nil {
		return nil, fmt.Errorf("invalid claims")
	}
	if time.Now().Unix() > c.Exp {
		return nil, fmt.Errorf("token expired")
	}
	return &c, nil
}

// ── Middleware ────────────────────────────────────────────────────────────────

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
		c.Set(userIDKey, claims.Sub)
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
