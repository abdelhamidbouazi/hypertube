package middlewares

import (
	"server/internal/services"

	"github.com/golang-jwt/jwt/v5"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
)

var Authenticated echo.MiddlewareFunc

func SetupJWT(config echojwt.Config) {
	Authenticated = func(next echo.HandlerFunc) echo.HandlerFunc {
		return echojwt.WithConfig(config)(func(c echo.Context) error {
			data := c.Get("user").(*jwt.Token).Claims.(jwt.MapClaims)
			c.Set("data", data)
			return next(c)
		})
	}
}

func AttachUser(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		data := c.Get("data").(jwt.MapClaims)
		if data == nil {
			return jwt.ErrTokenNotValidYet
		}
		email := data["email"].(string)
		user, err := services.GetUserByEmail(email)
		if err != nil {
			return echo.ErrBadRequest
		}
		c.Set("model", user)
		// to get a field inside User struct - cast it to models.User
		// u := c.Get("model").(models.User)
		// fmt.Println("model ", u.Email)
		return next(c)
	}
}
