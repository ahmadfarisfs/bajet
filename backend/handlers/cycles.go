package handlers

import (
	"math"
	"net/http"
	"time"

	"github.com/ahmadfarisfs/bajet/database"
	"github.com/ahmadfarisfs/bajet/models"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

type CycleRequest struct {
	StartDate    string              `json:"start_date"`
	EndDate      string              `json:"end_date"`
	TotalBudget  float64             `json:"total_budget"`
	DivisionMode models.DivisionMode `json:"division_mode"`
	NumPeriods   int                 `json:"num_periods"`
}

// validCycle is a CycleRequest after parsing and defaulting.
type validCycle struct {
	start, end time.Time
	budget     float64
	mode       models.DivisionMode
	n          int
}

func parseCycleRequest(req *CycleRequest) (validCycle, string) {
	start, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		return validCycle{}, "invalid start_date, use YYYY-MM-DD"
	}
	end, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		return validCycle{}, "invalid end_date, use YYYY-MM-DD"
	}
	if !end.After(start) {
		return validCycle{}, "end_date must be after start_date"
	}
	v := validCycle{start: start, end: end, budget: req.TotalBudget, mode: req.DivisionMode, n: req.NumPeriods}
	if v.budget <= 0 {
		return validCycle{}, "total_budget must be greater than 0"
	}
	switch v.mode {
	case "":
		v.mode = models.ModeEqual
	case models.ModeEqual, models.ModeBehavioral, models.ModeMenurun, models.ModeProgresif:
	default:
		return validCycle{}, "invalid division_mode"
	}
	if v.n < 1 || v.n > 12 {
		v.n = 4
	}
	if daysInclusive(start, end) < v.n {
		return validCycle{}, "date range is shorter than the number of periods"
	}
	return v, ""
}

func daysInclusive(start, end time.Time) int {
	return int(end.Sub(start).Hours()/24) + 1
}

func periodsByNumber(db *gorm.DB) *gorm.DB {
	return db.Order("period_number asc")
}

func GetCycles(c echo.Context) error {
	uid := userID(c)
	if cached, ok := cacheGetCycles(uid); ok {
		return c.JSON(http.StatusOK, cached)
	}

	var cycles []models.Cycle
	database.DB.Preload("Periods", periodsByNumber).
		Where("user_id = ?", uid).Order("created_at desc").Find(&cycles)
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
	if err := database.DB.Preload("Periods", periodsByNumber).
		Where("id = ? AND user_id = ?", id, uid).First(&cycle).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "cycle not found"})
	}
	cacheSetCycle(uid, id, cycle)
	return c.JSON(http.StatusOK, cycle)
}

func CreateCycle(c echo.Context) error {
	req := new(CycleRequest)
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	v, msg := parseCycleRequest(req)
	if msg != "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": msg})
	}

	cycle := models.Cycle{
		UserID:       userID(c),
		StartDate:    v.start,
		EndDate:      v.end,
		TotalBudget:  v.budget,
		DivisionMode: v.mode,
		NumPeriods:   v.n,
	}
	err := database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&cycle).Error; err != nil {
			return err
		}
		periods := generatePeriods(cycle)
		if err := tx.Create(&periods).Error; err != nil {
			return err
		}
		cycle.Periods = periods
		return nil
	})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	invalidateUser(userID(c))
	return c.JSON(http.StatusCreated, cycle)
}

