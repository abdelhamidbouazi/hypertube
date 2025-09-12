package handlers

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

type MovieHandler struct {
	movieService       *services.MovieService
	torrentService     *services.TorrentService
	transcodingService *services.TranscodingService
	db                 *gorm.DB
}

func NewMovieHandler(ms *services.MovieService, ts *services.TorrentService, tcs *services.TranscodingService, db *gorm.DB) *MovieHandler {
	return &MovieHandler{
		movieService:       ms,
		torrentService:     ts,
		transcodingService: tcs,
		db:                 db,
	}
}

func (h *MovieHandler) GetMovieDetails(c echo.Context) error {
	movieID := c.Param("id")

	details, err := h.movieService.GetMovieDetails(movieID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Movie not found")
	}

	h.loadComments(details)

	var downloadedMovie models.DownloadedMovie
	err = h.db.Where("movie_id = ?", details.ID).First(&downloadedMovie).Error
	if err == nil {
		details.IsAvailable = true
		details.StreamURL = fmt.Sprintf("/api/stream/%d", details.ID)
	}

	return c.JSON(http.StatusOK, details)
}

func (h *MovieHandler) loadComments(details *models.MovieDetails) {
	var comments []models.Comment
	err := h.db.Where("movie_id = ?", details.ID).Order("created_at DESC").Find(&comments).Error
	if err != nil {
		return
	}

	details.Comments = comments
}

func (h *MovieHandler) serveVideoFile(c echo.Context, filePath string) error {
	return h.transcodingService.TranscodeIfNeeded(filePath, c.Response(), c.Request())
}

func (h *MovieHandler) streamPartialFile(c echo.Context, dl *models.TorrentDownload) error {
	dl.Mu.RLock()
	videoFile := dl.VideoFile
	dl.Mu.RUnlock()

	if videoFile == nil {
		return echo.NewHTTPError(http.StatusNotFound, "Video file not available")
	}

	reader := videoFile.NewReader()
	defer reader.Close()

	c.Response().Header().Set("Content-Type", "video/mp4")
	c.Response().Header().Set("Accept-Ranges", "bytes")
	c.Response().Header().Set("Connection", "keep-alive")

	http.ServeContent(c.Response(), c.Request(), "", time.Now(), reader)
	return nil
}

func (h *MovieHandler) SearchMovies(c echo.Context) error {
	query := c.QueryParam("q")
	year := c.QueryParam("year")

	if query == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Query parameter 'q' is required")
	}

	movies, err := h.movieService.SearchMovies(query, year)
	if err != nil {
		log.Printf("Search error: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to search movies")
	}

	return c.JSON(http.StatusOK, movies)
}

func (h *MovieHandler) StreamMovie(c echo.Context) error {
	movieID := c.Param("id")
	quality := c.QueryParam("quality")
	if quality == "" {
		quality = "720p" // default quality
	}

	var downloadedMovie models.DownloadedMovie
	err := h.db.Where("movie_id = ? AND quality = ?", movieID, quality).First(&downloadedMovie).Error

	if err == nil && downloadedMovie.FilePath != "" {
		if _, err := os.Stat(downloadedMovie.FilePath); err == nil {
			return h.serveVideoFile(c, downloadedMovie.FilePath)
		}
	}

	downloadKey := fmt.Sprintf("%s-%s", movieID, quality)
	h.torrentService.Mu.RLock()
	dl, exists := h.torrentService.Downloads[downloadKey]
	h.torrentService.Mu.RUnlock()

	if exists && dl != nil {
		dl.Mu.RLock()
		streamingReady := dl.StreamingReady
		dl.Mu.RUnlock()

		if streamingReady {
			return h.streamPartialFile(c, dl)
		} else {
			return echo.NewHTTPError(http.StatusProcessing, "Movie is still downloading, not ready for streaming yet")
		}
	}

	return echo.NewHTTPError(http.StatusNotFound, "Movie not available for streaming")
}

// SearchTorrents handles torrent search requests
func (h *MovieHandler) SearchTorrents(c echo.Context) error {
	title := c.QueryParam("title")
	year := c.QueryParam("year")

	if title == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Title parameter is required")
	}

	results, err := h.movieService.SearchTorrents(title, year)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to search torrents")
	}

	return c.JSON(http.StatusOK, results)
}

// DownloadTorrent handles torrent download requests
func (h *MovieHandler) DownloadTorrent(c echo.Context) error {
	var req struct {
		MovieID int    `json:"movie_id"`
		Magnet  string `json:"magnet"`
		Quality string `json:"quality"`
	}

	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	download, err := h.torrentService.GetOrStartDownload(req.MovieID, req.Magnet, req.Quality)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to start download: %v", err))
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"movie_id": download.MovieID,
		"quality":  download.Quality,
		"status":   download.Status,
		"progress": download.Progress,
	})
}

// GetTorrentProgress handles torrent progress requests
func (h *MovieHandler) GetTorrentProgress(c echo.Context) error {
	movieID := c.QueryParam("movie_id")
	quality := c.QueryParam("quality")

	if movieID == "" || quality == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "movie_id and quality parameters are required")
	}

	downloadKey := fmt.Sprintf("%s-%s", movieID, quality)

	h.torrentService.Mu.RLock()
	download, exists := h.torrentService.Downloads[downloadKey]
	h.torrentService.Mu.RUnlock()

	if !exists {
		return echo.NewHTTPError(http.StatusNotFound, "Download not found")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"movie_id":     download.MovieID,
		"quality":      download.Quality,
		"status":       download.Status,
		"progress":     download.Progress,
		"stream_ready": download.StreamReady,
	})
}
