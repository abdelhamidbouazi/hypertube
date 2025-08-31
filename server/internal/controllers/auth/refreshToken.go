package auth

import (
	"net/http"
	"server/internal/models"
	"slices"

	"github.com/labstack/echo/v4"
)

// refresh tokens must be revoked each time - AKA token ROTATION
func RefreshToken(c echo.Context) error {
	token := c.Request().Header.Get("RefreshToken")
	user := c.Get("model").(models.User)
	if !slices.Contains(user.Tokens, token) {
		return echo.ErrUnauthorized
	}
	response, err := RevokeToken(user, token)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, response)
}
