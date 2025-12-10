package auth

import (
	"net/http"
	"server/internal/models"
	"slices"

	"github.com/labstack/echo/v4"
)

// RefreshToken godoc
//
//	@Summary		RefreshToken
//	@Description	Revoke old token if expires and get new refresh token and access token
//	@Tags			auth
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			RefreshToken	header		string	true	"old refresh token to renew"
//	@Success		200				{object}	RevokeTokenRes
//	@Failure		401				{object}	utils.HTTPErrorUnauthorized
//	@Router			/auth/refreshToken [post]
func RefreshToken(c echo.Context) error {
	// refresh tokens must be revoked each time - AKA token ROTATION
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
