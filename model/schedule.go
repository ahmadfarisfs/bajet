package model

type FiscalRange struct {
	StartDate string `json:"start_date" bson:"start_date" form:"start_date" validate:"required"`
	EndDate   string `json:"end_date" bson:"end_date" form:"end_date" validate:"required"`
}
