package routes

import (
	"server/internal/controllers"
	"server/internal/middlewares"

	"github.com/labstack/echo/v4"
)

func AddMovieRouter(movieRouter *echo.Group, movieController *controllers.MovieController) {
	movieRouter.GET("/popular", movieController.GetMovies, middlewares.AccessTokenExtractor, middlewares.AttachUserOptional)
	movieRouter.GET("/search", movieController.SearchMovies)
	movieRouter.GET("/:id", movieController.GetMovieDetails, middlewares.AccessTokenExtractor, middlewares.AttachUserOptional)
	movieRouter.GET("/:id/:source", movieController.GetMovieDetails, middlewares.AccessTokenExtractor, middlewares.AttachUserOptional)
}

func AddCommentRouter(commentRouter *echo.Group, commentController *controllers.CommentController) {
	commentRouter.POST("/add", commentController.AddComment, middlewares.Authenticated, middlewares.AttachUser)
	commentRouter.GET("", commentController.GetComments, middlewares.Authenticated, middlewares.AttachUser)
	commentRouter.GET("/:id", commentController.GetCommentByID, middlewares.Authenticated, middlewares.AttachUser)
	commentRouter.PATCH("/:id", commentController.UpdateComment, middlewares.Authenticated, middlewares.AttachUser)
	commentRouter.DELETE("/:id", commentController.DeleteComment, middlewares.Authenticated, middlewares.AttachUser)
}
