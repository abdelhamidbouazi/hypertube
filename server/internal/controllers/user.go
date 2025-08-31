package controllers

import (
	"fmt"
	"net/http"
	"server/internal/services"

	"github.com/labstack/echo/v4"
)

func GetUsers(c echo.Context) error {
	for i := range c.Request().Header {
		fmt.Printf("%s", i)
	}
	users, err := services.GetUsers()
	if err != nil {
		response := echo.Map{
			"message": err.Error(),
		}
		return c.JSON(http.StatusOK, response)
	}

	response := echo.Map{
		"data": users,
	}

	return c.JSON(http.StatusOK, response)
}
