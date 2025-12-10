package internal

import (
	"fmt"
	"net/http"
	"server/internal/controllers"
	"server/internal/middlewares"
	"server/internal/routes"
	"server/internal/services"
	"server/internal/services/oauth2"
	"strconv"

	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	echoSwagger "github.com/swaggo/echo-swagger"
)

var Server *echo.Echo

var (
	movieService        *services.MovieService
	torrentService      *services.TorrentService
	movieController     *controllers.MovieController
	torrentController   *controllers.TorrentController
	subtitleController  *controllers.SubtitleController
	commentController   *controllers.CommentController
	websocketController *controllers.WebSocketController
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

	websocketController = controllers.NewWebSocketController()

	movieController = controllers.NewMovieController(
		movieService,
		torrentService,
		services.PostgresDB(),
		websocketController,
	)

	torrentController = controllers.NewTorrentController(
		torrentService,
		movieService,
	)

	subtitleController = controllers.NewSubtitleController(services.PostgresDB())

	commentController = controllers.NewCommentController(services.PostgresDB())
}

func Init(config string) {
	services.LoadConfig(config)
	services.LoadVideoTranscoderConfig("video_transcoder.yml")
	services.LoadTemplateConfig()
	services.LoadMailDialer()
	services.LoadValidator()
	services.LoadAWSBucket()
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
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization, "RefreshToken"},
		AllowOrigins: services.Conf.CORS.Origins,
	}))
	Server.Use(middleware.Logger())
	config := echojwt.Config{
		SigningKey: []byte(services.Conf.JWT.SigningKey),
	}

	refreshTokenConfig := echojwt.Config{
		SigningKey:  []byte(services.Conf.JWT.SigningKey),
		TokenLookup: "header:RefreshToken",
	}

	middlewares.SetupJWT(config, refreshTokenConfig)
	setupSwagger(Server)

	Server.POST("/forgot-password", controllers.ForgotPassword)
	Server.POST("/reset-password", controllers.ResetPassword)
	Server.GET("/healthcheck", func(c echo.Context) error {
		return c.JSON(http.StatusOK, echo.Map{
			"message": "success",
		})
	})
	routes.AddAuthRouter(Server.Group("/auth"))
	routes.AddOAuthRouter(Server.Group("/oauth2"))
	routes.AddUserRouter(Server.Group("/users"))

	movieGroup := Server.Group("/movies")
	movieGroup.GET("/search", movieController.SearchMovies)
	movieGroup.GET("/:id", movieController.GetMovieDetails, middlewares.Authenticated, middlewares.AttachUser)
	routes.AddMovieRouter(movieGroup, movieController)

	torrentGroup := Server.Group("/torrents")
	torrentGroup.POST("/download", torrentController.StartDownload)
	torrentGroup.GET("/progress", torrentController.GetDownloadProgress)

	streamGroup := Server.Group("/stream", middlewares.Authenticated, middlewares.AttachUser)
	streamGroup.GET("/*", movieController.ServeHLSFile)

	commentGroup := Server.Group("/comments")
	commentGroup.POST("/add", commentController.AddComment, middlewares.Authenticated, middlewares.AttachUser)
	commentGroup.GET("", commentController.GetComments, middlewares.Authenticated, middlewares.AttachUser)
	commentGroup.GET("/:id", commentController.GetCommentByID, middlewares.Authenticated, middlewares.AttachUser)
	commentGroup.PATCH("/:id", commentController.UpdateComment, middlewares.Authenticated, middlewares.AttachUser)
	commentGroup.DELETE("/:id", commentController.DeleteComment, middlewares.Authenticated, middlewares.AttachUser)

	subtitleGroup := Server.Group("/subtitles")
	subtitleGroup.GET("", subtitleController.GetSubtitles, middlewares.Authenticated, middlewares.AttachUser)
	subtitleGroup.GET("/languages", subtitleController.GetAvailableLanguages, middlewares.Authenticated, middlewares.AttachUser)
	subtitleGroup.GET("/recommendations", subtitleController.GetSubtitleRecommendations, middlewares.Authenticated, middlewares.AttachUser)

	Server.GET("/ws/:movieId", websocketController.HandleWebSocket)
}

func setupSwagger(s *echo.Echo) {
	s.GET("/swagger/*", echoSwagger.WrapHandler)
}

func StartServer() {
	services.Logger.Info("Starting Server")
	port := strconv.Itoa(services.Conf.HTTP.PORT)
	if err := Server.Start(":" + port); err != http.ErrServerClosed {
		Server.Logger.Fatal(err)
	}
}
