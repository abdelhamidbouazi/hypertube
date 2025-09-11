package internal

import (
	"fmt"
	"net/http"
	"server/internal/controllers"
	"server/internal/middlewares"
	"server/internal/models"
	"server/internal/routes"
	"server/internal/services"
	"server/internal/services/oauth2"
	"strconv"
	"time"

	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

var Server *echo.Echo

func Init(config string) {
	services.LoadConfig(config)
	services.LoadTemplateConfig()
	services.LoadMailDialer()
	services.LoadValidator()
	services.LoadDatabase()
	oauth2.LoadConfig()
	services.LoadValidator()
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

	movieService := services.NewMovieService(
		services.Conf.MOVIE_APIS.TMDB.APIKey,
		services.Conf.MOVIE_APIS.OMDB.APIKey,
	)

	torrentService := services.NewTorrentService(
		services.Conf.STREAMING.DownloadDir,
		services.PostgresDB(),
	)

	transcodingService := services.NewTranscodingService()

	movieGroup := Server.Group("/movies")
	movieGroup.GET("/search", func(c echo.Context) error {
		query := c.QueryParam("q")
		year := c.QueryParam("year")

		if query == "" {
			return echo.NewHTTPError(400, "Query parameter 'q' is required")
		}

		movies, err := movieService.SearchMovies(query, year)
		if err != nil {
			return echo.NewHTTPError(500, "Failed to search movies")
		}

		return c.JSON(200, movies)
	})

	movieGroup.GET("/:id", func(c echo.Context) error {
		movieID := c.Param("id")

		details, err := movieService.GetMovieDetails(movieID)
		if err != nil {
			return echo.NewHTTPError(404, "Movie not found")
		}

		return c.JSON(200, details)
	}, middlewares.Authenticated, middlewares.AttachUser)

	torrentGroup := Server.Group("/torrents")
	torrentGroup.GET("/search", func(c echo.Context) error {
		title := c.QueryParam("title")
		year := c.QueryParam("year")

		if title == "" {
			return echo.NewHTTPError(400, "Title parameter is required")
		}

		results, err := movieService.SearchTorrents(title, year)
		if err != nil {
			return echo.NewHTTPError(500, "Failed to search torrents")
		}

		return c.JSON(200, results)
	})

	torrentGroup.POST("/download", func(c echo.Context) error {
		var req struct {
			MovieID int    `json:"movie_id"`
			Magnet  string `json:"magnet"`
			Quality string `json:"quality"`
		}

		if err := c.Bind(&req); err != nil {
			return echo.NewHTTPError(400, "Invalid request body")
		}

		download, err := torrentService.GetOrStartDownload(req.MovieID, req.Magnet, req.Quality)
		if err != nil {
			return echo.NewHTTPError(500, fmt.Sprintf("Failed to start download: %v", err))
		}

		return c.JSON(200, map[string]interface{}{
			"movie_id": download.MovieID,
			"quality":  download.Quality,
			"status":   download.Status,
			"progress": download.Progress,
		})
	})

	torrentGroup.GET("/progress", func(c echo.Context) error {
		movieID := c.QueryParam("movie_id")
		quality := c.QueryParam("quality")

		if movieID == "" || quality == "" {
			return echo.NewHTTPError(400, "movie_id and quality parameters are required")
		}

		downloadKey := fmt.Sprintf("%s-%s", movieID, quality)

		torrentService.Mu.RLock()
		download, exists := torrentService.Downloads[downloadKey]
		torrentService.Mu.RUnlock()

		if !exists {
			return echo.NewHTTPError(404, "Download not found")
		}

		return c.JSON(200, map[string]interface{}{
			"movie_id":     download.MovieID,
			"quality":      download.Quality,
			"status":       download.Status,
			"progress":     download.Progress,
			"stream_ready": download.StreamReady,
		})
	})

	streamGroup := Server.Group("/stream")
	streamGroup.GET("/:id", func(c echo.Context) error {
		movieID := c.Param("id")
		quality := c.QueryParam("quality")
		if quality == "" {
			quality = "720p"
		}

		var downloadedMovie models.DownloadedMovie
		err := services.PostgresDB().Where("movie_id = ? AND quality = ?", movieID, quality).First(&downloadedMovie).Error
		if err != nil {
			return echo.NewHTTPError(404, "Movie not found or not downloaded")
		}

		services.PostgresDB().Model(&downloadedMovie).Update("last_watched", time.Now())

		if err := transcodingService.TranscodeIfNeeded(downloadedMovie.FilePath, c.Response().Writer, c.Request()); err != nil {
			return echo.NewHTTPError(500, fmt.Sprintf("Failed to stream movie: %v", err))
		}

		return nil
	}, middlewares.Authenticated, middlewares.AttachUser)
}

func StartServer() {
	services.Logger.Info("Starting Server")
	port := strconv.Itoa(services.Conf.HTTP.PORT)
	if err := Server.Start(":" + port); err != http.ErrServerClosed {
		Server.Logger.Fatal(err)
	}
}
