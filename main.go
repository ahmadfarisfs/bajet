package main

import (
	"log"
	"net/http"

	"bajetapp/config"
	"bajetapp/routes"
	"bajetapp/utils"
	"bajetapp/views/pages"

	"bajetapp/db"

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
	config, err := config.LoadConfig()
	if err != nil {
		log.Fatal("error loading config: %w", err)
	}

	// Initialize Database
	client, mongo, err := db.ConnectMongoDB(config.MongoURI, config.MongoDatabase)
	if err != nil {
		log.Fatal("error connecting to MongoDB: %w", err)
	}

	e := echo.New()

	// Add logging middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(session.Middleware(sessions.NewCookieStore([]byte(config.SessionSecret))))

	staticRoot := e.Group("/static")
	staticRoot.Use(middleware.StaticWithConfig(middleware.StaticConfig{
		Root:   "./views/static",
		Browse: false,
	}))

	// Routes
	e.GET("/", handleLogin)
	e.GET("/main", handleMain)
	routes.NewAuthRoutes(config.GoogleClientID, config.GoogleClientSecret, config.RedirectURL, e)

	e.Logger.Fatal(e.Start(":8080"))
}

func handleMain(c echo.Context) error {
	if !utils.IsAuthenticated(c) {
		c.Redirect(http.StatusTemporaryRedirect, "/")
	}

	cpmnt := pages.MainPage()
	Render(c, &cpmnt)
	return nil
}

// Login page
func handleLogin(c echo.Context) error {
	if utils.IsAuthenticated(c) {
		c.Redirect(http.StatusTemporaryRedirect, "/profile")
	}

	cpmnt := pages.Login()
	Render(c, &cpmnt)
	return nil
}

func Render(echoCtx echo.Context, component *templ.Component) {
	(*component).Render(echoCtx.Request().Context(), echoCtx.Response().Writer)
}
