package controllers

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
	"strconv"
	"strings"
	
	"server/internal/models"
	"server/internal/services"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

type MovieController struct {
	movieService       *services.MovieService
	torrentService     *services.TorrentService
	transcodingService *services.TranscodingService
	db                 *gorm.DB
}

func NewMovieController(ms *services.MovieService, ts *services.TorrentService, tcs *services.TranscodingService, db *gorm.DB) *MovieController {
	return &MovieController{
		movieService:       ms,
		torrentService:     ts,
		transcodingService: tcs,
		db:                 db,
	}
}

// MovieDetailsDoc is a doc-only schema for MovieDetails that avoids gorm.Model in nested types
type MovieDetailsDoc struct {
	ID           int                      `json:"id"`
	Title        string                   `json:"title"`
	Overview     string                   `json:"overview"`
	ReleaseDate  string                   `json:"release_date"`
	Runtime      int                      `json:"runtime"`
	PosterPath   string                   `json:"poster_path"`
	BackdropPath string                   `json:"backdrop_path"`
	VoteAverage  float64                  `json:"vote_average"`
	IMDbID       string                   `json:"imdb_id"`
	Language     string                   `json:"original_language,omitempty"`
	IsAvailable  bool                     `json:"is_available"`
	StreamURL    string                   `json:"stream_url"`
	Cast         []models.Cast            `json:"cast"`
	Director     []models.Person          `json:"director"`
	Producer     []models.Person          `json:"producer"`
	Genres       []models.Genre           `json:"genres"`
	Comments     []CommentResponse        `json:"comments"`
}

