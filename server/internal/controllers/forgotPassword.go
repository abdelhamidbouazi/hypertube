package controllers

import (
	"bytes"
	"log"
	"net/http"
	"server/internal/services"
	resetpassword "server/internal/services/reset-password"
	"server/internal/services/users"

	"github.com/labstack/echo/v4"
	gomail "gopkg.in/mail.v2"

	"github.com/google/uuid"
)

type ForgotPasswordPayload struct {
	Email string `json:"email" validate:"email" example:"example@email.com"`
}

type TemplateArgs struct {
	Name string
	URL  string
}

type ForgotPasswordRes struct {
	Message string `json:"message" example:"success"`
}

// Forgot Password godoc
//
//	@Summary		Forgot password
//	@Description	Request reset password
//	@Tags			reset-password
//	@Accept			json
//	@Produce		json
//	@Param			ForgotPasswordPayload	body		ForgotPasswordPayload	true	"json body to send to reset password"
//	@Success		200						{object}	ForgotPasswordRes
//	@Failure		401						{object}	utils.HTTPErrorUnauthorized
//	@Failure		400						{object}	utils.HTTPError
//	@Router			/forgot-password [post]
func ForgotPassword(c echo.Context) error {
	var body ForgotPasswordPayload

	err := c.Bind(&body)
	if err != nil {
		return err
	}

	err = services.ValidateStruct(body)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	user, err := users.GetUserByEmail(body.Email)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	token := resetpassword.NewResetPassword{
		Token: uuid.New().String(),
		Email: body.Email,
	}

	err = resetpassword.CreateNew(token)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	msg := gomail.NewMessage()

	msg.SetHeader("From", services.Conf.SMTP.Gmail.Mail)
	msg.SetHeader("To", body.Email)
	msg.SetHeader("Subject", "Reset your password")

	url := services.Conf.UI.Address + "/" +
		services.Conf.UI.ResetPasswordRoute + "?" +
		services.Conf.UI.ResetPasswordTokenQuery + "=" +
		token.Token + "&email=" + user.Email

	b := TemplateArgs{user.FirstName, url}

	t := services.ResetPasswordTemplate()

	var tpl bytes.Buffer
	if err := t.Execute(&tpl, b); err != nil {
		log.Fatal(err)
	}

	msg.SetBody("text/html", tpl.String())

	d := services.MailDialer()

	if err := d.DialAndSend(msg); err != nil {
		log.Fatal(err)
	}

	return c.JSON(http.StatusOK, ForgotPasswordRes{"success"})
}
