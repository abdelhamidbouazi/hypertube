package auth

import (
	"net/http"
	"server/internal/services"

	"github.com/labstack/echo/v4"
)

func Register(c echo.Context) error {
	var newUser services.CreateUserType

	err := c.Bind(&newUser)
	if err != nil {
		return echo.ErrBadRequest
	}

	err = services.ValidateStruct(newUser)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	err = services.CreateUser(newUser)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.String(http.StatusOK, "success")
}