// GetMovieDetails godoc
//
//	@Summary      Movie details
//	@Description  Get detailed information for a movie by ID
//	@Tags         movies
//	@Accept       json
//	@Produce      json
//	@Param        id   path     string  true  "Movie ID"
//	@Security     JWT
//	@Success      200  {object} controllers.MovieDetailsDoc
//	@Failure      401  {object} utils.HTTPErrorUnauthorized
//	@Failure      404  {object} utils.HTTPError
//	@Router       /movies/{id} [get]
func (c *MovieController) GetMovieDetails(ctx echo.Context) error {
	movieID := ctx.Param("id")

	details, err := c.movieService.GetMovieDetails(movieID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Movie not found")
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

func (c *MovieController) serveVideoFile(ctx echo.Context, filePath string) error {
	return c.transcodingService.TranscodeIfNeeded(filePath, ctx.Response(), ctx.Request())
}

func (c *MovieController) streamPartialFile(ctx echo.Context, dl *models.TorrentDownload) error {
	dl.Mu.RLock()
	videoFile := dl.VideoFile
	dl.Mu.RUnlock()

	if videoFile == nil {
		return echo.NewHTTPError(http.StatusNotFound, "Video file not available")
	}

	reader := videoFile.NewReader()
	defer reader.Close()

	ctx.Response().Header().Set("Content-Type", "video/mp4")
	ctx.Response().Header().Set("Accept-Ranges", "bytes")
	ctx.Response().Header().Set("Connection", "keep-alive")

	http.ServeContent(ctx.Response(), ctx.Request(), "", time.Now(), reader)
	return nil
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
		log.Printf("Search error: %v", err)
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
//	@Param        page        query    int     false  "Page number (default 1)"
//	@Param        genres      query    string  false  "Comma-separated genre names or IDs (e.g., Action,Drama or 28,18)"
//	@Param        yearFrom    query    int     false  "Release year from (inclusive)"
//	@Param        yearTo      query    int     false  "Release year to (inclusive)"
//	@Param        minRating   query    number  false  "Minimum TMDB rating (0-10)"
//	@Param        sort        query    string  false  "Sort by: year, year_asc, year_desc, rating (default popularity)"
//	@Success      200  {array}   models.Movie
//	@Failure      500  {object}  utils.HTTPError
//	@Router       /movies [get]
func (c *MovieController) GetMovies(ctx echo.Context) error {
	page := 1
	if p := ctx.QueryParam("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	// Parse filters
	genresParam := ctx.QueryParam("genres")           // comma-separated names or ids
	yearFromParam := ctx.QueryParam("yearFrom")       // int
	yearToParam := ctx.QueryParam("yearTo")           // int
	minRatingParam := ctx.QueryParam("minRating")     // float
	sortParam := ctx.QueryParam("sort")               // year|year_asc|year_desc|rating

	// If no filters at all, keep existing behavior (popular)
	if genresParam == "" && yearFromParam == "" && yearToParam == "" && minRatingParam == "" && sortParam == "" {
		movies, err := c.movieService.GetPopularMovies(page)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to fetch movies")
		}
		return ctx.JSON(http.StatusOK, movies)
	}

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

	movies, err := c.movieService.DiscoverMovies(params)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to fetch movies with filters")
	}
	return ctx.JSON(http.StatusOK, movies)
}

// PopularMovies godoc
//
//	@Summary      Popular movies
//	@Description  Get a list of popular movies from TMDB
//	@Tags         movies
//	@Accept       json
//	@Produce      json
//	@Success      200   {array}   models.Movie
//	@Failure      400   {object}  utils.HTTPError
//	@Failure      500   {object}  utils.HTTPError
//	@Router       /movies/popular [get]
func (c *MovieController) PopularMovies(ctx echo.Context) error {
	page := 1
	if p := ctx.QueryParam("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}
	movies, err := c.movieService.GetPopularMovies(page)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to fetch popular movies")
	}
	return ctx.JSON(http.StatusOK, movies)
}

// RandomMovies godoc
//
//	@Summary      Random movies
//	@Description  Get a random subset of movies from TMDB discover endpoint
//	@Tags         movies
//	@Accept       json
//	@Produce      json
//	@Success      200        {array}   models.Movie
//	@Failure      400        {object}  utils.HTTPError
//	@Failure      500        {object}  utils.HTTPError
//	@Router       /movies/random [get]
func (c *MovieController) RandomMovies(ctx echo.Context) error {
	page := 1
	limit := 20
	if p := ctx.QueryParam("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}
	if l := ctx.QueryParam("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	movies, err := c.movieService.GetRandomMovies(page, limit)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to fetch random movies")
	}
	return ctx.JSON(http.StatusOK, movies)
}

// StreamMovie godoc
//
//	@Summary      Stream a movie by ID
//	@Description  Streams a movie if downloaded or from an active torrent. Returns 102 if still downloading.
//	@Tags         stream
//	@Accept       */*
//	@Produce      video/mp4
//	@Param        id       path   string  true  "Movie ID"
//	@Param        quality  query  string  false "Desired quality (default: 720p)"
//	@Security     JWT
//	@Success      200  {file}    binary  "Video stream"
//	@Failure      102  {object}  utils.HTTPError  "Still downloading"
//	@Failure      401  {object}  utils.HTTPErrorUnauthorized
//	@Failure      404  {object}  utils.HTTPError
//	@Router       /stream/{id} [get]
func (c *MovieController) StreamMovie(ctx echo.Context) error {
	movieID := ctx.Param("id")
	quality := ctx.QueryParam("quality")
	if quality == "" {
		quality = "720p" // default quality
	}

	var downloadedMovie models.DownloadedMovie
	err := c.db.Where("movie_id = ? AND quality = ?", movieID, quality).First(&downloadedMovie).Error

	if err == nil && downloadedMovie.FilePath != "" {
		if _, err := os.Stat(downloadedMovie.FilePath); err == nil {
			return c.serveVideoFile(ctx, downloadedMovie.FilePath)
		}
	}

	downloadKey := fmt.Sprintf("%s-%s", movieID, quality)
	c.torrentService.Mu.RLock()
	dl, exists := c.torrentService.Downloads[downloadKey]
	c.torrentService.Mu.RUnlock()

	if exists && dl != nil {
		dl.Mu.RLock()
		streamingReady := dl.StreamingReady
		dl.Mu.RUnlock()

		if streamingReady {
			return c.streamPartialFile(ctx, dl)
		} else {
			return echo.NewHTTPError(http.StatusProcessing, "Movie is still downloading, not ready for streaming yet")
		}
	}

	return echo.NewHTTPError(http.StatusNotFound, "Movie not available for streaming")
}

// SearchTorrents handles torrent search requests
func (c *MovieController) SearchTorrents(ctx echo.Context) error {
	title := ctx.QueryParam("title")
	year := ctx.QueryParam("year")

	if title == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Title parameter is required")
	}

	results, err := c.movieService.SearchTorrents(title, year)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to search torrents")
	}

	return ctx.JSON(http.StatusOK, results)
}

// DownloadTorrent handles torrent download requests
func (c *MovieController) DownloadTorrent(ctx echo.Context) error {
	var req struct {
		MovieID int    `json:"movie_id"`
		Magnet  string `json:"magnet"`
		Quality string `json:"quality"`
	}

	if err := ctx.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	download, err := c.torrentService.GetOrStartDownload(req.MovieID, req.Magnet, req.Quality)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to start download: %v", err))
	}

	return ctx.JSON(http.StatusOK, map[string]interface{}{
		"movie_id": download.MovieID,
		"quality":  download.Quality,
		"status":   download.Status,
		"progress": download.Progress,
	})
}

// GetTorrentProgress handles torrent progress requests
func (c *MovieController) GetTorrentProgress(ctx echo.Context) error {
	movieID := ctx.QueryParam("movie_id")
	quality := ctx.QueryParam("quality")

	if movieID == "" || quality == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "movie_id and quality parameters are required")
	}

	downloadKey := fmt.Sprintf("%s-%s", movieID, quality)

	c.torrentService.Mu.RLock()
	download, exists := c.torrentService.Downloads[downloadKey]
	c.torrentService.Mu.RUnlock()

	if !exists {
		return echo.NewHTTPError(http.StatusNotFound, "Download not found")
	}

	return ctx.JSON(http.StatusOK, map[string]interface{}{
		"movie_id":     download.MovieID,
		"quality":      download.Quality,
		"status":       download.Status,
		"progress":     download.Progress,
		"stream_ready": download.StreamReady,
	})
}
