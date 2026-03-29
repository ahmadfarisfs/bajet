package model

type PeriodPlan struct {
	PeriodNumber int
	StartDate    string
	EndDate      string
	Days         int
	Budget       int
}

const (
	AdjustmentUniformDailyBudget  = "uniform_daily_budget"
	AdjustmentUniformPeriodLength = "uniform_period_length"
	AdjustmentAIMaxSave           = "ai_max_save"
)

type OnboardingDraft struct {
	StartDate            string
	EndDate              string
	PreviousCycleEndDate string
	CurrencyCode         string
	Language             string
	TotalBudget          int
	PeriodCount          int
	AdjustmentMode       string
	Plans                []PeriodPlan
	Error                string
}
