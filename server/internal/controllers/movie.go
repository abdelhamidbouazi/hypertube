package controllers

import (
	"fmt"
	"net/http"
	"path/filepath"
	"server/internal/models"
	"server/internal/services"
	"server/internal/utils"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

type MovieController struct {
	movieService        *services.MovieService
	torrentService      *services.TorrentService
	db                  *gorm.DB
	websocketController *WebSocketController
}

func NewMovieController(ms *services.MovieService, ts *services.TorrentService, db *gorm.DB, wsc *WebSocketController) *MovieController {
	return &MovieController{
		movieService:        ms,
		torrentService:      ts,
		db:                  db,
		websocketController: wsc,
	}
}

// GetMovieDetails godoc
//
//	@Summary      Movie details
//	@Description  Get detailed information for a movie by ID
//	@Tags         movies
//	@Accept       json
//	@Produce      json
//	@Param        Authorization header   string     false   "Bearer token"
//	@Param        id       path   string  true   "Movie ID"
//	@Param        source   query  string  false  "Source (default: tmdb)"
//	@Success      200  {object} controllers.MovieDetailsDoc
//	@Failure      401  {object} utils.HTTPErrorUnauthorized
//	@Failure      404  {object} utils.HTTPError
//	@Router       /movies/{id} [get]
func (c *MovieController) GetMovieDetails(ctx echo.Context) error {
	movieID := ctx.Param("id")
	source := ctx.QueryParam("source")

	source = "tmdb"

	s, err := c.movieService.GetSource(source)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}

	details, err := s.GetMovieDetails(movieID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Movie not found")
	}

	if ctx.Get("model") != nil {
		user := ctx.Get("model").(models.User)
		var watchHistory models.WatchHistory
		err = c.db.Model(&models.WatchHistory{}).Where("user_id = ? AND movie_id = ?", user.ID, movieID).First(&watchHistory).Error
		details.IsWatched = err == nil
	}

	c.loadComments(details)

	var downloadedMovie models.DownloadedMovie
	err = c.db.Where("movie_id = ?", details.ID).First(&downloadedMovie).Error
	if err == nil {
		details.IsAvailable = true
		details.StreamURL = fmt.Sprintf("/api/stream/%d", details.ID)
	}

	return ctx.JSON(http.StatusOK, details)
}

func (c *MovieController) loadComments(details *models.MovieDetails) {
	var comments []models.Comment
	err := c.db.Where("movie_id = ?", details.ID).Order("created_at DESC").Find(&comments).Error
	if err != nil {
		return
	}

	details.Comments = comments
}

// SearchMovies godoc
//
//	@Summary      Search movies
//	@Description  Search movies by title with optional year filter
//	@Tags         movies
//	@Accept       json
//	@Produce      json
//	@Param        q     query    string  true   "Search query (title)"
//	@Param        year  query    string  false  "Release year"
//	@Success      200   {array}   models.Movie
//	@Failure      400   {object}  utils.HTTPError
//	@Failure      500   {object}  utils.HTTPError
//	@Router       /movies/search [get]
func (c *MovieController) SearchMovies(ctx echo.Context) error {
	query := ctx.QueryParam("q")
	year := ctx.QueryParam("year")

	if query == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Query parameter 'q' is required")
	}

	movies, err := c.movieService.SearchMovies(query, year)
	if err != nil {
		// log.Printf("Search error: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to search movies")
	}

	return ctx.JSON(http.StatusOK, movies)
}

