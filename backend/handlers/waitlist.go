package handlers

import (
	"net/http"
	"strings"

	"github.com/ahmadfarisfs/bajet/database"
	"github.com/ahmadfarisfs/bajet/models"
	"github.com/labstack/echo/v4"
)

func JoinWaitlist(c echo.Context) error {
	var body struct {
		Email  string `json:"email"`
		Source string `json:"source"`
	}
	if err := c.Bind(&body); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"ok": false, "error": "Invalid request"})
	}

	email := strings.TrimSpace(strings.ToLower(body.Email))
	if email == "" || !strings.Contains(email, "@") {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"ok": false, "error": "Invalid email"})
	}

	source := strings.TrimSpace(body.Source)
	if source == "" {
		source = "landing"
	}

	entry := models.Waitlist{
		Email:  email,
		Source: source,
		IP:     c.RealIP(),
	}

	if err := database.DB.Create(&entry).Error; err != nil {
		// Duplicate — silently succeed so double-submits don't confuse users
		if strings.Contains(err.Error(), "UNIQUE") || strings.Contains(err.Error(), "unique") || strings.Contains(err.Error(), "duplicate") {
			return c.JSON(http.StatusOK, map[string]interface{}{"ok": true, "message": "You're on the list!"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{"ok": false, "error": "Server error"})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{"ok": true, "message": "You're on the list!"})
}

func WaitlistCount(c echo.Context) error {
	var count int64
	database.DB.Model(&models.Waitlist{}).Count(&count)
	return c.JSON(http.StatusOK, map[string]interface{}{"count": count})
}

func AdminWaitlist(c echo.Context) error {
	var entries []models.Waitlist
	database.DB.Order("created_at desc").Find(&entries)
	return c.JSON(http.StatusOK, entries)
}
