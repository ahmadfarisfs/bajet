package main

import (
	"database/sql"
	"log"
	"net/http"
	"time"

	"bajetapp/config"
	"bajetapp/db"
	"bajetapp/routes"
	"bajetapp/services"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

var (
	cycleService *services.CycleService
	authService  *services.AuthLocalService
)

type GoPlaygorundValidator struct {
	validator *validator.Validate
}

func (cv *GoPlaygorundValidator) Validate(i interface{}) error {
	if err := cv.validator.Struct(i); err != nil {
		// Optionally, you could return the error to give each route more control over the status code
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return nil
}

func main() {
	config, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("error loading config: %v", err)
	}

	// Initialize SQLite database.
	database, err := db.ConnectSQLite(config.SQLitePath)
	if err != nil {
		log.Fatalf("error connecting to SQLite: %v", err)
	}
	defer func(database *sql.DB) {
		if err := database.Close(); err != nil {
			log.Printf("error closing sqlite database: %v", err)
		}
	}(database)

	if err := db.InitSQLiteSchema(database); err != nil {
		log.Fatalf("error initializing SQLite schema: %v", err)
	}

	e := echo.New()
	e.Use(middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(10)))
	e.Use(middleware.RequestID())
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.Gzip())
	e.Use(middleware.CORS())
	e.Use(middleware.Secure())
	e.Use(middleware.BodyLimit("2M"))
	e.Use(middleware.CSRFWithConfig(
		middleware.CSRFConfig{
			TokenLookup: "form:_csrf",
		},
	))
	e.Use(middleware.RemoveTrailingSlash())
	e.Use(middleware.ContextTimeout(time.Second * time.Duration(config.ContextTimeoutSecond)))
	e.Use(session.Middleware(sessions.NewCookieStore([]byte(config.SessionSecret))))

	e.Debug = true
	e.Validator = &GoPlaygorundValidator{validator: validator.New()}

	staticRoot := e.Group("/static")
	staticRoot.Use(middleware.StaticWithConfig(middleware.StaticConfig{
		Root:   "./views/static",
		Browse: false,
	}))

	cycleService = services.NewCycleService(database, config.DefaultBudget, config.CycleStartDay)
	authService = services.NewAuthLocalService(database)
	routes.NewCycleRoutes(cycleService, authService, e)

	e.Logger.Fatal(e.Start(":8080"))
}