// GetMovies godoc
//
//	@Summary      Movies
//	@Description  Get movies. Defaults to popular when no filters provided. Supports filters via TMDB discover.
//	@Tags         movies
//	@Accept       json
//	@Produce      json
//	@Param        Authorization header   string     false   "Bearer token"
//	@Param        source        query    string     false  "Source of Movie"
//	@Param        page        query    int     false  "Page number (default 1)"
//	@Param        genres      query    string  false  "Comma-separated genre names or IDs (e.g., Action,Drama or 28,18)"
//	@Param        yearFrom    query    int     false  "Release year from (inclusive)"
//	@Param        yearTo      query    int     false  "Release year to (inclusive)"
//	@Param        minRating   query    number  false  "Minimum TMDB rating (0-10)"
//	@Param        sort        query    string  false  "Sort by: year, year_asc, year_desc, rating (default popularity)"
//	@Success      200  {array}   services.DiscoverMoviesResp
//	@Failure      500  {object}  utils.HTTPError
//	@Router       /movies/popular [get]
func (c *MovieController) GetMovies(ctx echo.Context) error {
	source := ctx.QueryParam("source")
	page := 1
	if p := ctx.QueryParam("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	// Parse filters
	genresParam := ctx.QueryParam("genres")       // comma-separated names or ids
	yearFromParam := ctx.QueryParam("yearFrom")   // int
	yearToParam := ctx.QueryParam("yearTo")       // int
	minRatingParam := ctx.QueryParam("minRating") // float
	sortParam := ctx.QueryParam("sort")           // year|year_asc|year_desc|rating

	var yearFromPtr, yearToPtr *int
	if yearFromParam != "" {
		if v, err := strconv.Atoi(yearFromParam); err == nil && v > 1800 && v < 3000 {
			yearFromPtr = &v
		}
	}
	if yearToParam != "" {
		if v, err := strconv.Atoi(yearToParam); err == nil && v > 1800 && v < 3000 {
			yearToPtr = &v
		}
	}
	var minRatingPtr *float64
	if minRatingParam != "" {
		if v, err := strconv.ParseFloat(minRatingParam, 64); err == nil && v >= 0 && v <= 10 {
			minRatingPtr = &v
		}
	}

	var genres []string
	if genresParam != "" {
		for _, g := range strings.Split(genresParam, ",") {
			if s := strings.TrimSpace(g); s != "" {
				genres = append(genres, s)
			}
		}
	}

	params := services.MovieDiscoverParams{
		Genres:    genres,
		YearFrom:  yearFromPtr,
		YearTo:    yearToPtr,
		MinRating: minRatingPtr,
		Sort:      sortParam,
		Page:      page,
	}

	var userID *uint = nil

	if ctx.Get("model") == nil {
		userID = nil
	} else {
		user := ctx.Get("model").(models.User)
		userID = &user.ID
	}

	movies, err := c.movieService.DiscoverMovies(params, userID, source)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to fetch movies with filters %s", err.Error()))
	}
	return ctx.JSON(http.StatusOK, movies)
}

func validateAndExtractMovieFilePath(path string) (int, error) {
	var movieId int
	var err error

	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) < 2 {
		return movieId, echo.NewHTTPError(http.StatusBadRequest, "Invalid path")
	}

	movieIDStr := parts[1]
	movieId, err = strconv.Atoi(movieIDStr)
	if err != nil {
		return movieId, echo.NewHTTPError(http.StatusBadRequest, "Invalid movie ID")
	}

	return movieId, nil
}

// serveHLSFile godoc
//
// @Summary      Serve HLS video or playlist file
// @Description  Streams HLS (.m3u8 playlist or .ts segment) files for a given movie. Triggers transcoding if not already done.
// @Tags         stream
// @Produce      application/vnd.apple.mpegurl,video/MP2T
// @Param        movieID  path      int     true  "Movie ID"
// @Param        file     path      string  true  "HLS file path (e.g. master.m3u8, 720p/0.ts)"
// @Success      200      {file}    binary  "HLS file"
// @Failure      400      {object}  utils.HTTPError
// @Failure      404      {object}  utils.HTTPError
// @Failure      500      {object}  utils.HTTPError
// @Security     ApiKeyAuth
// @Router       /stream/{movieID}/{file} [get]
func (c *MovieController) ServeHLSFile(ctx echo.Context) error {
	path := ctx.Request().URL.Path

	movieID, err := validateAndExtractMovieFilePath(path)
	if err != nil {
		return err
	}

	outputDir := services.VideoTranscoderConf.Output.Directory
	filePath := filepath.Join(outputDir, strings.TrimPrefix(path, "/stream/"))

	if utils.CheckFileExits(filePath) != nil {
		err = c.movieService.EnsureMovieIsPartiallyDownloadedAndStartedTranscoding(movieID, outputDir)
		if err != nil {
			return err
		}

		if err := utils.WaitForFile(filePath); err != nil {
			return ctx.JSON(http.StatusAccepted, echo.Map{})
		}
	}

	if strings.HasSuffix(filePath, ".ts") {
		if userModel := ctx.Get("model"); userModel != nil {
			user := userModel.(models.User)
			c.movieService.TrackUserSegment(user.ID, movieID)
		}
	}

	return ctx.File(filePath)
}
