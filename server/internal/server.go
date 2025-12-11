package internal

import (
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
	commentController   *controllers.CommentController
	websocketController *controllers.WebSocketController
)

func InitServices() {
	movieService = services.NewMovieService(
		services.Conf.MOVIE_APIS.TMDB.APIKey,
		services.Conf.MOVIE_APIS.OMDB.APIKey,
		services.Conf.MOVIE_APIS.WATCHMODE.APIKey,
		services.PostgresDB(),
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
}

func LoadServer() {
	services.Logger.Debug("Loading Server")

	Server = echo.New()
	Server.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization, "RefreshToken"},
		AllowOrigins: services.Conf.CORS.Origins,
	}))
	// Server.Use(middleware.Logger())
	config := echojwt.Config{
		SigningKey: []byte(services.Conf.JWT.SigningKey),
	}

	refreshTokenConfig := echojwt.Config{
		SigningKey:  []byte(services.Conf.JWT.SigningKey),
		TokenLookup: "header:RefreshToken",
	}

	accessTokenConfigExtractor := echojwt.Config{
		SigningKey:  []byte(services.Conf.JWT.SigningKey),
		TokenLookup: "header:Authorization:Bearer ",
	}

	middlewares.SetupJWT(config, refreshTokenConfig, accessTokenConfigExtractor)
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
	routes.AddMovieRouter(Server.Group("/movies"), movieController)
	routes.AddCommentRouter(Server.Group("/comments"), commentController)

	streamGroup := Server.Group("/stream", middlewares.Authenticated, middlewares.AttachUser)
	streamGroup.GET("/*", movieController.ServeHLSFile)
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
