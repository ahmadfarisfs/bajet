package mwr

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
)

const sessionUserIDKey = "user_id"

func SetSessionUserID(c echo.Context, userID int64) error {
	sess, err := session.Get("session", c)
	if err != nil {
		return err
	}
	sess.Values[sessionUserIDKey] = strconv.FormatInt(userID, 10)
	return sess.Save(c.Request(), c.Response())
}

func ClearSession(c echo.Context) error {
	sess, err := session.Get("session", c)
	if err != nil {
		return err
	}
	sess.Options.MaxAge = -1
	return sess.Save(c.Request(), c.Response())
}

func GetSessionUserID(c echo.Context) (int64, bool) {
	sess, err := session.Get("session", c)
	if err != nil {
		return 0, false
	}
	raw, ok := sess.Values[sessionUserIDKey].(string)
	if !ok || raw == "" {
		return 0, false
	}
	userID, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return 0, false
	}
	return userID, true
}

func RequireSessionAuth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		userID, ok := GetSessionUserID(c)
		if !ok {
			return c.Redirect(http.StatusSeeOther, "/login")
		}
		c.Set(sessionUserIDKey, userID)
		return next(c)
	}
}
