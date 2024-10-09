package routes

import (
	"bajetapp/utils"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type AuthRoutes struct {
	googleOauthConfig *oauth2.Config
}

func NewAuthRoutes(googleClientID, googleClientSecret, redirectURL string, e *echo.Echo) *AuthRoutes {
	route := &AuthRoutes{
		googleOauthConfig: &oauth2.Config{
			ClientID:     googleClientID,
			ClientSecret: googleClientSecret,
			RedirectURL:  redirectURL,
			Scopes: []string{
				"https://www.googleapis.com/auth/userinfo.profile",
				"https://www.googleapis.com/auth/userinfo.email",
			},
			Endpoint: google.Endpoint,
		},
	}
	route.registerRoutes(e)
	return route
}

func (ar *AuthRoutes) registerRoutes(e *echo.Echo) {
	e.GET("/auth/login", ar.handleGoogleLogin)
	e.GET("/auth/callback", ar.handleGoogleCallback)
	e.GET("/auth/logout", ar.handleLogout)
	e.GET("/profile", ar.handleProfile)
}

// Google login handler
func (ar *AuthRoutes) handleGoogleLogin(c echo.Context) error {
	// Generate a random OAuth state string
	oauthStateString, err := utils.GenerateRandomString(32)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to generate OAuth state string")
	}
	// Store the OAuth state string in the session
	sess, _ := session.Get("session", c)
	sess.Values["oauthStateString"] = oauthStateString
	if err := sess.Save(c.Request(), c.Response()); err != nil {
		return c.String(http.StatusInternalServerError, fmt.Sprintf("Failed to save session: %s", err.Error()))
	}

	url := ar.googleOauthConfig.AuthCodeURL(oauthStateString)
	return c.Redirect(http.StatusTemporaryRedirect, url)
}

// Google callback handler
func (ar *AuthRoutes) handleGoogleCallback(c echo.Context) error {
	state := c.QueryParam("state")

	// Retrieve the OAuth state string from the session
	sess, _ := session.Get("session", c)
	oauthStateString, ok := sess.Values["oauthStateString"].(string)
	if !ok || state != oauthStateString {
		return c.String(http.StatusBadRequest, "Invalid OAuth state string")
	}

	code := c.QueryParam("code")
	token, err := ar.googleOauthConfig.Exchange(context.Background(), code)
	if err != nil {
		log.Println("Failed to exchange token:", err)
		return c.Redirect(http.StatusTemporaryRedirect, "/")
	}

	client := ar.googleOauthConfig.Client(context.Background(), token)
	userInfoResp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		log.Println("Failed to get user info:", err)
		return c.Redirect(http.StatusTemporaryRedirect, "/")
	}
	defer userInfoResp.Body.Close()

	userInfo, err := io.ReadAll(userInfoResp.Body)
	if err != nil {
		log.Println("Failed to read user info:", err)
		return c.Redirect(http.StatusTemporaryRedirect, "/")
	}

	// Store the user info in the session
	sess, _ = session.Get("session", c)
	sess.Values["user_info"] = string(userInfo)
	sess.Save(c.Request(), c.Response())

	return c.Redirect(http.StatusTemporaryRedirect, "/profile")
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
