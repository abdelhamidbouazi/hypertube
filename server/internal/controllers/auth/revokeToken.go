package auth

import (
	"server/internal/models"
	"server/internal/services"
	"server/internal/services/users"
	"slices"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

type JwtCustomClaims struct {
	jwt.RegisteredClaims
	ID uint `json:"id"`
}

func RevokeToken(user models.User, userToken string) (echo.Map, error) {
	var ret echo.Map

	expiresIn := time.Now().Add(services.Conf.JWT.AccessTkExpiresAt)
	RefreshExpiresIn := time.Now().Add(services.Conf.JWT.RefreshTkExpiresAt)

	claims := &JwtCustomClaims{
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresIn),
		},
		user.ID,
	}

	rclaims := &JwtCustomClaims{
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(RefreshExpiresIn),
		},
		user.ID,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	rtoken := jwt.NewWithClaims(jwt.SigningMethodHS256, rclaims)

	accessToken, err := token.SignedString([]byte(services.Conf.JWT.SigningKey))
	if err != nil {
		return ret, err
	}

	refreshToken, err := rtoken.SignedString([]byte(services.Conf.JWT.SigningKey))
	if err != nil {
		return ret, err
	}

	// remove old refresh token if exists
	user.Tokens = slices.DeleteFunc(user.Tokens, func(cmp string) bool {
		return cmp == userToken
	})

	// store refresh token in database
	user.Tokens = append(user.Tokens, refreshToken)
	res := users.UpdateUser(user)

	if res != nil {
		return echo.Map{}, res
	}

	return echo.Map{
		"AccessToken":           accessToken,
		"RefreshToken":          refreshToken,
		"TokenType":             "Bearer",
		"ExpiresIn":             expiresIn.Unix(),
		"RefreshTokenExpiresIn": RefreshExpiresIn.Unix(),
	}, nil
}
