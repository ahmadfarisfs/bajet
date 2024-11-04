package mwr

import (
	"bajetapp/model"
	"encoding/json"
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

func GetLoginInfo(c echo.Context) *model.GoogleUserInfo {
	userInfo := c.Get("user_info")
	if userInfo == nil {
		return nil
	}

	userInfoBytes, err := json.Marshal(userInfo)
	if err != nil {
		return nil
	}

	var user model.GoogleUserInfo
	err = json.Unmarshal(userInfoBytes, &user)
	if err != nil {
		return nil
	}

	return &user
}
