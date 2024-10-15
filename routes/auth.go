package routes

import (
	"bajetapp/services"
	"bajetapp/utils"
	"fmt"
	"net/http"

	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
)

type AuthRoutes struct {
	svc *services.AuthService
}

func NewAuthRoutes(svc *services.AuthService, e *echo.Echo) *AuthRoutes {
	route := &AuthRoutes{
		svc: svc,
	}
	route.registerRoutes(e)
	return route
}

func (ar *AuthRoutes) registerRoutes(e *echo.Echo) {
	e.GET("/auth/login", ar.handleGoogleLogin)
	e.GET("/auth/callback", ar.handleGoogleCallback)
	e.GET("/auth/logout", ar.handleLogout) //TODO:use auth middleware
	e.GET("/profile", ar.handleProfile)    //TODO:use auth middleware
}

// Google login handler
func (ar *AuthRoutes) handleGoogleLogin(c echo.Context) error {
	redirectUrl, err := ar.svc.GoogleLogin(c)
	if err != nil {
		// TODO: change with error template
		return c.String(http.StatusInternalServerError, fmt.Sprintf("failed to save session: %s", err.Error()))
	}
	return c.Redirect(http.StatusTemporaryRedirect, redirectUrl)
}

// Google callback handler
func (ar *AuthRoutes) handleGoogleCallback(c echo.Context) error {
	state := c.QueryParam("state")
	code := c.QueryParam("code")
	_, err := ar.svc.ProcessGoogleCallback(c, state, code)
	if err != nil {
		// TODO: change with error template
		return c.String(http.StatusInternalServerError, fmt.Sprintf("failed to process Google callback: %s", err.Error()))
	}
	return c.Redirect(http.StatusTemporaryRedirect, "/main")
}

// Protected profile page
func (ar *AuthRoutes) handleProfile(c echo.Context) error {
	if !utils.IsAuthenticated(c) {
		return c.Redirect(http.StatusTemporaryRedirect, "/")
	}
	sess, _ := session.Get("session", c)
	userInfo := sess.Values["user_info"]
	html := fmt.Sprintf(`
	<html>
	<body>
		<h1>User Profile</h1>
		<p>%v</p>
		<form action="/auth/logout" method="GET">
			<button type="submit">Logout</button>
		</form>
	</body>
	</html>`, userInfo)
	return c.HTML(http.StatusOK, html)
}

// Logout handler
func (ar *AuthRoutes) handleLogout(c echo.Context) error {
	sess, _ := session.Get("session", c)
	sess.Options.MaxAge = -1 // This will delete the session
	sess.Save(c.Request(), c.Response())
	return c.Redirect(http.StatusTemporaryRedirect, "/")
}
