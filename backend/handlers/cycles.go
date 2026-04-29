package handlers

import (
	"math"
	"net/http"
	"time"

	"github.com/ahmadfarisfs/bajet/database"
	"github.com/ahmadfarisfs/bajet/models"
	"github.com/labstack/echo/v4"
)

type CreateCycleRequest struct {
	StartDate    string              `json:"start_date"`
	EndDate      string              `json:"end_date"`
	TotalBudget  float64             `json:"total_budget"`
	DivisionMode models.DivisionMode `json:"division_mode"`
}

func GetCycles(c echo.Context) error {
	var cycles []models.Cycle
	database.DB.Order("created_at desc").Find(&cycles)

	for i := range cycles {
		database.DB.Where("cycle_id = ?", cycles[i].ID).Order("period_number asc").Find(&cycles[i].Periods)
	}
	return c.JSON(http.StatusOK, cycles)
}

func GetCycle(c echo.Context) error {
	id := c.Param("id")
	var cycle models.Cycle
	if err := database.DB.First(&cycle, id).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "cycle not found"})
	}
	database.DB.Where("cycle_id = ?", cycle.ID).Order("period_number asc").Find(&cycle.Periods)
	return c.JSON(http.StatusOK, cycle)
}

func CreateCycle(c echo.Context) error {
	req := new(CreateCycleRequest)
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid start_date, use YYYY-MM-DD"})
	}
	endDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid end_date, use YYYY-MM-DD"})
	}
	if !endDate.After(startDate) {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "end_date must be after start_date"})
	}

	if req.TotalBudget <= 0 {
		req.TotalBudget = 2000000
	}
	if req.DivisionMode == "" {
		req.DivisionMode = models.ModeEqual
	}

	cycle := models.Cycle{
		StartDate:    startDate,
		EndDate:      endDate,
		TotalBudget:  req.TotalBudget,
		DivisionMode: req.DivisionMode,
	}
	if err := database.DB.Create(&cycle).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	periods := generatePeriods(cycle)
	for i := range periods {
		database.DB.Create(&periods[i])
	}

	database.DB.Where("cycle_id = ?", cycle.ID).Order("period_number asc").Find(&cycle.Periods)
	return c.JSON(http.StatusCreated, cycle)
}

func DeleteCycle(c echo.Context) error {
	id := c.Param("id")
	var cycle models.Cycle
	if err := database.DB.First(&cycle, id).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "cycle not found"})
	}
	database.DB.Where("cycle_id = ?", cycle.ID).Delete(&models.Period{})
	database.DB.Delete(&cycle)
	return c.JSON(http.StatusOK, map[string]string{"message": "deleted"})
}

func generatePeriods(cycle models.Cycle) []models.Period {
	totalDays := int(cycle.EndDate.Sub(cycle.StartDate).Hours()/24) + 1
	dist := calcDistribution(totalDays, cycle.DivisionMode)
	budgetPerPeriod := math.Round(cycle.TotalBudget / 4)

	periods := make([]models.Period, 4)
	current := cycle.StartDate
	for i := 0; i < 4; i++ {
		pEnd := current.AddDate(0, 0, dist[i]-1)
		periods[i] = models.Period{
			CycleID:      cycle.ID,
			PeriodNumber: i + 1,
			StartDate:    current,
			EndDate:      pEnd,
			Budget:       budgetPerPeriod,
			Status:       models.StatusOpen,
		}
		current = pEnd.AddDate(0, 0, 1)
	}
	return periods
}

// calcDistribution splits totalDays into 4 periods.
//
// Equal mode: days distributed as evenly as possible, extra days front-loaded.
//   30d → [8,8,7,7]  31d → [8,8,8,7]
//
// Behavioral mode: front-loads one extra day to P1, creating a descending pattern.
//   30d → [9,7,7,7]  31d → [9,8,7,7]
func calcDistribution(totalDays int, mode models.DivisionMode) [4]int {
	base := totalDays / 4
	extra := totalDays % 4

	dist := [4]int{base, base, base, base}
	for i := 0; i < extra; i++ {
		dist[i]++
	}

	if mode == models.ModeBehavioral {
		if extra >= 2 {
			// Shift one day from the last enriched position to P1
			dist[0]++
			dist[extra-1]--
		} else {
			// Take from P4 to give P1 an extra day
			dist[0]++
			dist[3]--
		}
	}

	return dist
}
