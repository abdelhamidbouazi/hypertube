package controllers

import (
	"net/http"
	"server/internal/services/users"

	"github.com/labstack/echo/v4"
)

func GetUsers(c echo.Context) error {
	users, err := users.GetUsers()
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

func GetMe(c echo.Context) error {
	return c.JSON(http.StatusOK, c.Get("model"))
}
