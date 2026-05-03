package main

import (
	"log"
	"net/http"
	"os"

	"github.com/ahmadfarisfs/bajet/database"
	"github.com/ahmadfarisfs/bajet/handlers"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "bajet.db"
	}
	if err := database.Init(dbPath); err != nil {
		log.Fatalf("failed to init database: %v", err)
	}

	e := echo.New()
	e.HideBanner = true

	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "${method} ${uri} → ${status}\n",
	}))
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodDelete},
		AllowHeaders: []string{"Content-Type", "Authorization"},
	}))
	e.Use(middleware.Recover())

	// Waitlist — no auth required
	e.POST("/api/waitlist", handlers.JoinWaitlist)
	e.GET("/api/waitlist/count", handlers.WaitlistCount)
	e.GET("/admin/waitlist-entries", handlers.AdminWaitlist)

	api := e.Group("/api")
	api.Use(handlers.AuthMiddleware)
	api.GET("/cycles", handlers.GetCycles)
	api.POST("/cycles", handlers.CreateCycle)
	api.GET("/cycles/:id", handlers.GetCycle)
	api.DELETE("/cycles/:id", handlers.DeleteCycle)
	api.POST("/periods/:id/checkin", handlers.CheckIn)
	api.DELETE("/periods/:id/checkin", handlers.UndoCheckIn)

	// Root serves the landing page; everything else falls through to the Svelte SPA
	e.GET("/", func(c echo.Context) error {
		return c.File("../frontend/dist/landing.html")
	})
	e.Static("/", "../frontend/dist")
	e.File("/*", "../frontend/dist/index.html")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	e.Logger.Fatal(e.Start(":" + port))
}
