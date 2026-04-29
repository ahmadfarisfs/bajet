package models

import "time"

type DivisionMode string

const (
	ModeEqual      DivisionMode = "equal"
	ModeBehavioral DivisionMode = "behavioral"
)

type PeriodStatus string

const (
	StatusOpen      PeriodStatus = "open"
	StatusCompleted PeriodStatus = "completed"
)

type ResultType string

const (
	ResultSisa    ResultType = "sisa"
	ResultDefisit ResultType = "defisit"
)

type Cycle struct {
	ID           uint         `json:"id" gorm:"primarykey"`
	UserID       string       `json:"-" gorm:"index;not null;default:''"`
	StartDate    time.Time    `json:"start_date"`
	EndDate      time.Time    `json:"end_date"`
	TotalBudget  float64      `json:"total_budget"`
	DivisionMode DivisionMode `json:"division_mode"`
	CreatedAt    time.Time    `json:"created_at"`
	Periods      []Period     `json:"periods" gorm:"foreignKey:CycleID"`
}

type Period struct {
	ID           uint         `json:"id" gorm:"primarykey"`
	CycleID      uint         `json:"cycle_id"`
	PeriodNumber int          `json:"period_number"`
	StartDate    time.Time    `json:"start_date"`
	EndDate      time.Time    `json:"end_date"`
	Budget       float64      `json:"budget"`
	Status       PeriodStatus `json:"status"`
	ResultType   ResultType   `json:"result_type"`
	ResultAmount float64      `json:"result_amount"`
	CreatedAt    time.Time    `json:"created_at"`
}
