package routes

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"bajetapp/model"
	"bajetapp/mwr"
	"bajetapp/services"
	"bajetapp/utils"
	"bajetapp/views/pages"

	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type CycleRoutes struct {
	cycleSvc *services.CycleService
	authSvc  *services.AuthLocalService
}

func NewCycleRoutes(cycleSvc *services.CycleService, authSvc *services.AuthLocalService, e *echo.Echo) *CycleRoutes {
	r := &CycleRoutes{cycleSvc: cycleSvc, authSvc: authSvc}
	r.registerRoutes(e)
	return r
}

func (cr *CycleRoutes) registerRoutes(e *echo.Echo) {
	e.GET("/", cr.handleRoot)
	e.GET("/login", cr.handleLoginPage)
	e.POST("/login", cr.handleLogin)
	e.GET("/register", cr.handleRegisterPage)
	e.POST("/register", cr.handleRegister)
	e.GET("/logout", cr.handleLogout)
	e.POST("/logout", cr.handleLogout)

	protected := e.Group("", mwr.RequireSessionAuth)
	protected.GET("/app", cr.handleDashboard)
	protected.GET("/cycle-history", cr.handleCycleHistoryPage)
	protected.POST("/cycle-history/clear", cr.handleClearCycleHistory)
	protected.GET("/onboarding/cycle", cr.handleOnboardingCyclePage)
	protected.POST("/onboarding/cycle", cr.handleOnboardingCycle)
	protected.GET("/onboarding/budget", cr.handleOnboardingBudgetPage)
	protected.POST("/onboarding/budget", cr.handleOnboardingBudget)
	protected.POST("/periods/:id/complete", cr.handleCompletePeriod)
	protected.GET("/preferences", cr.handlePreferencesPage)
	protected.POST("/preferences", cr.handleUpdatePreferences)
}

func (cr *CycleRoutes) handleRoot(c echo.Context) error {
	if _, ok := mwr.GetSessionUserID(c); ok {
		return c.Redirect(http.StatusSeeOther, "/app")
	}
	return c.Redirect(http.StatusSeeOther, "/login")
}

func (cr *CycleRoutes) handleLoginPage(c echo.Context) error {
	if _, ok := mwr.GetSessionUserID(c); ok {
		return c.Redirect(http.StatusSeeOther, "/app")
	}
	token := csrfToken(c)
	component := pages.LoginPage(token, "", resolveGuestLanguage(c, ""))
	utils.Render(c, &component)
	return nil
}

func (cr *CycleRoutes) handleRegisterPage(c echo.Context) error {
	if _, ok := mwr.GetSessionUserID(c); ok {
		return c.Redirect(http.StatusSeeOther, "/app")
	}
	token := csrfToken(c)
	component := pages.RegisterPage(token, "", resolveGuestLanguage(c, ""))
	utils.Render(c, &component)
	return nil
}

func (cr *CycleRoutes) handleRegister(c echo.Context) error {
	var req struct {
		Email    string `form:"email" validate:"required,email"`
		Password string `form:"password" validate:"required,min=6"`
		Language string `form:"language"`
	}
	if err := utils.BindValidate(c, &req); err != nil {
		return cr.renderRegisterWithError(c, err.Error(), req.Language)
	}

	userID, err := cr.authSvc.Register(c.Request().Context(), req.Email, req.Password)
	if err != nil {
		return cr.renderRegisterWithError(c, err.Error(), req.Language)
	}

	if lang := resolveGuestLanguage(c, req.Language); lang == "id" {
		_ = cr.cycleSvc.UpdateUserPreferences(c.Request().Context(), userID, "USD", "id")
	}

	if err := mwr.SetSessionUserID(c, userID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to save session")
	}

	return c.Redirect(http.StatusSeeOther, "/onboarding/cycle")
}

func (cr *CycleRoutes) handleLogin(c echo.Context) error {
	var req struct {
		Email    string `form:"email" validate:"required,email"`
		Password string `form:"password" validate:"required"`
		Language string `form:"language"`
	}
	if err := utils.BindValidate(c, &req); err != nil {
		return cr.renderLoginWithError(c, err.Error(), req.Language)
	}

	userID, err := cr.authSvc.Login(c.Request().Context(), req.Email, req.Password)
	if err != nil {
		return cr.renderLoginWithError(c, err.Error(), req.Language)
	}

	if err := mwr.SetSessionUserID(c, userID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to save session")
	}

	hasCycle, err := cr.cycleSvc.HasAnyCycle(c.Request().Context(), userID)
	if err != nil {
		return err
	}
	if !hasCycle {
		return c.Redirect(http.StatusSeeOther, "/onboarding/cycle")
	}

	return c.Redirect(http.StatusSeeOther, "/app")
}

