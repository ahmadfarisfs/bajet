package model

import "time"

const (
	PeriodStatusOpen      = "open"
	PeriodStatusCompleted = "completed"
	InputTypeSisa         = "sisa"
	InputTypeDefisit      = "defisit"
	PeriodTimingInactive  = "inactive"
	PeriodTimingEarly     = "early"
	PeriodTimingOnTrack   = "on_track"
	PeriodTimingLate      = "late"
)

type Cycle struct {
	ID          int64
	StartDate   time.Time
	EndDate     time.Time
	TotalBudget int
	PeriodCount int
	CreatedAt   time.Time
}

type CycleHistoryItem struct {
	ID               int64
	StartDate        time.Time
	EndDate          time.Time
	TotalBudget      int
	PeriodCount      int
	TotalPeriods     int
	CompletedPeriods int
	TotalSavings     int
	TotalDefisit     int
	IsCompleted      bool
}

type Period struct {
	ID           int64
	CycleID      int64
	PeriodNumber int
	StartDate    time.Time
	EndDate      time.Time
	Budget       int
	Status       string
	InputType    string
	InputAmount  int
	Savings      int
	Overspending int
	CompletedAt  *time.Time
}

type Dashboard struct {
	Cycle            Cycle
	Periods          []Period
	VisiblePeriods   []Period
	PeriodTiming     string
	CurrencyCode     string
	Language         string
	TotalSavings     int
	TotalSpending    int
	TotalDefisit     int
	NetResult        int
	OverallSavings   int
	OverallSpending  int
	OverallDefisit   int
	OverallNetResult int
	CanStartNext     bool
	History          []CycleHistoryItem
}
