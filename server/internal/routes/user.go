package routes

import (
	"server/internal/controllers"
	"server/internal/middlewares"

	"github.com/labstack/echo/v4"
)

func AddUserRouter(usersRouter *echo.Group) {
	usersRouter.GET("", controllers.GetUsers, middlewares.Authenticated, middlewares.AttachUser)
}
