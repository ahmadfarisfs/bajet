package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ahmadfarisfs/bajet/database"
	"github.com/ahmadfarisfs/bajet/models"
	"github.com/labstack/echo/v4"
)

func TestCalcBudgetsSumToTotal(t *testing.T) {
	for _, mode := range []models.DivisionMode{models.ModeEqual, models.ModeBehavioral, models.ModeMenurun, models.ModeProgresif} {
		for n := 1; n <= 12; n++ {
			var sum float64
			for _, b := range calcBudgets(1_000_000, n, mode) {
				sum += b
			}
			if sum != 1_000_000 {
				t.Errorf("%s n=%d: budgets sum to %v, want 1000000", mode, n, sum)
			}
		}
	}
}

func TestCalcDistributionNeverEmpty(t *testing.T) {
	for _, mode := range []models.DivisionMode{models.ModeEqual, models.ModeBehavioral} {
		for n := 1; n <= 12; n++ {
			for days := n; days <= 40; days++ {
				total := 0
				for i, d := range calcDistribution(days, n, mode) {
					if d < 1 {
						t.Fatalf("%s days=%d n=%d: period %d has %d days", mode, days, n, i+1, d)
					}
					total += d
				}
				if total != days {
					t.Fatalf("%s days=%d n=%d: distribution covers %d days", mode, days, n, total)
				}
			}
		}
	}
}

// ── HTTP flow ────────────────────────────────────────────────────────────────

func setupServer(t *testing.T) (*echo.Echo, string) {
	t.Helper()
	t.Setenv("DATABASE_URL", "")
	if err := database.Init(filepath.Join(t.TempDir(), "test.db")); err != nil {
		t.Fatal(err)
	}
	e := echo.New()
	api := e.Group("/api", AuthMiddleware)
	api.GET("/cycles", GetCycles)
	api.POST("/cycles", CreateCycle)
	api.GET("/cycles/:id", GetCycle)
	api.PUT("/cycles/:id", UpdateCycle)
	api.POST("/periods/:id/checkin", CheckIn)
	token, err := issueToken("user-1", "Test", "t@example.com", "")
	if err != nil {
		t.Fatal(err)
	}
	return e, token
}

func call(t *testing.T, e *echo.Echo, token, method, path, body string) (int, models.Cycle, string) {
	t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	var cycle models.Cycle
	json.Unmarshal(rec.Body.Bytes(), &cycle)
	return rec.Code, cycle, rec.Body.String()
}

func TestCycleLifecycle(t *testing.T) {
	e, token := setupServer(t)
	today := time.Now().UTC().Format("2006-01-02")
	end := time.Now().UTC().AddDate(0, 0, 29).Format("2006-01-02")

	// Too many periods for the range is rejected instead of producing broken periods.
	code, _, _ := call(t, e, token, "POST", "/api/cycles",
		`{"start_date":"2026-01-01","end_date":"2026-01-05","total_budget":100,"num_periods":12}`)
	if code != http.StatusBadRequest {
		t.Fatalf("short range: got %d, want 400", code)
	}

	code, created, body := call(t, e, token, "POST", "/api/cycles",
		`{"start_date":"`+today+`","end_date":"`+end+`","total_budget":1000000,"division_mode":"equal","num_periods":3}`)
	if code != http.StatusCreated || len(created.Periods) != 3 {
		t.Fatalf("create: %d %s", code, body)
	}
	id := jsonID(created.ID)

	// Future period can't be checked in; the current one can.
	code, _, _ = call(t, e, token, "POST", "/api/periods/"+jsonID(created.Periods[2].ID)+"/checkin",
		`{"result_type":"sisa","result_amount":1}`)
	if code != http.StatusBadRequest {
		t.Fatalf("future check-in: got %d, want 400", code)
	}
	code, _, body = call(t, e, token, "POST", "/api/periods/"+jsonID(created.Periods[0].ID)+"/checkin",
		`{"result_type":"defisit","result_amount":50000}`)
	if code != http.StatusOK {
		t.Fatalf("check-in: %d %s", code, body)
	}

	// After a check-in, reshaping is refused but a budget change keeps the result.
	code, _, _ = call(t, e, token, "PUT", "/api/cycles/"+id,
		`{"start_date":"`+today+`","end_date":"`+end+`","total_budget":1000000,"num_periods":4}`)
	if code != http.StatusConflict {
		t.Fatalf("reshape after check-in: got %d, want 409", code)
	}
	code, updated, body := call(t, e, token, "PUT", "/api/cycles/"+id,
		`{"start_date":"`+today+`","end_date":"`+end+`","total_budget":1500000,"division_mode":"menurun","num_periods":3}`)
	if code != http.StatusOK {
		t.Fatalf("budget update: %d %s", code, body)
	}
	var sum float64
	for _, p := range updated.Periods {
		sum += p.Budget
	}
	if sum != 1_500_000 || updated.Periods[0].Budget <= updated.Periods[2].Budget {
		t.Fatalf("budgets not redistributed front-loaded: %+v", updated.Periods)
	}

	code, got, _ := call(t, e, token, "GET", "/api/cycles/"+id, "")
	if code != http.StatusOK || got.Periods[0].Status != models.StatusCompleted || got.Periods[0].ResultAmount != 50000 {
		t.Fatalf("check-in lost after update: %+v", got.Periods[0])
	}
}

func jsonID(id uint) string {
	b, _ := json.Marshal(id)
	return string(b)
}
