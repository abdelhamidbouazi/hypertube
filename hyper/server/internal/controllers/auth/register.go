package auth

import (
	"mime/multipart"
	"net/http"
	"server/internal/services"
	"server/internal/services/users"

	"github.com/labstack/echo/v4"
)

type RegisterUserType struct {
	FirstName string                `form:"firstname" validate:"required,min=4" example:"Alan"`
	LastName  string                `form:"lastname" validate:"required,min=4" example:"Turing"`
	Email     string                `form:"email" validate:"required,email" example:"example@email.com"`
	Password  string                `form:"password" example:"j8Kt603ql0RV"`
	Username  string                `form:"username" validate:"required" example:"fturing"`
	Picture   *multipart.FileHeader `form:"picture"`
}

// Register godoc
//
//	@Summary		Register
//	@Description	Register new user
//	@Tags			auth
//	@Accept			multipart/form-data
//	@Produce		json
//	@Param			firstname	formData	string	true	"First Name"
//	@Param			lastname	formData	string	true	"Last Name"
//	@Param			email		formData	string	true	"Email"
//	@Param			password	formData	string	true	"Password"
//	@Param			username	formData	string	true	"Username"
//	@Param			picture		formData	file	false	"Profile Picture"
//	@Success		200			{string}	string	"success message"
//	@Failure		400			{object}	utils.HTTPError
//	@Router			/auth/register [post]
func Register(c echo.Context) error {
	var newUser RegisterUserType

	err := c.Bind(&newUser)
	if err != nil {
		return err
	}

	var user users.CreateUserType
	user.Provider = ""
	user.Email = newUser.Email
	user.FirstName = newUser.FirstName
	user.LastName = newUser.LastName
	user.Password = newUser.Password
	user.Username = newUser.Username
	if newUser.Picture != nil {
		avatarUrl, err := services.UploadPicture(newUser.Picture, user.Username)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err)
		}
		user.Avatar = avatarUrl
	}

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

	err = services.ValidatePassword(newUser.Password)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
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
