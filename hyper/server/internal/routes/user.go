package routes

import (
	"server/internal/controllers"
	"server/internal/middlewares"

	"github.com/labstack/echo/v4"
)

func AddUserRouter(usersRouter *echo.Group) {
	usersRouter.GET("", controllers.GetUsers, middlewares.Authenticated, middlewares.AttachUser)

	usersRouter.PATCH("", controllers.UpdateUser, middlewares.Authenticated, middlewares.AttachUser)

	usersRouter.GET("/me", controllers.GetMe, middlewares.Authenticated, middlewares.AttachUser)

	usersRouter.POST("/me/upload-avatar", controllers.UploadPicture, middlewares.Authenticated, middlewares.AttachUser)
	usersRouter.PATCH("/me", controllers.UpdateUser, middlewares.Authenticated, middlewares.AttachUser)
	usersRouter.GET("/stats", controllers.GetUserStats, middlewares.Authenticated, middlewares.AttachUser)

	usersRouter.GET("/watch-history", controllers.GetUserWatchHistory, middlewares.Authenticated, middlewares.AttachUser)

	usersRouter.GET("/:username", controllers.GetUserByUsername, middlewares.Authenticated, middlewares.AttachUser)
}
