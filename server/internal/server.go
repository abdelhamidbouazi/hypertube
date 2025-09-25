package internal

import (
	"fmt"
	"net/http"
	"server/internal/controllers"
	"server/internal/handlers"
	"server/internal/middlewares"
	"server/internal/routes"
	"server/internal/services"
	"server/internal/services/oauth2"
	"strconv"

	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

var Server *echo.Echo

var (
	movieService       *services.MovieService
	torrentService     *services.TorrentService
	transcodingService *services.TranscodingService
	movieHandler       *handlers.MovieHandler
	commentHandler     *handlers.CommentHandler
)

func InitServices() {
	movieService = services.NewMovieService(
		services.Conf.MOVIE_APIS.TMDB.APIKey,
		services.Conf.MOVIE_APIS.OMDB.APIKey,
	)

	torrentService = services.NewTorrentService(
		services.Conf.STREAMING.DownloadDir,
		services.PostgresDB(),
	)

	transcodingService = services.NewTranscodingService()

	movieHandler = handlers.NewMovieHandler(
		movieService,
		torrentService,
		transcodingService,
		services.PostgresDB(),
	)

	commentHandler = handlers.NewCommentHandler(services.PostgresDB())
}

func Init(config string) {
	services.LoadConfig(config)
	services.LoadTemplateConfig()
	services.LoadMailDialer()
	services.LoadValidator()
	services.LoadDatabase()
	oauth2.LoadConfig()
	services.LoadValidator()
	InitServices()
	LoadServer()

	fmt.Println("mail=", services.Conf.SMTP.Gmail.Mail, " password=", services.Conf.SMTP.Gmail.Password)

	// seeds.AddUsersSeeds()
}

func LoadServer() {
	services.Logger.Debug("Loading Server")

	Server = echo.New()
	Server.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
		AllowOrigins: services.Conf.CORS.Origins,
	}))
	Server.Use(middleware.Logger())
	config := echojwt.Config{
		SigningKey: []byte(services.Conf.JWT.SigningKey),
	}

	middlewares.SetupJWT(config)

	Server.POST("/forgot-password", controllers.ForgotPassword)
	Server.POST("/reset-password", controllers.ResetPassword)
	routes.AddAuthRouter(Server.Group("/auth"))
	routes.AddOAuthRouter(Server.Group("/oauth2"))
	routes.AddUserRouter(Server.Group("/users"))

	movieGroup := Server.Group("/movies")
	movieGroup.GET("/search", movieHandler.SearchMovies)
	movieGroup.GET("/:id", movieHandler.GetMovieDetails, middlewares.Authenticated, middlewares.AttachUser)

	torrentGroup := Server.Group("/torrents")
	torrentGroup.GET("/search", movieHandler.SearchTorrents)
	torrentGroup.POST("/download", movieHandler.DownloadTorrent)
	torrentGroup.GET("/progress", movieHandler.GetTorrentProgress)

	streamGroup := Server.Group("/stream")
	streamGroup.GET("/:id", movieHandler.StreamMovie, middlewares.Authenticated, middlewares.AttachUser)

	commentGroup := Server.Group("/comments")
	commentGroup.POST("/add", commentHandler.AddComment, middlewares.Authenticated, middlewares.AttachUser)
	commentGroup.GET("", commentHandler.GetComments, middlewares.Authenticated, middlewares.AttachUser)
	commentGroup.GET("/:id", commentHandler.GetCommentByID, middlewares.Authenticated, middlewares.AttachUser)
	commentGroup.PATCH("/:id", commentHandler.UpdateComment, middlewares.Authenticated, middlewares.AttachUser)
	commentGroup.DELETE("/:id", commentHandler.DeleteComment, middlewares.Authenticated, middlewares.AttachUser)
}

func StartServer() {
	services.Logger.Info("Starting Server")
	port := strconv.Itoa(services.Conf.HTTP.PORT)
	if err := Server.Start(":" + port); err != http.ErrServerClosed {
		Server.Logger.Fatal(err)
	}
}
