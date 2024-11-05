package model

import (
	"time"

	"cloud.google.com/go/civil"
)

type Transaction struct {
	ID            string    `json:"id" bson:"_id" validate:"-"`
	UserID        string    `json:"user_id" bson:"user_id" validate:"-"`
	Amount        int       `json:"amount" bson:"amount" form:"amount" validate:"required"`
	Description   string    `json:"description" bson:"description" form:"description" validate:"required"`
	Category      string    `json:"category" bson:"category" form:"category" validate:"required"`
	TransactionAt time.Time `json:"transaction_at" bson:"transaction_at" form:"transaction_at" validate:"required"`
	CreatedAt     time.Time `json:"created_at" bson:"created_at"`
}

type TransactionStats struct {
	TotalExpense     int        `json:"total_expense" bson:"total_expense"`
	TotalIncome      int        `json:"total_income" bson:"total_income"`
	DateStart        civil.Date `json:"date_start" bson:"date_start"`
	DateEnd          civil.Date `json:"date_end" bson:"date_end"`
	TransactionCount int        `json:"transaction_count" bson:"transaction_count"`
}
