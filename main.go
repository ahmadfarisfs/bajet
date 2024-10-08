package main

import (
	"log"
	"net/http"
	"os"

	"bajetapp/routes"
	"bajetapp/utils"
	"bajetapp/views"

	"github.com/a-h/templ"
	"github.com/gorilla/sessions"
	"github.com/joho/godotenv"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func init() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}
}

func main() {
	e := echo.New()

	// Add logging middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Initialize session middleware
	sessionSecret := os.Getenv("SESSION_SECRET")
	if sessionSecret == "" {
		log.Fatal("SESSION_SECRET environment variable is required")
	}
	e.Use(session.Middleware(sessions.NewCookieStore([]byte(sessionSecret))))

	// Routes
	e.GET("/", handleMain)
	routes.NewAuthRoutes(os.Getenv("GOOGLE_CLIENT_ID"), os.Getenv("GOOGLE_CLIENT_SECRET"), os.Getenv("REDIRECT_URL"), os.Getenv("OAUTH_STATE_STRING"), e)

	e.Logger.Fatal(e.Start(":8080"))
}

// Home page
func handleMain(c echo.Context) error {
	if utils.IsAuthenticated(c) {
		c.Redirect(http.StatusTemporaryRedirect, "/profile")
	}
	// html := `<html><body><a href="/auth/login">Google Log In</a></body></html>`
	// return c.HTML(http.StatusOK, html)
	cpmnt := views.Index("Home", "Welcome ")
	Render(c, &cpmnt)
	return nil
}

func Render(echoCtx echo.Context, component *templ.Component) {
	(*component).Render(echoCtx.Request().Context(), echoCtx.Response().Writer)
}
