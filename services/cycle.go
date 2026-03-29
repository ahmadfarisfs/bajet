package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"sort"
	"time"

	"bajetapp/model"
)

const dateLayout = "2006-01-02"

var ErrNoCycle = errors.New("no cycle found")
var ErrCycleOverlap = errors.New("new cycle cannot overlap with an existing cycle")

// CycleService handles the simplified cycle -> period -> result flow.
type CycleService struct {
	db            *sql.DB
	defaultBudget int
	cycleStartDay int
}

func NewCycleService(db *sql.DB, defaultBudget int, cycleStartDay int) *CycleService {
	return &CycleService{
		db:            db,
		defaultBudget: defaultBudget,
		cycleStartDay: cycleStartDay,
	}
}

func (cs *CycleService) GetDashboardByUser(ctx context.Context, userID int64, now time.Time) (model.Dashboard, error) {
	cycle, err := cs.findLatestUnfinishedCycle(ctx, userID)
	if err == ErrNoCycle {
		cycle, err = cs.findCurrentOrLatestCycle(ctx, userID, now)
	}
	if err != nil {
		return model.Dashboard{}, err
	}

	currencyCode, language, err := cs.GetUserPreferences(ctx, userID)
	if err != nil {
		return model.Dashboard{}, err
	}

	periods, err := cs.listPeriodsByCycleID(ctx, cycle.ID)
	if err != nil {
		return model.Dashboard{}, err
	}

	dashboard := model.Dashboard{
		Cycle:          cycle,
		Periods:        periods,
		VisiblePeriods: buildVisiblePeriods(periods),
		PeriodTiming:   resolvePeriodTiming(periods, now),
		CurrencyCode:   currencyCode,
		Language:       language,
	}

	for _, period := range periods {
		if period.Status != model.PeriodStatusCompleted {
			continue
		}
		dashboard.TotalSavings += period.Savings
		dashboard.TotalDefisit += period.Overspending
		dashboard.TotalSpending += period.Budget - period.Savings + period.Overspending
	}
	dashboard.NetResult = dashboard.TotalSavings - dashboard.TotalDefisit

	allCompleted := len(periods) > 0
	for _, period := range periods {
		if period.Status != model.PeriodStatusCompleted {
			allCompleted = false
			break
		}
	}
	dashboard.CanStartNext = allCompleted

	history, err := cs.listCycleHistoryByUser(ctx, userID, 12)
	if err != nil {
		return model.Dashboard{}, err
	}
	dashboard.History = history

	overallSavings, overallDefisit, overallSpending, err := cs.getOverallImpactByUser(ctx, userID)
	if err != nil {
		return model.Dashboard{}, err
	}
	dashboard.OverallSavings = overallSavings
	dashboard.OverallDefisit = overallDefisit
	dashboard.OverallSpending = overallSpending
	dashboard.OverallNetResult = overallSavings - overallDefisit

	return dashboard, nil
}

func (cs *CycleService) getOverallImpactByUser(ctx context.Context, userID int64) (int, int, int, error) {
	var savings int
	var defisit int
	var spending int
	err := cs.db.QueryRowContext(
		ctx,
		`SELECT
			COALESCE(SUM(p.savings), 0) AS total_savings,
			COALESCE(SUM(p.overspending), 0) AS total_defisit,
			COALESCE(SUM(p.budget - p.savings + p.overspending), 0) AS total_spending
		 FROM periods p
		 JOIN cycles c ON c.id = p.cycle_id
		 WHERE c.user_id = ?
		   AND p.status = ?`,
		userID,
		model.PeriodStatusCompleted,
	).Scan(&savings, &defisit, &spending)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("query overall impact: %w", err)
	}
	return savings, defisit, spending, nil
}

func (cs *CycleService) GetUserPreferences(ctx context.Context, userID int64) (string, string, error) {
	var currencyCode string
	var language string
	err := cs.db.QueryRowContext(
		ctx,
		`SELECT COALESCE(currency_code, 'USD'), COALESCE(language, 'en') FROM users WHERE id = ? LIMIT 1`,
		userID,
	).Scan(&currencyCode, &language)
	if err != nil {
		if err == sql.ErrNoRows {
			return "USD", "en", nil
		}
		return "", "", fmt.Errorf("query user preferences: %w", err)
	}
	return currencyCode, language, nil
}

