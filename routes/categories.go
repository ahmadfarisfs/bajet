package routes

import "github.com/labstack/echo/v4"

type CategoryRoutes struct {
}

func NewCategoryRoutes(e *echo.Echo) *CategoryRoutes {
	return &CategoryRoutes{}
}