func (cr *CycleRoutes) handleLogout(c echo.Context) error {
	if err := mwr.ClearSession(c); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to clear session")
	}
	return c.Redirect(http.StatusSeeOther, "/login")
}

func (cr *CycleRoutes) handleDashboard(c echo.Context) error {
	userID := c.Get("user_id").(int64)
	dashboard, err := cr.cycleSvc.GetDashboardByUser(c.Request().Context(), userID, time.Now())
	if err != nil {
		if err == services.ErrNoCycle {
			return c.Redirect(http.StatusSeeOther, "/onboarding/cycle")
		}
		return fmt.Errorf("load dashboard: %w", err)
	}

	token := csrfToken(c)
	component := pages.MainPage(dashboard, token)
	utils.Render(c, &component)
	return nil
}

func (cr *CycleRoutes) handleCycleHistoryPage(c echo.Context) error {
	userID := c.Get("user_id").(int64)
	dashboard, err := cr.cycleSvc.GetDashboardByUser(c.Request().Context(), userID, time.Now())
	if err != nil {
		if err == services.ErrNoCycle {
			return c.Redirect(http.StatusSeeOther, "/onboarding/cycle")
		}
		return fmt.Errorf("load cycle history: %w", err)
	}

	viewMode := strings.ToLower(strings.TrimSpace(c.QueryParam("view")))
	if viewMode != "table" && viewMode != "card" {
		viewMode = "card"
	}

	page := 1
	if raw := strings.TrimSpace(c.QueryParam("page")); raw != "" {
		if parsed, parseErr := strconv.Atoi(raw); parseErr == nil && parsed > 0 {
			page = parsed
		}
	}

	const pageSize = 10
	history, totalItems, err := cr.cycleSvc.GetCycleHistoryPageByUser(c.Request().Context(), userID, page, pageSize)
	if err != nil {
		return fmt.Errorf("load paged cycle history: %w", err)
	}

	totalPages := 1
	if totalItems > 0 {
		totalPages = (totalItems + pageSize - 1) / pageSize
	}

	if page > totalPages {
		page = totalPages
		history, totalItems, err = cr.cycleSvc.GetCycleHistoryPageByUser(c.Request().Context(), userID, page, pageSize)
		if err != nil {
			return fmt.Errorf("reload paged cycle history: %w", err)
		}
	}

	dashboard.History = history

	component := pages.CycleHistoryPage(dashboard, csrfToken(c), viewMode, page, pageSize, totalItems, totalPages)
	utils.Render(c, &component)
	return nil
}

func (cr *CycleRoutes) handleClearCycleHistory(c echo.Context) error {
	userID := c.Get("user_id").(int64)
	if err := cr.cycleSvc.ClearCompletedCycleHistoryByUser(c.Request().Context(), userID); err != nil {
		return fmt.Errorf("clear cycle history: %w", err)
	}
	return c.Redirect(http.StatusSeeOther, "/cycle-history")
}

func (cr *CycleRoutes) handleOnboardingCyclePage(c echo.Context) error {
	userID := c.Get("user_id").(int64)
	start, end, err := cr.cycleSvc.SuggestNextWindowByUser(c.Request().Context(), userID, time.Now())
	if err != nil {
		return fmt.Errorf("suggest next cycle window: %w", err)
	}
	previousCycleEndDate := ""
	if latest, latestErr := cr.cycleSvc.GetLatestCycleByUser(c.Request().Context(), userID); latestErr == nil {
		previousCycleEndDate = latest.EndDate.Format("2006-01-02")
	}
	currencyCode := "USD"
	language := "en"
	if currentCurrency, currentLanguage, prefErr := cr.cycleSvc.GetUserPreferences(c.Request().Context(), userID); prefErr == nil {
		if strings.TrimSpace(currentCurrency) != "" {
			currencyCode = currentCurrency
		}
		if strings.TrimSpace(currentLanguage) != "" {
			language = currentLanguage
		}
	}
	if strings.TrimSpace(currencyCode) == "" {
		currencyCode = "USD"
	}
	if strings.TrimSpace(language) == "" {
		language = "en"
	}
	draft := model.OnboardingDraft{
		StartDate:            start.Format("2006-01-02"),
		EndDate:              end.Format("2006-01-02"),
		PreviousCycleEndDate: previousCycleEndDate,
		CurrencyCode:         currencyCode,
		Language:             language,
		TotalBudget:          2_000_000,
		PeriodCount:          4,
	}
	component := pages.OnboardingCyclePage(draft, csrfToken(c))
	utils.Render(c, &component)
	return nil
}

