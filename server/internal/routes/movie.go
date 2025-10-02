package routes

import (
	"server/internal/controllers"
	"server/internal/middlewares"

	"github.com/labstack/echo/v4"
)

func AddMovieRouter(movieRouter *echo.Group, movieController *controllers.MovieController) {
	movieRouter.GET("/search", movieController.SearchMovies)
	movieRouter.GET("/:id", movieController.GetMovieDetails, middlewares.Authenticated, middlewares.AttachUser)
}

func AddStreamRouter(streamRouter *echo.Group, movieController *controllers.MovieController) {
	streamRouter.GET("/:id", movieController.StreamMovie, middlewares.Authenticated, middlewares.AttachUser)
}

func AddTorrentRouter(torrentRouter *echo.Group, torrentController *controllers.TorrentController) {
	torrentRouter.GET("/search", torrentController.SearchTorrents, middlewares.Authenticated, middlewares.AttachUser)
	torrentRouter.POST("/download", torrentController.StartDownload, middlewares.Authenticated, middlewares.AttachUser)
	torrentRouter.GET("/progress", torrentController.GetDownloadProgress, middlewares.Authenticated, middlewares.AttachUser)
}

func AddCommentRouter(commentRouter *echo.Group, commentController *controllers.CommentController) {
	commentRouter.POST("/add", commentController.AddComment, middlewares.Authenticated, middlewares.AttachUser)
	commentRouter.GET("", commentController.GetComments, middlewares.Authenticated, middlewares.AttachUser)
	commentRouter.GET("/:id", commentController.GetCommentByID, middlewares.Authenticated, middlewares.AttachUser)
	commentRouter.PATCH("/:id", commentController.UpdateComment, middlewares.Authenticated, middlewares.AttachUser)
	commentRouter.DELETE("/:id", commentController.DeleteComment, middlewares.Authenticated, middlewares.AttachUser)
}

func AddSubtitleRouter(subtitleRouter *echo.Group, subtitleController *controllers.SubtitleController) {
	subtitleRouter.GET("", subtitleController.GetSubtitles, middlewares.Authenticated, middlewares.AttachUser)
}
