package routes

import (
	"server/internal/handlers"
	"server/internal/middlewares"

	"github.com/labstack/echo/v4"
)

func AddMovieRouter(movieRouter *echo.Group, movieHandler *handlers.MovieHandler) {
	movieRouter.GET("/search", movieHandler.SearchMovies)
	movieRouter.GET("/:id", movieHandler.GetMovieDetails, middlewares.Authenticated, middlewares.AttachUser)
}

func AddStreamRouter(streamRouter *echo.Group, movieHandler *handlers.MovieHandler) {
	streamRouter.GET("/:id", movieHandler.StreamMovie, middlewares.Authenticated, middlewares.AttachUser)
}

func AddTorrentRouter(torrentRouter *echo.Group, torrentHandler *handlers.TorrentHandler) {
	torrentRouter.GET("/search", torrentHandler.SearchTorrents, middlewares.Authenticated, middlewares.AttachUser)
	torrentRouter.POST("/download", torrentHandler.StartDownload, middlewares.Authenticated, middlewares.AttachUser)
	torrentRouter.GET("/progress", torrentHandler.GetDownloadProgress, middlewares.Authenticated, middlewares.AttachUser)
}

func AddCommentRouter(commentRouter *echo.Group, commentHandler *handlers.CommentHandler) {
	commentRouter.POST("/add", commentHandler.AddComment, middlewares.Authenticated, middlewares.AttachUser)
	commentRouter.GET("", commentHandler.GetComments, middlewares.Authenticated, middlewares.AttachUser)
	commentRouter.GET("/:id", commentHandler.GetCommentByID, middlewares.Authenticated, middlewares.AttachUser)
	commentRouter.PATCH("/:id", commentHandler.UpdateComment, middlewares.Authenticated, middlewares.AttachUser)
	commentRouter.DELETE("/:id", commentHandler.DeleteComment, middlewares.Authenticated, middlewares.AttachUser)
}

func AddSubtitleRouter(subtitleRouter *echo.Group, subtitleHandler *handlers.SubtitleHandler) {
	subtitleRouter.GET("", subtitleHandler.GetSubtitles, middlewares.Authenticated, middlewares.AttachUser)
}
