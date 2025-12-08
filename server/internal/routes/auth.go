package routes

import (
	"server/internal/controllers/auth"
	"server/internal/middlewares"

	"github.com/labstack/echo/v4"
)

func AddAuthRouter(authRouter *echo.Group) {
	authRouter.POST("/register", auth.Register)
	authRouter.POST("/login", auth.Login)
	authRouter.POST("/refreshToken", auth.RefreshToken, middlewares.RefreshTokenExtractor, middlewares.AttachUser)
	authRouter.DELETE("/logout", auth.Logout, middlewares.Authenticated, middlewares.AttachUser)
}