func (cs *CycleService) UpdateUserPreferences(ctx context.Context, userID int64, currencyCode, language string) error {
	res, err := cs.db.ExecContext(
		ctx,
		`UPDATE users SET currency_code = ?, language = ? WHERE id = ?`,
		currencyCode,
		language,
		userID,
	)
	if err != nil {
		return fmt.Errorf("update user preferences: %w", err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("update user preferences rows: %w", err)
	}
	if affected == 0 {
		return fmt.Errorf("user not found")
	}
	return nil
}

func buildVisiblePeriods(periods []model.Period) []model.Period {
	if len(periods) == 0 {
		return periods
	}

	firstOpen := -1
	for i, period := range periods {
		if period.Status == model.PeriodStatusOpen {
			firstOpen = i
			break
		}
	}

	if firstOpen == -1 {
		reversed := make([]model.Period, len(periods))
		for i, p := range periods {
			reversed[len(periods)-1-i] = p
		}
		return reversed
	}

	visible := make([]model.Period, 0, firstOpen+1)
	visible = append(visible, periods[firstOpen])
	for i := firstOpen - 1; i >= 0; i-- {
		visible = append(visible, periods[i])
	}

	return visible
}

func resolvePeriodTiming(periods []model.Period, now time.Time) string {
	if len(periods) == 0 {
		return model.PeriodTimingInactive
	}

	var active *model.Period
	for i := range periods {
		if periods[i].Status == model.PeriodStatusOpen {
			active = &periods[i]
			break
		}
	}

	if active == nil {
		return model.PeriodTimingInactive
	}

	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	start := time.Date(active.StartDate.Year(), active.StartDate.Month(), active.StartDate.Day(), 0, 0, 0, 0, active.StartDate.Location())
	end := time.Date(active.EndDate.Year(), active.EndDate.Month(), active.EndDate.Day(), 0, 0, 0, 0, active.EndDate.Location())

	if today.Before(start) {
		return model.PeriodTimingEarly
	}
	if today.After(end) {
		return model.PeriodTimingLate
	}
	return model.PeriodTimingOnTrack
}

func (cs *CycleService) findLatestUnfinishedCycle(ctx context.Context, userID int64) (model.Cycle, error) {
	var cycle model.Cycle
	var startDate string
	var endDate string
	var createdAt string

	err := cs.db.QueryRowContext(
		ctx,
		`SELECT c.id, c.start_date, c.end_date, c.total_budget, c.period_count, c.created_at
		 FROM cycles c
		 WHERE c.user_id = ?
		   AND EXISTS (
			   SELECT 1
			   FROM periods p
			   WHERE p.cycle_id = c.id AND p.status != ?
		   )
		 ORDER BY c.start_date DESC, c.id DESC
		 LIMIT 1`,
		userID,
		model.PeriodStatusCompleted,
	).Scan(&cycle.ID, &startDate, &endDate, &cycle.TotalBudget, &cycle.PeriodCount, &createdAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return model.Cycle{}, ErrNoCycle
		}
		return model.Cycle{}, fmt.Errorf("find latest unfinished cycle: %w", err)
	}

	cycle.StartDate, _ = time.ParseInLocation(dateLayout, startDate, time.Local)
	cycle.EndDate, _ = time.ParseInLocation(dateLayout, endDate, time.Local)
	cycle.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	return cycle, nil
}

