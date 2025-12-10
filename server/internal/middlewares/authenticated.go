package middlewares

import (
	"server/internal/services/users"

	"github.com/golang-jwt/jwt/v5"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
)

var (
	Authenticated         echo.MiddlewareFunc
	RefreshTokenExtractor echo.MiddlewareFunc
	AccessTokenExtractor  echo.MiddlewareFunc
)

func SetupJWT(config, refreshTokenConfig, accessTokenConfig echojwt.Config) {
	Authenticated = func(next echo.HandlerFunc) echo.HandlerFunc {
		return echojwt.WithConfig(config)(func(c echo.Context) error {
			data := c.Get("user").(*jwt.Token).Claims.(jwt.MapClaims)
			c.Set("data", data)
			return next(c)
		})
	}

	RefreshTokenExtractor = func(next echo.HandlerFunc) echo.HandlerFunc {
		return echojwt.WithConfig(refreshTokenConfig)(func(c echo.Context) error {
			data := c.Get("user").(*jwt.Token).Claims.(jwt.MapClaims)
			c.Set("data", data)
			return next(c)
		})
	}

	AccessTokenExtractor = func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Try to extract token using JWT middleware
			err := echojwt.WithConfig(accessTokenConfig)(func(c echo.Context) error {
				if c.Get("user") == nil {
					c.Set("data", nil)
				} else {
					data := c.Get("user").(*jwt.Token).Claims.(jwt.MapClaims)
					c.Set("data", data)
				}
				return next(c)
			})(c)

			// If JWT middleware returns error (invalid/missing token), continue without data
			if err != nil {
				c.Set("data", nil)
				return next(c)
			}

			return err
		}
	}
}

func AttachUserOptional(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		dataValue := c.Get("data")
		if dataValue == nil {
			c.Set("model", nil)
			return next(c)
		}

		data, ok := dataValue.(jwt.MapClaims)
		if !ok || data == nil {
			c.Set("model", nil)
			return next(c)
		}

		id := data["id"].(float64)
		user, err := users.GetUserById(id)
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

func AttachUser(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		data := c.Get("data").(jwt.MapClaims)
		if data == nil {
			return jwt.ErrTokenNotValidYet
		}
		id := data["id"].(float64)
		user, err := users.GetUserById(id)
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
