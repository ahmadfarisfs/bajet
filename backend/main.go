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
	}))
	e.Use(middleware.Recover())

	api := e.Group("/api")
	api.GET("/cycles", handlers.GetCycles)
	api.POST("/cycles", handlers.CreateCycle)
	api.GET("/cycles/:id", handlers.GetCycle)
	api.DELETE("/cycles/:id", handlers.DeleteCycle)
	api.POST("/periods/:id/checkin", handlers.CheckIn)
	api.DELETE("/periods/:id/checkin", handlers.UndoCheckIn)

	// Serve frontend static files in production
	e.Static("/", "../frontend/dist")
	e.File("/*", "../frontend/dist/index.html")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	e.Logger.Fatal(e.Start(":" + port))
}
