package auth

import (
	"net/http"
	"server/internal/models"
	"server/internal/services"
	"slices"

	"github.com/labstack/echo/v4"
)

func Logout(c echo.Context) error {
	token := c.Request().Header.Get("RefreshToken")
	user := c.Get("model").(models.User)
	if !slices.Contains(user.Tokens, token) {
		return echo.ErrUnauthorized
	}

	user.Tokens = slices.DeleteFunc(user.Tokens, func(cmp string) bool {
		return cmp == token
	})

	err := services.UpdateUser(user)
	if err != nil {
		return echo.ErrInternalServerError
	}

	return c.String(http.StatusOK, "success")
}
