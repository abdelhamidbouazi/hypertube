package controllers

import (
	"net/http"
	"server/internal/services"
	resetpassword "server/internal/services/reset-password"

	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

type ResetPasswordPayload struct {
	Token       string `json:"token" validate:"required,uuid"`
	Email       string `json:"email" validate:"required,email"`
	NewPassword string `json:"password" validate:"required,min=8"`
}

type ResetPasswordRes struct {
	Message string `json:"message" example:"success"`
}

// Reset Password godoc
//
//	@Summary		Reset password
//	@Description	Renew old password
//	@Tags			reset-password
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			ResetPasswordPayload	body		ResetPasswordPayload	true	"json body to send to renew old password"
//	@Success		200						{object}	ResetPasswordRes
//	@Router			/reset-password [post]
func ResetPassword(c echo.Context) error {
	var body ResetPasswordPayload

	err := c.Bind(&body)
	if err != nil {
		return echo.ErrBadRequest
	}

	err = services.ValidateStruct(body)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	count, err := resetpassword.CountTokensByEmail(body.Token, body.Email)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}
	if count == 0 {
		return echo.NewHTTPError(http.StatusBadRequest, echo.Map{
			"message": "please recheck the url we send you over email or re-reset your password",
		})
	}

	err = resetpassword.DeleteByEmail(body.Email)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	user, err := services.GetUserByEmail(body.Email)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(body.NewPassword), 8)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	user.Password = string(hashed)
	err = services.UpdateUser(user)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusOK, ResetPasswordRes{"success"})
}
