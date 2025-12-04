package auth

import (
	"net/http"
	"server/internal/services"
	"server/internal/services/users"

	"github.com/labstack/echo/v4"
)

type RegisterUserType struct {
	FirstName string `validate:"required,min=4" example:"Alan"`
	LastName  string `validate:"required,min=4" example:"Turing"`
	Email     string `validate:"required,email" example:"example@email.com"`
	Password  string `validate:"required,min=8" example:"j8Kt603ql0RV"`
	Username  string `validate:"required" example:"fturing"`
}

// Register godoc
//
//	@Summary		Register
//	@Description	Register new user
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			RegisterUserType	body		RegisterUserType	true	"register credentials to send"
//	@Success		200					{string}	string				"success message"
//	@Failure		400					{object}	utils.HTTPError
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
	user.Username = newUser.Username

	err = services.ValidateStruct(newUser)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	isTaken, err := IsUserNameAlreadyTaken(user.Username)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	if isTaken {
		return echo.NewHTTPError(http.StatusBadRequest, "username is already taken")
	}

	_, err = users.CreateUser(user)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.String(http.StatusOK, "success")
}

func IsUserNameAlreadyTaken(username string) (bool, error) {
	_, err := users.GetUserByUsername(username)
	if err == nil {
		return true, nil
	}
	if err.Error() == "record not found" {
		return false, nil
	}
	return false, err
}