func (cr *CycleRoutes) handleOnboardingCycle(c echo.Context) error {
	var req struct {
		StartDate    string `form:"start_date" validate:"required"`
		EndDate      string `form:"end_date" validate:"required"`
		CurrencyCode string `form:"currency_code" validate:"required,oneof=USD IDR EUR SGD GBP JPY"`
		Language     string `form:"language" validate:"required,oneof=en id"`
	}
	if err := utils.BindValidate(c, &req); err != nil {
		return cr.renderCycleOnboardingWithError(c, req.StartDate, req.EndDate, req.CurrencyCode, req.Language, err.Error())
	}
	req.CurrencyCode = strings.ToUpper(strings.TrimSpace(req.CurrencyCode))
	req.Language = strings.ToLower(strings.TrimSpace(req.Language))

	if err := cr.cycleSvc.UpdateUserPreferences(c.Request().Context(), c.Get("user_id").(int64), req.CurrencyCode, req.Language); err != nil {
		return cr.renderCycleOnboardingWithError(c, req.StartDate, req.EndDate, req.CurrencyCode, req.Language, err.Error())
	}

	if _, err := cr.cycleSvc.GetSuggestedPlan(c.Request().Context(), c.Get("user_id").(int64), req.StartDate, req.EndDate, 2_000_000, 4, model.AdjustmentUniformPeriodLength); err != nil {
		return cr.renderCycleOnboardingWithError(c, req.StartDate, req.EndDate, req.CurrencyCode, req.Language, err.Error())
	}

	sess, _ := session.Get("session", c)
	sess.Values["onboarding_start_date"] = req.StartDate
	sess.Values["onboarding_end_date"] = req.EndDate
	if err := sess.Save(c.Request(), c.Response()); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to save onboarding state")
	}

	return c.Redirect(http.StatusSeeOther, "/onboarding/budget")
}

func (cr *CycleRoutes) handleOnboardingBudgetPage(c echo.Context) error {
	startDate, endDate, ok := cr.readOnboardingDates(c)
	if !ok {
		return c.Redirect(http.StatusSeeOther, "/onboarding/cycle")
	}
	userID := c.Get("user_id").(int64)
	currencyCode := "USD"
	language := "en"
	if currentCurrency, currentLanguage, prefErr := cr.cycleSvc.GetUserPreferences(c.Request().Context(), userID); prefErr == nil {
		if strings.TrimSpace(currentCurrency) != "" {
			currencyCode = currentCurrency
		}
		if strings.TrimSpace(currentLanguage) != "" {
			language = currentLanguage
		}
	}

	periodCount := 4
	if raw := c.QueryParam("period_count"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil {
			periodCount = parsed
		}
	}

	adjustmentMode := strings.TrimSpace(c.QueryParam("adjustment_mode"))
	if adjustmentMode == "" {
		adjustmentMode = model.AdjustmentUniformPeriodLength
	}
	if adjustmentMode != model.AdjustmentUniformDailyBudget && adjustmentMode != model.AdjustmentUniformPeriodLength && adjustmentMode != model.AdjustmentAIMaxSave {
		adjustmentMode = model.AdjustmentUniformPeriodLength
	}

	plans, err := cr.cycleSvc.GetSuggestedPlan(c.Request().Context(), userID, startDate, endDate, 2_000_000, periodCount, adjustmentMode)
	if err != nil {
		return cr.renderBudgetOnboardingWithError(c, startDate, endDate, 2_000_000, 4, adjustmentMode, err.Error())
	}

	draft := model.OnboardingDraft{
		StartDate:      startDate,
		EndDate:        endDate,
		CurrencyCode:   currencyCode,
		Language:       language,
		TotalBudget:    2_000_000,
		PeriodCount:    periodCount,
		AdjustmentMode: adjustmentMode,
		Plans:          plans,
	}

	component := pages.OnboardingBudgetPage(draft, csrfToken(c))
	utils.Render(c, &component)
	return nil
}

