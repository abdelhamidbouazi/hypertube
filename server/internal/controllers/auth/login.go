package auth

import (
	"net/http"
	"server/internal/services"
	"server/internal/services/users"

	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

type LoginUserType struct {
	// Email    string `validate:"required" example:"example@email.com"`
	Username string `validate:"required" example:"fturing"`
	Password string `validate:"required" example:"j8Kt603ql0RV"`
}

// Login godoc
//
//	@Summary		Login
//	@Description	login using email and password
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			loginUserRequest	body		LoginUserType	true	"json body to send with username and password"
//	@Success		200					{object}	RevokeTokenRes
//	@Failure		401					{object}	utils.HTTPErrorUnauthorized
//	@Failure		400					{object}	utils.HTTPError
//	@Router			/auth/login [post]
func Login(c echo.Context) error {
	var newUser LoginUserType

	err := c.Bind(&newUser)
	if err != nil {
		return echo.ErrBadRequest
	}

	err = services.ValidateStruct(newUser)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	user, err := users.GetUserByUsername(newUser.Username)
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
