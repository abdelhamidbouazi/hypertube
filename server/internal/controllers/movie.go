package controllers

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

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
//  @Summary      Movies
//  @Description  Get a default list of movies
//  @Tags         movies
//  @Accept       json
//  @Produce      json
//  @Success      200  {array}   models.Movie
//  @Failure      500  {object}  utils.HTTPError
//  @Router       /movies [get]
func (c *MovieController) GetMovies(ctx echo.Context) error {
	movies, err := c.movieService.GetMovies()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to fetch movies")
	}
	return ctx.JSON(http.StatusOK, movies)
}

// PopularMovies godoc
//
//  @Summary      Popular movies
//  @Description  Get a list of popular movies from TMDB
//  @Tags         movies
//  @Accept       json
//  @Produce      json
//  @Success      200   {array}   models.Movie
//  @Failure      400   {object}  utils.HTTPError
//  @Failure      500   {object}  utils.HTTPError
//  @Router       /movies/popular [get]
func (c *MovieController) PopularMovies(ctx echo.Context) error {
	movies, err := c.movieService.GetPopularMovies(1)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to fetch popular movies")
	}
	return ctx.JSON(http.StatusOK, movies)
}

// RandomMovies godoc
//
//  @Summary      Random movies
//  @Description  Get a random subset of movies from TMDB discover endpoint
//  @Tags         movies
//  @Accept       json
//  @Produce      json
//  @Success      200        {array}   models.Movie
//  @Failure      400        {object}  utils.HTTPError
//  @Failure      500        {object}  utils.HTTPError
//  @Router       /movies/random [get]
func (c *MovieController) RandomMovies(ctx echo.Context) error {
	movies, err := c.movieService.GetRandomMovies(1, 20)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to fetch random movies")
	}
	return ctx.JSON(http.StatusOK, movies)
}

// StreamMovie godoc
//
//  @Summary      Stream a movie by ID
//  @Description  Streams a movie if downloaded or from an active torrent. Returns 102 if still downloading.
//  @Tags         stream
//  @Accept       */*
//  @Produce      video/mp4
//  @Param        id       path   string  true  "Movie ID"
//  @Param        quality  query  string  false "Desired quality (default: 720p)"
//  @Security     JWT
//  @Success      200  {file}    binary  "Video stream"
//  @Failure      102  {object}  utils.HTTPError  "Still downloading"
//  @Failure      401  {object}  utils.HTTPErrorUnauthorized
//  @Failure      404  {object}  utils.HTTPError
//  @Router       /stream/{id} [get]
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