func (cr *CycleRoutes) handleOnboardingBudget(c echo.Context) error {
	userID := c.Get("user_id").(int64)
	startDate, endDate, ok := cr.readOnboardingDates(c)
	if !ok {
		return c.Redirect(http.StatusSeeOther, "/onboarding/cycle")
	}

	var req struct {
		TotalBudget    int    `form:"total_budget" validate:"gt=0"`
		PeriodCount    int    `form:"period_count" validate:"gte=2,lte=8"`
		AdjustmentMode string `form:"adjustment_mode" validate:"required,oneof=uniform_daily_budget uniform_period_length ai_max_save"`
	}
	if err := utils.BindValidate(c, &req); err != nil {
		return cr.renderBudgetOnboardingWithError(c, startDate, endDate, req.TotalBudget, req.PeriodCount, req.AdjustmentMode, err.Error())
	}
	req.AdjustmentMode = strings.TrimSpace(req.AdjustmentMode)

	if _, err := cr.cycleSvc.CreateCycleForUser(c.Request().Context(), userID, startDate, endDate, req.TotalBudget, req.PeriodCount, req.AdjustmentMode); err != nil {
		if err == services.ErrCycleOverlap {
			return cr.renderBudgetOnboardingWithError(c, startDate, endDate, req.TotalBudget, req.PeriodCount, req.AdjustmentMode, "new cycle cannot overlap with previous cycle unless previous cycle is deleted")
		}
		return cr.renderBudgetOnboardingWithError(c, startDate, endDate, req.TotalBudget, req.PeriodCount, req.AdjustmentMode, err.Error())
	}

	sess, _ := session.Get("session", c)
	delete(sess.Values, "onboarding_start_date")
	delete(sess.Values, "onboarding_end_date")
	_ = sess.Save(c.Request(), c.Response())

	return c.Redirect(http.StatusSeeOther, "/app")
}

func (cr *CycleRoutes) handleCompletePeriod(c echo.Context) error {
	userID := c.Get("user_id").(int64)
	periodID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid period id")
	}

	var req struct {
		InputType string `form:"input_type" validate:"required,oneof=sisa defisit"`
		Amount    int    `form:"amount" validate:"gte=0"`
	}
	if err := utils.BindValidate(c, &req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if err := cr.cycleSvc.CompletePeriodByUser(c.Request().Context(), userID, periodID, req.InputType, req.Amount); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return c.Redirect(http.StatusSeeOther, "/app")
}

func (cr *CycleRoutes) handlePreferencesPage(c echo.Context) error {
	userID := c.Get("user_id").(int64)
	currencyCode, language, err := cr.cycleSvc.GetUserPreferences(c.Request().Context(), userID)
	if err != nil {
		return fmt.Errorf("load preferences: %w", err)
	}

	component := pages.PreferencesPage(currencyCode, language, csrfToken(c), "")
	utils.Render(c, &component)
	return nil
}

func (cr *CycleRoutes) handleUpdatePreferences(c echo.Context) error {
	userID := c.Get("user_id").(int64)

	var req struct {
		CurrencyCode string `form:"currency_code" validate:"required,oneof=USD IDR EUR SGD GBP JPY"`
		Language     string `form:"language" validate:"required,oneof=en id"`
	}
	if err := utils.BindValidate(c, &req); err != nil {
		return cr.renderPreferencesWithError(c, req.CurrencyCode, req.Language, err.Error())
	}

	req.CurrencyCode = strings.ToUpper(strings.TrimSpace(req.CurrencyCode))
	req.Language = strings.ToLower(strings.TrimSpace(req.Language))

	if err := cr.cycleSvc.UpdateUserPreferences(c.Request().Context(), userID, req.CurrencyCode, req.Language); err != nil {
		return cr.renderPreferencesWithError(c, req.CurrencyCode, req.Language, err.Error())
	}

	return c.Redirect(http.StatusSeeOther, "/preferences")
}

func csrfToken(c echo.Context) string {
	rawToken := c.Get(middleware.DefaultCSRFConfig.ContextKey)
	if rawToken == nil {
		return ""
	}
	token, _ := rawToken.(string)
	return token
}

func (cr *CycleRoutes) readOnboardingDates(c echo.Context) (string, string, bool) {
	sess, err := session.Get("session", c)
	if err != nil {
		return "", "", false
	}
	startDate, _ := sess.Values["onboarding_start_date"].(string)
	endDate, _ := sess.Values["onboarding_end_date"].(string)
	if startDate == "" || endDate == "" {
		return "", "", false
	}
	return startDate, endDate, true
}

func (cr *CycleRoutes) renderLoginWithError(c echo.Context, msg, fallbackLanguage string) error {
	component := pages.LoginPage(csrfToken(c), msg, resolveGuestLanguage(c, fallbackLanguage))
	utils.Render(c, &component)
	return nil
}

func (cr *CycleRoutes) renderRegisterWithError(c echo.Context, msg, fallbackLanguage string) error {
	component := pages.RegisterPage(csrfToken(c), msg, resolveGuestLanguage(c, fallbackLanguage))
	utils.Render(c, &component)
	return nil
}

