package handlers

import (
	"net/http"
	"time"

	"github.com/ahmadfarisfs/bajet/database"
	"github.com/ahmadfarisfs/bajet/models"
	"github.com/labstack/echo/v4"
)

type CheckInRequest struct {
	ResultType   models.ResultType `json:"result_type"`
	ResultAmount float64           `json:"result_amount"`
}

func CheckIn(c echo.Context) error {
	id := c.Param("id")
	req := new(CheckInRequest)
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	var period models.Period
	if err := database.DB.
		Joins("JOIN cycles ON cycles.id = periods.cycle_id AND cycles.user_id = ?", userID(c)).
		First(&period, id).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "period not found"})
	}
	if period.Status == models.StatusCompleted {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "period already completed"})
	}
	// Dates are stored as UTC midnight and users sit anywhere from UTC-12 to
	// UTC+14, so a period has started for someone once now+14h reaches it.
	if period.StartDate.After(time.Now().UTC().Add(14 * time.Hour)) {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "period has not started yet"})
	}
	if req.ResultType != models.ResultSisa && req.ResultType != models.ResultDefisit {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "result_type must be 'sisa' or 'defisit'"})
	}
	if req.ResultAmount < 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "result_amount must be non-negative"})
	}

	period.ResultType = req.ResultType
	period.ResultAmount = req.ResultAmount
	period.Status = models.StatusCompleted

	database.DB.Save(&period)
	invalidateUser(userID(c))
	return c.JSON(http.StatusOK, period)
}

func UndoCheckIn(c echo.Context) error {
	id := c.Param("id")
	var period models.Period
	if err := database.DB.
		Joins("JOIN cycles ON cycles.id = periods.cycle_id AND cycles.user_id = ?", userID(c)).
		First(&period, id).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "period not found"})
	}

	period.Status = models.StatusOpen
	period.ResultType = ""
	period.ResultAmount = 0
	database.DB.Save(&period)
	invalidateUser(userID(c))
	return c.JSON(http.StatusOK, period)
}
