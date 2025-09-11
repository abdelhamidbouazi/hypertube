package auth

import (
	"net/http"
	"server/internal/services"
	"server/internal/services/users"

	"github.com/labstack/echo/v4"
)

type RegisterUserType struct {
	FirstName string `validate:"required,min=4"`
	LastName  string `validate:"required,min=4"`
	Email     string `validate:"required,email"`
	Password  string `validate:"required,min=8"`
}

// Register godoc
//
//	@Summary		Register
//	@Description	Register new user
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			RegisterUserType	body		RegisterUserType	true	"register credentials to send"
//	@Success		200					{string}	string				"success"
//	@Router			/auth/register [post]
func Register(c echo.Context) error {
	var newUser RegisterUserType

	err := c.Bind(&newUser)
	if err != nil {
		return echo.ErrBadRequest
	}

	var user users.CreateUserType
	user.Provider = ""
	user.Email = newUser.Email
	user.FirstName = newUser.FirstName
	user.LastName = newUser.LastName
	user.Password = newUser.Password

	err = services.ValidateStruct(newUser)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	_, err = users.CreateUser(user)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.String(http.StatusOK, "success")
}
