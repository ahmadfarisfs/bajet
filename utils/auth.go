package utils

import (
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
)

// IsAuthenticated checks if the user is authenticated by looking for user_info in the session
func IsAuthenticated(c echo.Context) bool {
	sess, _ := session.Get("session", c)
	userInfo := sess.Values["user_info"]
	return userInfo != nil
}
