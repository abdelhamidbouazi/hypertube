package controllers

import (
	"net/http"
	"server/internal/services/users"

	"github.com/labstack/echo/v4"
)

// Get users info godoc
//
//	@Summary		List of users
//	@Description	get users info
//	@Tags			users
//	@Produce		json
//	@Success		200	{array}	models.User
//	@Security		JWT
//	@Router			/users [get]
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

// Get user info godoc
//
//	@Summary		User info
//	@Description	get current user info
//	@Tags			users
//	@Security		JWT
//	@Produce		json
//	@Success		200	{object}	models.User
//	@Router			/users/me [get]
func GetMe(c echo.Context) error {
	return c.JSON(http.StatusOK, c.Get("model"))
}
