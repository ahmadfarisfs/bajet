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
	NumPeriods   int                 `json:"num_periods"`
}

func GetCycles(c echo.Context) error {
	uid := userID(c)
	if cached, ok := cacheGetCycles(uid); ok {
		return c.JSON(http.StatusOK, cached)
	}

	var cycles []models.Cycle
	database.DB.Where("user_id = ?", uid).Order("created_at desc").Find(&cycles)
	for i := range cycles {
		database.DB.Where("cycle_id = ?", cycles[i].ID).Order("period_number asc").Find(&cycles[i].Periods)
	}
	cacheSetCycles(uid, cycles)
	return c.JSON(http.StatusOK, cycles)
}

func GetCycle(c echo.Context) error {
	id := c.Param("id")
	uid := userID(c)
	if cached, ok := cacheGetCycle(uid, id); ok {
		return c.JSON(http.StatusOK, cached)
	}

	var cycle models.Cycle
	if err := database.DB.Where("id = ? AND user_id = ?", id, uid).First(&cycle).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "cycle not found"})
	}
	database.DB.Where("cycle_id = ?", cycle.ID).Order("period_number asc").Find(&cycle.Periods)
	cacheSetCycle(uid, id, cycle)
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
	if req.NumPeriods < 1 || req.NumPeriods > 12 {
		req.NumPeriods = 4
	}

	cycle := models.Cycle{
		UserID:       userID(c),
		StartDate:    startDate,
		EndDate:      endDate,
		TotalBudget:  req.TotalBudget,
		DivisionMode: req.DivisionMode,
		NumPeriods:   req.NumPeriods,
	}
	if err := database.DB.Create(&cycle).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	periods := generatePeriods(cycle)
	for i := range periods {
		database.DB.Create(&periods[i])
	}

	database.DB.Where("cycle_id = ?", cycle.ID).Order("period_number asc").Find(&cycle.Periods)
	invalidateUser(userID(c))
	return c.JSON(http.StatusCreated, cycle)
}

func DeleteCycle(c echo.Context) error {
	id := c.Param("id")
	var cycle models.Cycle
	if err := database.DB.Where("id = ? AND user_id = ?", id, userID(c)).First(&cycle).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "cycle not found"})
	}
	database.DB.Where("cycle_id = ?", cycle.ID).Delete(&models.Period{})
	database.DB.Delete(&cycle)
	invalidateUser(userID(c))
	return c.JSON(http.StatusOK, map[string]string{"message": "deleted"})
}

func generatePeriods(cycle models.Cycle) []models.Period {
	n := cycle.NumPeriods
	if n < 1 {
		n = 4
	}
	totalDays := int(cycle.EndDate.Sub(cycle.StartDate).Hours()/24) + 1
	dist := calcDistribution(totalDays, n, cycle.DivisionMode)
	budgets := calcBudgets(cycle.TotalBudget, n, cycle.DivisionMode)

	periods := make([]models.Period, n)
	current := cycle.StartDate
	for i := 0; i < n; i++ {
		pEnd := current.AddDate(0, 0, dist[i]-1)
		periods[i] = models.Period{
			CycleID:      cycle.ID,
			PeriodNumber: i + 1,
			StartDate:    current,
			EndDate:      pEnd,
			Budget:       budgets[i],
			Status:       models.StatusOpen,
		}
		current = pEnd.AddDate(0, 0, 1)
	}
	return periods
}

// calcBudgets returns the budget for each period based on division mode.
// Progresif: back-loaded (last period gets most), Menurun: front-loaded (first gets most).
// Equal/Behavioral: equal budget per period.
func calcBudgets(totalBudget float64, n int, mode models.DivisionMode) []float64 {
	budgets := make([]float64, n)
	switch mode {
	case models.ModeProgresif:
		// Back-loaded: weight[i] = i+1
		totalWeight := n * (n + 1) / 2
		var sum float64
		for i := 0; i < n-1; i++ {
			b := math.Floor(totalBudget * float64(i+1) / float64(totalWeight))
			budgets[i] = b
			sum += b
		}
		budgets[n-1] = totalBudget - sum
	case models.ModeMenurun:
		// Front-loaded: weight[i] = n-i
		totalWeight := n * (n + 1) / 2
		var sum float64
		for i := 0; i < n-1; i++ {
			b := math.Floor(totalBudget * float64(n-i) / float64(totalWeight))
			budgets[i] = b
			sum += b
		}
		budgets[n-1] = totalBudget - sum
	default:
		// equal / behavioral: equal budget per period
		base := math.Round(totalBudget / float64(n))
		for i := range budgets {
			budgets[i] = base
		}
	}
	return budgets
}

// calcDistribution splits totalDays into n periods.
//
// Equal/Menurun/Progresif: days distributed as evenly as possible, extra days front-loaded.
// Behavioral: same as Equal but P1 gets one extra day stolen from the last enriched period.
func calcDistribution(totalDays int, n int, mode models.DivisionMode) []int {
	base := totalDays / n
	extra := totalDays % n

	dist := make([]int, n)
	for i := range dist {
		dist[i] = base
	}
	for i := 0; i < extra; i++ {
		dist[i]++
	}

	if mode == models.ModeBehavioral {
		if extra >= 2 {
			dist[0]++
			dist[extra-1]--
		} else {
			dist[0]++
			dist[n-1]--
		}
	}

	return dist
}
