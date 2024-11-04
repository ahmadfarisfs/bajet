package routes

import (
	"bajetapp/model"
	"bajetapp/mwr"
	"bajetapp/services"
	"bajetapp/utils"
	"fmt"

	"cloud.google.com/go/civil"
	"github.com/labstack/echo/v4"
)

type TransactionRoutes struct {
	tsvc *services.TransactionService
}

func NewTransactionRoutes(tsvc *services.TransactionService, e *echo.Echo) *TransactionRoutes {
	ret := &TransactionRoutes{
		tsvc: tsvc,
	}
	ret.registerRoutes(e)
	return ret
}

func (tr *TransactionRoutes) registerRoutes(e *echo.Echo) {
	trxGroup := e.Group("/trx", mwr.AuthMiddleware)
	trxGroup.GET("", tr.handleGetTransactions)
	trxGroup.POST("", tr.handleCreateTransaction)
	trxGroup.DELETE("/:id", tr.handleDeleteTransaction)
}

func (tr *TransactionRoutes) handleGetTransactions(c echo.Context) error {
	type request struct {
		StartDate civil.Date `query:"start_date"`
		EndDate   civil.Date `query:"end_date"`
	}
	var req request
	if err := utils.BindValidate(c, &req); err != nil {
		return fmt.Errorf("failed to bind validate payload: %w", err)
	}

	userInfo := mwr.GetLoginInfo(c)
	if userInfo == nil {
		return fmt.Errorf("failed to get login info")
	}

	_, err := tr.tsvc.GetTransactions(c.Request().Context(), userInfo.ID, req.StartDate, req.EndDate)
	if err != nil {
		return fmt.Errorf("failed to get transactions: %w", err)
	}
	/* render data if success */
	return c.JSON(200, "Get transactions")
}

func (tr *TransactionRoutes) handleCreateTransaction(c echo.Context) error {

	var trx model.Transaction
	if err := utils.BindValidate(c, &trx); err != nil {
		return fmt.Errorf("failed to bind validate payload: %w", err)
	}

	userInfo := mwr.GetLoginInfo(c)
	if userInfo == nil {
		return fmt.Errorf("failed to get login info")
	}

	trx.UserID = userInfo.ID
	return tr.tsvc.CreateTransaction(c.Request().Context(), trx)

}

func (tr *TransactionRoutes) handleDeleteTransaction(c echo.Context) error {
	return c.JSON(200, "Delete transaction")
}
