package auth

import (
	"net/http"
	"server/internal/services"
	"server/internal/services/users"

	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

type LoginUserType struct {
	Email    string `validate:"required"`
	Password string `validate:"required"`
}

func Login(c echo.Context) error {
	var newUser LoginUserType

	err := c.Bind(&newUser)
	if err != nil {
		return err
	}

	err = services.ValidateStruct(newUser)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	user, err := users.GetUserByEmail(newUser.Email)
	if err != nil {
		return echo.ErrUnauthorized
	}

	if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(newUser.Password)) != nil {
		return echo.ErrUnauthorized
	}

	response, err := RevokeToken(user, "")
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, response)
}
