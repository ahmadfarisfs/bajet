package mwr

import (
	"bajetapp/model"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
)

func AuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		sess, err := session.Get("session", c)
		if err != nil {
			return c.Redirect(http.StatusTemporaryRedirect, "/")
		}
		userInfo := sess.Values["user_info"]
		if userInfo == nil {
			return c.Redirect(http.StatusTemporaryRedirect, "/")
		}
		c.Set("user_info", userInfo)
		return next(c)
	}
}

func GetLoginInfo(c echo.Context) (*model.GoogleUserInfo, error) {
	userInfo := c.Get("user_info")
	if userInfo == nil {
		return nil, fmt.Errorf("user info not found")
	}

	var user model.GoogleUserInfo
	if err := json.Unmarshal([]byte(userInfo.(string)), &user); err != nil {
		return nil, fmt.Errorf("failed to unmarshal user info: %w", err)
	}

	return &user, nil
}
