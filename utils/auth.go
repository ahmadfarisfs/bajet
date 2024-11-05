package utils

import (
	"crypto/rand"
	"encoding/base64"

	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
)

// IsAuthenticated checks if the user is authenticated by looking for user_info in the session
func IsAuthenticated(c echo.Context) bool {
	sess, _ := session.Get("session", c)
	userInfo := sess.Values["user_info"]
	return userInfo != nil
}

func GenerateRandomString(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