func (cs *CycleService) CompletePeriodByUser(ctx context.Context, userID, periodID int64, inputType string, amount int) error {
	if amount < 0 {
		return fmt.Errorf("amount must be >= 0")
	}
	if inputType != model.InputTypeSisa && inputType != model.InputTypeDefisit {
		return fmt.Errorf("invalid input type")
	}

	savings := 0
	over := 0
	if inputType == model.InputTypeSisa {
		savings = amount
	} else {
		over = amount
	}

	res, err := cs.db.ExecContext(
		ctx,
		`UPDATE periods
		 SET status = ?, input_type = ?, input_amount = ?, savings = ?, overspending = ?, completed_at = ?
		 WHERE id = ? AND status = ? AND cycle_id IN (SELECT id FROM cycles WHERE user_id = ?)`,
		model.PeriodStatusCompleted,
		inputType,
		amount,
		savings,
		over,
		time.Now().Format(time.RFC3339),
		periodID,
		model.PeriodStatusOpen,
		userID,
	)
	if err != nil {
		return fmt.Errorf("complete period: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("complete period rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("period already completed or not found")
	}

	return nil
}

func (cs *CycleService) HasAnyCycle(ctx context.Context, userID int64) (bool, error) {
	var count int
	err := cs.db.QueryRowContext(ctx, `SELECT COUNT(1) FROM cycles WHERE user_id = ?`, userID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("count cycles: %w", err)
	}
	return count > 0, nil
}

func (cs *CycleService) GetSuggestedPlan(ctx context.Context, userID int64, startDate, endDate string, totalBudget int, periodCount int, adjustmentMode string) ([]model.PeriodPlan, error) {
	if periodCount < 2 || periodCount > 8 {
		return nil, fmt.Errorf("period count must be between 2 and 8")
	}
	if totalBudget <= 0 {
		totalBudget = cs.defaultBudget
	}

	start, err := time.ParseInLocation(dateLayout, startDate, time.Local)
	if err != nil {
		return nil, fmt.Errorf("invalid start date")
	}
	end, err := time.ParseInLocation(dateLayout, endDate, time.Local)
	if err != nil {
		return nil, fmt.Errorf("invalid end date")
	}
	if end.Before(start) {
		return nil, fmt.Errorf("end date must be after start date")
	}

	aiWeights := []float64(nil)
	if normalizeAdjustmentMode(adjustmentMode) == model.AdjustmentAIMaxSave {
		weights, err := cs.getAIMaxSaveWeights(ctx, userID, periodCount)
		if err != nil {
			return nil, err
		}
		aiWeights = weights
	}

	return cs.buildPeriodPlans(start, end, totalBudget, periodCount, adjustmentMode, aiWeights), nil
}

func (cs *CycleService) CreateCycleForUser(ctx context.Context, userID int64, startDate, endDate string, totalBudget, periodCount int, adjustmentMode string) (model.Cycle, error) {
	if periodCount < 2 || periodCount > 8 {
		return model.Cycle{}, fmt.Errorf("period count must be between 2 and 8")
	}
	if totalBudget <= 0 {
		return model.Cycle{}, fmt.Errorf("budget must be greater than 0")
	}

	start, err := time.ParseInLocation(dateLayout, startDate, time.Local)
	if err != nil {
		return model.Cycle{}, fmt.Errorf("invalid start date")
	}
	end, err := time.ParseInLocation(dateLayout, endDate, time.Local)
	if err != nil {
		return model.Cycle{}, fmt.Errorf("invalid end date")
	}
	if end.Before(start) {
		return model.Cycle{}, fmt.Errorf("end date must be after start date")
	}

	if existing, err := cs.findCycleByWindow(ctx, userID, start, end); err == nil {
		return existing, nil
	} else if err != sql.ErrNoRows {
		return model.Cycle{}, err
	}

	overlap, err := cs.hasOverlappingCycle(ctx, userID, start, end)
	if err != nil {
		return model.Cycle{}, err
	}
	if overlap {
		return model.Cycle{}, ErrCycleOverlap
	}

	aiWeights := []float64(nil)
	if normalizeAdjustmentMode(adjustmentMode) == model.AdjustmentAIMaxSave {
		weights, err := cs.getAIMaxSaveWeights(ctx, userID, periodCount)
		if err != nil {
			return model.Cycle{}, err
		}
		aiWeights = weights
	}

	plans := cs.buildPeriodPlans(start, end, totalBudget, periodCount, adjustmentMode, aiWeights)

	tx, err := cs.db.BeginTx(ctx, nil)
	if err != nil {
		return model.Cycle{}, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	createdAt := time.Now().Format(time.RFC3339)
	res, err := tx.ExecContext(
		ctx,
		`INSERT INTO cycles(user_id, start_date, end_date, total_budget, period_count, created_at)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		userID,
		start.Format(dateLayout),
		end.Format(dateLayout),
		totalBudget,
		periodCount,
		createdAt,
	)
	if err != nil {
		return model.Cycle{}, fmt.Errorf("insert cycle: %w", err)
	}

	cycleID, err := res.LastInsertId()
	if err != nil {
		return model.Cycle{}, fmt.Errorf("insert cycle id: %w", err)
	}

	for _, plan := range plans {
		_, err = tx.ExecContext(
			ctx,
			`INSERT INTO periods(
				cycle_id, period_number, start_date, end_date, budget, status, created_at
			) VALUES (?, ?, ?, ?, ?, ?, ?)`,
			cycleID,
			plan.PeriodNumber,
			plan.StartDate,
			plan.EndDate,
			plan.Budget,
			model.PeriodStatusOpen,
			createdAt,
		)
		if err != nil {
			return model.Cycle{}, fmt.Errorf("insert period %d: %w", plan.PeriodNumber, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return model.Cycle{}, fmt.Errorf("commit cycle transaction: %w", err)
	}

	return model.Cycle{
		ID:          cycleID,
		StartDate:   start,
		EndDate:     end,
		TotalBudget: totalBudget,
		PeriodCount: periodCount,
		CreatedAt:   time.Now(),
	}, nil
}

func (cs *CycleService) SuggestDefaultWindow(now time.Time) (time.Time, time.Time) {
	loc := now.Location()
	y, m, _ := now.Date()
	start := time.Date(y, m, cs.cycleStartDay, 0, 0, 0, 0, loc)
	if now.Before(start) {
		start = start.AddDate(0, -1, 0)
	}
	end := start.AddDate(0, 1, -1)
	return start, end
}

func (cs *CycleService) SuggestNextWindowByUser(ctx context.Context, userID int64, now time.Time) (time.Time, time.Time, error) {
	latest, err := cs.findLatestCycleByUser(ctx, userID)
	if err != nil {
		if err == ErrNoCycle {
			start, end := cs.SuggestDefaultWindow(now)
			return start, end, nil
		}
		return time.Time{}, time.Time{}, err
	}

	days := int(latest.EndDate.Sub(latest.StartDate).Hours()/24) + 1
	if days < 1 {
		days = 30
	}
	start := latest.EndDate.AddDate(0, 0, 1)
	end := start.AddDate(0, 0, days-1)
	return start, end, nil
}

func (cs *CycleService) GetLatestCycleByUser(ctx context.Context, userID int64) (model.Cycle, error) {
	return cs.findLatestCycleByUser(ctx, userID)
}

func (cs *CycleService) findCurrentOrLatestCycle(ctx context.Context, userID int64, now time.Time) (model.Cycle, error) {
	var cycle model.Cycle
	var startDate string
	var endDate string
	var createdAt string

	err := cs.db.QueryRowContext(
		ctx,
		`SELECT id, start_date, end_date, total_budget, period_count, created_at
		 FROM cycles
		 WHERE user_id = ? AND start_date <= ? AND end_date >= ?
		 ORDER BY id DESC LIMIT 1`,
		userID,
		now.Format(dateLayout),
		now.Format(dateLayout),
	).Scan(&cycle.ID, &startDate, &endDate, &cycle.TotalBudget, &cycle.PeriodCount, &createdAt)

	if err == sql.ErrNoRows {
		err = cs.db.QueryRowContext(
			ctx,
			`SELECT id, start_date, end_date, total_budget, period_count, created_at
			 FROM cycles WHERE user_id = ? ORDER BY id DESC LIMIT 1`,
			userID,
		).Scan(&cycle.ID, &startDate, &endDate, &cycle.TotalBudget, &cycle.PeriodCount, &createdAt)
	}
	if err != nil {
		if err == sql.ErrNoRows {
			return model.Cycle{}, ErrNoCycle
		}
		return model.Cycle{}, fmt.Errorf("find cycle: %w", err)
	}

	cycle.StartDate, _ = time.ParseInLocation(dateLayout, startDate, time.Local)
	cycle.EndDate, _ = time.ParseInLocation(dateLayout, endDate, time.Local)
	cycle.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	return cycle, nil
}

func (cs *CycleService) findCycleByWindow(ctx context.Context, userID int64, start, end time.Time) (model.Cycle, error) {
	var cycle model.Cycle
	var startDate string
	var endDate string
	var createdAt string

	err := cs.db.QueryRowContext(
		ctx,
		`SELECT id, start_date, end_date, total_budget, period_count, created_at
		 FROM cycles WHERE user_id = ? AND start_date = ? AND end_date = ? LIMIT 1`,
		userID,
		start.Format(dateLayout),
		end.Format(dateLayout),
	).Scan(&cycle.ID, &startDate, &endDate, &cycle.TotalBudget, &cycle.PeriodCount, &createdAt)
	if err != nil {
		return model.Cycle{}, err
	}

	cycle.StartDate, _ = time.ParseInLocation(dateLayout, startDate, time.Local)
	cycle.EndDate, _ = time.ParseInLocation(dateLayout, endDate, time.Local)
	cycle.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	return cycle, nil
}

func (cs *CycleService) findLatestCycleByUser(ctx context.Context, userID int64) (model.Cycle, error) {
	var cycle model.Cycle
	var startDate string
	var endDate string
	var createdAt string

	err := cs.db.QueryRowContext(
		ctx,
		`SELECT id, start_date, end_date, total_budget, period_count, created_at
		 FROM cycles WHERE user_id = ? ORDER BY end_date DESC, id DESC LIMIT 1`,
		userID,
	).Scan(&cycle.ID, &startDate, &endDate, &cycle.TotalBudget, &cycle.PeriodCount, &createdAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return model.Cycle{}, ErrNoCycle
		}
		return model.Cycle{}, fmt.Errorf("find latest cycle: %w", err)
	}

	cycle.StartDate, _ = time.ParseInLocation(dateLayout, startDate, time.Local)
	cycle.EndDate, _ = time.ParseInLocation(dateLayout, endDate, time.Local)
	cycle.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	return cycle, nil
}

func (cs *CycleService) hasOverlappingCycle(ctx context.Context, userID int64, start, end time.Time) (bool, error) {
	var count int
	err := cs.db.QueryRowContext(
		ctx,
		`SELECT COUNT(1)
		 FROM cycles
		 WHERE user_id = ?
		   AND NOT (end_date < ? OR start_date > ?)`,
		userID,
		start.Format(dateLayout),
		end.Format(dateLayout),
	).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("check overlapping cycle: %w", err)
	}
	return count > 0, nil
}

func (cs *CycleService) listPeriodsByCycleID(ctx context.Context, cycleID int64) ([]model.Period, error) {
	rows, err := cs.db.QueryContext(
		ctx,
		`SELECT id, cycle_id, period_number, start_date, end_date, budget, status,
		        COALESCE(input_type, ''), input_amount, savings, overspending, COALESCE(completed_at, '')
		 FROM periods
		 WHERE cycle_id = ?
		 ORDER BY period_number ASC`,
		cycleID,
	)
	if err != nil {
		return nil, fmt.Errorf("query periods: %w", err)
	}
	defer rows.Close()

	periods := make([]model.Period, 0, 4)
	for rows.Next() {
		var period model.Period
		var startDate string
		var endDate string
		var completedAt string

		if err := rows.Scan(
			&period.ID,
			&period.CycleID,
			&period.PeriodNumber,
			&startDate,
			&endDate,
			&period.Budget,
			&period.Status,
			&period.InputType,
			&period.InputAmount,
			&period.Savings,
			&period.Overspending,
			&completedAt,
		); err != nil {
			return nil, fmt.Errorf("scan period: %w", err)
		}

		period.StartDate, _ = time.ParseInLocation(dateLayout, startDate, time.Local)
		period.EndDate, _ = time.ParseInLocation(dateLayout, endDate, time.Local)
		if completedAt != "" {
			parsed, err := time.Parse(time.RFC3339, completedAt)
			if err == nil {
				period.CompletedAt = &parsed
			}
		}
		periods = append(periods, period)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate periods: %w", err)
	}

	return periods, nil
}

func (cs *CycleService) listCycleHistoryByUser(ctx context.Context, userID int64, limit int) ([]model.CycleHistoryItem, error) {
	if limit <= 0 {
		limit = 12
	}

	rows, err := cs.db.QueryContext(
		ctx,
		`SELECT c.id, c.start_date, c.end_date, c.total_budget, c.period_count,
		        (SELECT COUNT(1) FROM periods p WHERE p.cycle_id = c.id) AS total_periods,
		        (SELECT COUNT(1) FROM periods p WHERE p.cycle_id = c.id AND p.status = ?) AS completed_periods,
		        COALESCE((SELECT SUM(p.savings) FROM periods p WHERE p.cycle_id = c.id), 0) AS total_savings,
		        COALESCE((SELECT SUM(p.overspending) FROM periods p WHERE p.cycle_id = c.id), 0) AS total_defisit
		 FROM cycles c
		 WHERE c.user_id = ?
		 ORDER BY c.id DESC
		 LIMIT ?`,
		model.PeriodStatusCompleted,
		userID,
		limit,
	)
	if err != nil {
		return nil, fmt.Errorf("query cycle history: %w", err)
	}
	defer rows.Close()

	history := make([]model.CycleHistoryItem, 0, limit)
	for rows.Next() {
		var item model.CycleHistoryItem
		var startDate string
		var endDate string

		if err := rows.Scan(
			&item.ID,
			&startDate,
			&endDate,
			&item.TotalBudget,
			&item.PeriodCount,
			&item.TotalPeriods,
			&item.CompletedPeriods,
			&item.TotalSavings,
			&item.TotalDefisit,
		); err != nil {
			return nil, fmt.Errorf("scan cycle history: %w", err)
		}

		item.StartDate, _ = time.ParseInLocation(dateLayout, startDate, time.Local)
		item.EndDate, _ = time.ParseInLocation(dateLayout, endDate, time.Local)
		item.IsCompleted = item.TotalPeriods > 0 && item.TotalPeriods == item.CompletedPeriods
		history = append(history, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate cycle history: %w", err)
	}

	return history, nil
}

func (cs *CycleService) GetCycleHistoryPageByUser(ctx context.Context, userID int64, page, pageSize int) ([]model.CycleHistoryItem, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	var total int
	if err := cs.db.QueryRowContext(ctx, `SELECT COUNT(1) FROM cycles WHERE user_id = ?`, userID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count cycle history: %w", err)
	}

	if total == 0 {
		return []model.CycleHistoryItem{}, 0, nil
	}

	offset := (page - 1) * pageSize
	rows, err := cs.db.QueryContext(
		ctx,
		`SELECT c.id, c.start_date, c.end_date, c.total_budget, c.period_count,
		        (SELECT COUNT(1) FROM periods p WHERE p.cycle_id = c.id) AS total_periods,
		        (SELECT COUNT(1) FROM periods p WHERE p.cycle_id = c.id AND p.status = ?) AS completed_periods,
		        COALESCE((SELECT SUM(p.savings) FROM periods p WHERE p.cycle_id = c.id), 0) AS total_savings,
		        COALESCE((SELECT SUM(p.overspending) FROM periods p WHERE p.cycle_id = c.id), 0) AS total_defisit
		 FROM cycles c
		 WHERE c.user_id = ?
		 ORDER BY c.id DESC
		 LIMIT ? OFFSET ?`,
		model.PeriodStatusCompleted,
		userID,
		pageSize,
		offset,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("query paged cycle history: %w", err)
	}
	defer rows.Close()

	history := make([]model.CycleHistoryItem, 0, pageSize)
	for rows.Next() {
		var item model.CycleHistoryItem
		var startDate string
		var endDate string

		if err := rows.Scan(
			&item.ID,
			&startDate,
			&endDate,
			&item.TotalBudget,
			&item.PeriodCount,
			&item.TotalPeriods,
			&item.CompletedPeriods,
			&item.TotalSavings,
			&item.TotalDefisit,
		); err != nil {
			return nil, 0, fmt.Errorf("scan paged cycle history: %w", err)
		}

		item.StartDate, _ = time.ParseInLocation(dateLayout, startDate, time.Local)
		item.EndDate, _ = time.ParseInLocation(dateLayout, endDate, time.Local)
		item.IsCompleted = item.TotalPeriods > 0 && item.TotalPeriods == item.CompletedPeriods
		history = append(history, item)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate paged cycle history: %w", err)
	}

	return history, total, nil
}

func (cs *CycleService) ClearCompletedCycleHistoryByUser(ctx context.Context, userID int64) error {
	_, err := cs.db.ExecContext(
		ctx,
		`DELETE FROM cycles
		 WHERE user_id = ?
		   AND EXISTS (
			   SELECT 1 FROM periods p WHERE p.cycle_id = cycles.id
		   )
		   AND NOT EXISTS (
			   SELECT 1 FROM periods p WHERE p.cycle_id = cycles.id AND p.status != ?
		   )`,
		userID,
		model.PeriodStatusCompleted,
	)
	if err != nil {
		return fmt.Errorf("clear cycle history: %w", err)
	}
	return nil
}

func normalizeAdjustmentMode(mode string) string {
	switch mode {
	case model.AdjustmentUniformDailyBudget, model.AdjustmentUniformPeriodLength, model.AdjustmentAIMaxSave:
		return mode
	default:
		return model.AdjustmentUniformPeriodLength
	}
}

func (cs *CycleService) buildPeriodPlans(start, end time.Time, totalBudget, periodCount int, adjustmentMode string, aiWeights []float64) []model.PeriodPlan {
	adjustmentMode = normalizeAdjustmentMode(adjustmentMode)
	nDays := int(end.Sub(start).Hours()/24) + 1
	baseDays := nDays / periodCount
	extraDays := nDays % periodCount

	periodDays := make([]int, periodCount)
	for i := 0; i < periodCount; i++ {
		periodDays[i] = baseDays
	}
	for i := 0; i < extraDays; i++ {
		periodDays[i]++
	}

	periodBudgets := make([]int, periodCount)
	if adjustmentMode == model.AdjustmentUniformDailyBudget {
		baseDailyBudget := totalBudget / nDays
		extraDailyBudget := totalBudget % nDays
		remainingExtra := extraDailyBudget
		for i, days := range periodDays {
			periodBudgets[i] = baseDailyBudget * days
			if remainingExtra > 0 {
				add := days
				if add > remainingExtra {
					add = remainingExtra
				}
				periodBudgets[i] += add
				remainingExtra -= add
			}
		}
	} else if adjustmentMode == model.AdjustmentAIMaxSave {
		periodBudgets = allocateByWeights(totalBudget, aiWeights, periodCount)
	} else {
		baseBudget := totalBudget / periodCount
		extraBudget := totalBudget % periodCount
		for i := 0; i < periodCount; i++ {
			periodBudgets[i] = baseBudget
		}
		for i := 0; i < extraBudget; i++ {
			periodBudgets[i]++
		}
	}

	plans := make([]model.PeriodPlan, 0, periodCount)
	periodStart := start
	for i := 1; i <= periodCount; i++ {
		days := periodDays[i-1]
		budget := periodBudgets[i-1]

		periodEnd := periodStart.AddDate(0, 0, days-1)
		plans = append(plans, model.PeriodPlan{
			PeriodNumber: i,
			StartDate:    periodStart.Format(dateLayout),
			EndDate:      periodEnd.Format(dateLayout),
			Days:         days,
			Budget:       budget,
		})

		periodStart = periodEnd.AddDate(0, 0, 1)
	}

	return plans
}

func allocateByWeights(totalBudget int, weights []float64, periodCount int) []int {
	if periodCount <= 0 {
		return []int{}
	}
	if len(weights) != periodCount {
		uniform := make([]float64, periodCount)
		for i := 0; i < periodCount; i++ {
			uniform[i] = 1
		}
		weights = uniform
	}

	totalWeight := 0.0
	for _, w := range weights {
		if w > 0 {
			totalWeight += w
		}
	}
	if totalWeight <= 0 {
		base := totalBudget / periodCount
		extra := totalBudget % periodCount
		budgets := make([]int, periodCount)
		for i := range budgets {
			budgets[i] = base
			if i < extra {
				budgets[i]++
			}
		}
		return budgets
	}

	type fractional struct {
		idx  int
		frac float64
	}

	budgets := make([]int, periodCount)
	fracs := make([]fractional, 0, periodCount)
	allocated := 0
	for i, w := range weights {
		normalized := w
		if normalized < 0 {
			normalized = 0
		}
		raw := (normalized / totalWeight) * float64(totalBudget)
		base := int(math.Floor(raw))
		budgets[i] = base
		allocated += base
		fracs = append(fracs, fractional{idx: i, frac: raw - float64(base)})
	}

	remaining := totalBudget - allocated
	sort.Slice(fracs, func(i, j int) bool {
		if fracs[i].frac == fracs[j].frac {
			return fracs[i].idx < fracs[j].idx
		}
		return fracs[i].frac > fracs[j].frac
	})
	for i := 0; i < remaining; i++ {
		budgets[fracs[i%len(fracs)].idx]++
	}

	return budgets
}

func (cs *CycleService) getAIMaxSaveWeights(ctx context.Context, userID int64, periodCount int) ([]float64, error) {
	if periodCount <= 0 {
		return nil, nil
	}

	var cycleID int64
	err := cs.db.QueryRowContext(
		ctx,
		`SELECT c.id
		 FROM cycles c
		 WHERE c.user_id = ?
		   AND EXISTS (SELECT 1 FROM periods p WHERE p.cycle_id = c.id)
		   AND NOT EXISTS (SELECT 1 FROM periods p WHERE p.cycle_id = c.id AND p.status != ?)
		 ORDER BY c.end_date DESC, c.id DESC
		 LIMIT 1`,
		userID,
		model.PeriodStatusCompleted,
	).Scan(&cycleID)
	if err != nil {
		if err == sql.ErrNoRows {
			return buildAIMaxSaveColdStartWeights(periodCount), nil
		}
		return nil, fmt.Errorf("load ai source cycle: %w", err)
	}

	rows, err := cs.db.QueryContext(
		ctx,
		`SELECT period_number, budget, savings, overspending
		 FROM periods
		 WHERE cycle_id = ?
		 ORDER BY period_number ASC`,
		cycleID,
	)
	if err != nil {
		return nil, fmt.Errorf("load ai source periods: %w", err)
	}
	defer rows.Close()

	source := make([]float64, 0, periodCount)
	for rows.Next() {
		var periodNumber, budget, savings, overspending int
		if err := rows.Scan(&periodNumber, &budget, &savings, &overspending); err != nil {
			return nil, fmt.Errorf("scan ai source period: %w", err)
		}
		if budget <= 0 {
			source = append(source, 1)
			continue
		}
		spending := budget - savings + overspending
		ratio := float64(spending) / float64(budget)
		if ratio < 0.7 {
			ratio = 0.7
		}
		if ratio > 1.6 {
			ratio = 1.6
		}
		source = append(source, ratio)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate ai source periods: %w", err)
	}
	if len(source) == 0 {
		return buildAIMaxSaveColdStartWeights(periodCount), nil
	}

	weights := make([]float64, periodCount)
	if len(source) == 1 {
		coldStart := buildAIMaxSaveColdStartWeights(periodCount)
		for i := range weights {
			weights[i] = (source[0] * 0.75) + (coldStart[i] * 0.25)
		}
		return weights, nil
	}

	for i := 0; i < periodCount; i++ {
		srcIdx := int(math.Round(float64(i) * float64(len(source)-1) / float64(periodCount-1)))
		if srcIdx < 0 {
			srcIdx = 0
		}
		if srcIdx >= len(source) {
			srcIdx = len(source) - 1
		}
		weights[i] = source[srcIdx]
	}

	if len(source) < 3 {
		coldStart := buildAIMaxSaveColdStartWeights(periodCount)
		for i := range weights {
			weights[i] = (weights[i] * 0.8) + (coldStart[i] * 0.2)
		}
	}

	return weights, nil
}

func buildAIMaxSaveColdStartWeights(periodCount int) []float64 {
	if periodCount <= 0 {
		return []float64{}
	}
	if periodCount == 1 {
		return []float64{1.0}
	}

	weights := make([]float64, periodCount)
	for i := 0; i < periodCount; i++ {
		x := float64(i) / float64(periodCount-1)
		edgeBias := math.Abs(x-0.5) * 2.0 // 0 at center, 1 at edges
		weight := 1.0 + (0.28 * edgeBias)
		if i == periodCount-1 {
			// Keep a slightly larger tail buffer to reduce end-cycle overspending risk.
			weight += 0.08
		}
		weights[i] = weight
	}
	return weights
}