func resolveGuestLanguage(c echo.Context, fallback string) string {
	if strings.TrimSpace(fallback) != "" {
		fallback = strings.ToLower(strings.TrimSpace(fallback))
		if fallback == "id" || fallback == "en" {
			return fallback
		}
	}
	lang := strings.ToLower(strings.TrimSpace(c.QueryParam("lang")))
	if lang == "id" || lang == "en" {
		return lang
	}
	return "en"
}

func (cr *CycleRoutes) renderCycleOnboardingWithError(c echo.Context, startDate, endDate, currencyCode, language, msg string) error {
	userID := c.Get("user_id").(int64)
	previousCycleEndDate := ""
	if latest, latestErr := cr.cycleSvc.GetLatestCycleByUser(c.Request().Context(), userID); latestErr == nil {
		previousCycleEndDate = latest.EndDate.Format("2006-01-02")
	}
	if currencyCode == "" || language == "" {
		currentCurrency, currentLanguage, prefErr := cr.cycleSvc.GetUserPreferences(c.Request().Context(), userID)
		if prefErr == nil {
			if currencyCode == "" {
				currencyCode = currentCurrency
			}
			if language == "" {
				language = currentLanguage
			}
		}
	}
	if currencyCode == "" {
		currencyCode = "USD"
	}
	if language == "" {
		language = "en"
	}
	if startDate == "" || endDate == "" {
		start, end, err := cr.cycleSvc.SuggestNextWindowByUser(c.Request().Context(), userID, time.Now())
		if err != nil {
			start, end = cr.cycleSvc.SuggestDefaultWindow(time.Now())
		}
		startDate = start.Format("2006-01-02")
		endDate = end.Format("2006-01-02")
	}
	draft := model.OnboardingDraft{
		StartDate:            startDate,
		EndDate:              endDate,
		PreviousCycleEndDate: previousCycleEndDate,
		CurrencyCode:         currencyCode,
		Language:             language,
		TotalBudget:          2_000_000,
		PeriodCount:          4,
		Error:                msg,
	}
	component := pages.OnboardingCyclePage(draft, csrfToken(c))
	utils.Render(c, &component)
	return nil
}

func (cr *CycleRoutes) renderBudgetOnboardingWithError(c echo.Context, startDate, endDate string, totalBudget, periodCount int, adjustmentMode, msg string) error {
	userID := c.Get("user_id").(int64)
	currencyCode := "USD"
	language := "en"
	if currentCurrency, currentLanguage, prefErr := cr.cycleSvc.GetUserPreferences(c.Request().Context(), userID); prefErr == nil {
		if strings.TrimSpace(currentCurrency) != "" {
			currencyCode = currentCurrency
		}
		if strings.TrimSpace(currentLanguage) != "" {
			language = currentLanguage
		}
	}
	if totalBudget <= 0 {
		totalBudget = 2_000_000
	}
	if periodCount < 2 || periodCount > 8 {
		periodCount = 4
	}
	if adjustmentMode != model.AdjustmentUniformDailyBudget && adjustmentMode != model.AdjustmentUniformPeriodLength && adjustmentMode != model.AdjustmentAIMaxSave {
		adjustmentMode = model.AdjustmentUniformPeriodLength
	}
	plans, _ := cr.cycleSvc.GetSuggestedPlan(c.Request().Context(), userID, startDate, endDate, totalBudget, periodCount, adjustmentMode)
	draft := model.OnboardingDraft{
		StartDate:      startDate,
		EndDate:        endDate,
		CurrencyCode:   currencyCode,
		Language:       language,
		TotalBudget:    totalBudget,
		PeriodCount:    periodCount,
		AdjustmentMode: adjustmentMode,
		Plans:          plans,
		Error:          msg,
	}
	component := pages.OnboardingBudgetPage(draft, csrfToken(c))
	utils.Render(c, &component)
	return nil
}

func (cr *CycleRoutes) renderPreferencesWithError(c echo.Context, currencyCode, language, msg string) error {
	userID := c.Get("user_id").(int64)

	currencyCode = strings.ToUpper(strings.TrimSpace(currencyCode))
	language = strings.ToLower(strings.TrimSpace(language))

	if currencyCode == "" || language == "" {
		currentCurrency, currentLanguage, err := cr.cycleSvc.GetUserPreferences(c.Request().Context(), userID)
		if err == nil {
			if currencyCode == "" {
				currencyCode = currentCurrency
			}
			if language == "" {
				language = currentLanguage
			}
		}
	}
	if currencyCode == "" {
		currencyCode = "USD"
	}
	if language == "" {
		language = "en"
	}

	component := pages.PreferencesPage(currencyCode, language, csrfToken(c), msg)
	utils.Render(c, &component)
	return nil
}