// UpdateCycle edits a cycle. Budget and division mode can always change: the
// period budgets are recalculated and check-in results are kept. Dates and the
// number of periods can only change while no period has been checked in,
// because the periods are rebuilt from scratch.
func UpdateCycle(c echo.Context) error {
	id := c.Param("id")
	uid := userID(c)
	req := new(CycleRequest)
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	v, msg := parseCycleRequest(req)
	if msg != "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": msg})
	}

	var cycle models.Cycle
	if err := database.DB.Preload("Periods", periodsByNumber).
		Where("id = ? AND user_id = ?", id, uid).First(&cycle).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "cycle not found"})
	}

	hasCheckIns := false
	for _, p := range cycle.Periods {
		if p.Status == models.StatusCompleted {
			hasCheckIns = true
			break
		}
	}
	reshape := !v.start.Equal(cycle.StartDate) || !v.end.Equal(cycle.EndDate) || v.n != len(cycle.Periods)
	if reshape && hasCheckIns {
		return c.JSON(http.StatusConflict, map[string]string{
			"error": "dates and number of periods can't change after a check-in; undo the check-ins first",
		})
	}

	cycle.StartDate, cycle.EndDate = v.start, v.end
	cycle.TotalBudget, cycle.DivisionMode, cycle.NumPeriods = v.budget, v.mode, v.n

	err := database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Omit("Periods").Save(&cycle).Error; err != nil {
			return err
		}
		if !hasCheckIns {
			// Nothing to preserve: rebuild periods so dates follow the new mode too.
			if err := tx.Where("cycle_id = ?", cycle.ID).Delete(&models.Period{}).Error; err != nil {
				return err
			}
			periods := generatePeriods(cycle)
			if err := tx.Create(&periods).Error; err != nil {
				return err
			}
			cycle.Periods = periods
			return nil
		}
		budgets := calcBudgets(cycle.TotalBudget, len(cycle.Periods), cycle.DivisionMode)
		for i := range cycle.Periods {
			cycle.Periods[i].Budget = budgets[i]
			if err := tx.Model(&cycle.Periods[i]).Update("budget", budgets[i]).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	invalidateUser(uid)
	return c.JSON(http.StatusOK, cycle)
}

func DeleteCycle(c echo.Context) error {
	id := c.Param("id")
	var cycle models.Cycle
	if err := database.DB.Where("id = ? AND user_id = ?", id, userID(c)).First(&cycle).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "cycle not found"})
	}
	err := database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("cycle_id = ?", cycle.ID).Delete(&models.Period{}).Error; err != nil {
			return err
		}
		return tx.Delete(&cycle).Error
	})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	invalidateUser(userID(c))
	return c.JSON(http.StatusOK, map[string]string{"message": "deleted"})
}

func generatePeriods(cycle models.Cycle) []models.Period {
	n := cycle.NumPeriods
	if n < 1 {
		n = 4
	}
	dist := calcDistribution(daysInclusive(cycle.StartDate, cycle.EndDate), n, cycle.DivisionMode)
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
// Every mode gives the rounding remainder to the last period so budgets sum to the total.
func calcBudgets(totalBudget float64, n int, mode models.DivisionMode) []float64 {
	weight := func(i int) float64 { return 1 }
	totalWeight := float64(n)
	switch mode {
	case models.ModeProgresif:
		weight = func(i int) float64 { return float64(i + 1) }
		totalWeight = float64(n * (n + 1) / 2)
	case models.ModeMenurun:
		weight = func(i int) float64 { return float64(n - i) }
		totalWeight = float64(n * (n + 1) / 2)
	}

	budgets := make([]float64, n)
	var sum float64
	for i := 0; i < n-1; i++ {
		budgets[i] = math.Floor(totalBudget * weight(i) / totalWeight)
		sum += budgets[i]
	}
	budgets[n-1] = totalBudget - sum
	return budgets
}

// calcDistribution splits totalDays into n periods.
//
// Equal/Menurun/Progresif: days distributed as evenly as possible, extra days front-loaded.
// Behavioral: same as Equal but P1 gets one extra day stolen from the last enriched period.
// The caller guarantees totalDays >= n; no period is ever left with zero days.
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

	if mode == models.ModeBehavioral && n > 1 {
		donor := n - 1
		if extra >= 2 {
			donor = extra - 1
		}
		if dist[donor] > 1 {
			dist[0]++
			dist[donor]--
		}
	}

	return dist
}
