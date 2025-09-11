package auth

import (
	"net/http"
	"server/internal/models"
	"server/internal/services/users"
	"slices"

	"github.com/labstack/echo/v4"
)

// Logout godoc
//
//	@Summary		Logout
//	@Description	Logout
//	@Tags			auth
//	@Security		JWT
//	@Accept			json
//	@Produce		plain
//	@Param			refreshToken	header		string	true	"refresh token that was sent on login or refresh token"
//	@Success		200				{string}	string	"success"
//	@Router			/auth/logout [delete]
func Logout(c echo.Context) error {
	token := c.Request().Header.Get("RefreshToken")
	user := c.Get("model").(models.User)
	if !slices.Contains(user.Tokens, token) {
		return echo.ErrUnauthorized
	}

	user.Tokens = slices.DeleteFunc(user.Tokens, func(cmp string) bool {
		return cmp == token
	})

	err := users.UpdateUser(user)
	if err != nil {
		return echo.ErrInternalServerError
	}

	return c.String(http.StatusOK, "success")
}
