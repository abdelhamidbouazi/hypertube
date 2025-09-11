package handlers

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"server/internal/models"
	"server/internal/services"

	"github.com/labstack/echo/v4"
)

type MovieHandler struct {
	movieService       *services.MovieService
	torrentService     *services.TorrentService
	transcodingService *services.TranscodingService
	db                 *sql.DB
}

func NewMovieHandler(ms *services.MovieService, ts *services.TorrentService, tcs *services.TranscodingService, db *sql.DB) *MovieHandler {
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

	var filePath string
	err = h.db.QueryRow("SELECT file_path FROM downloaded_movies WHERE movie_id = ?",
		details.ID).Scan(&filePath)
	if err == nil {
		details.IsAvailable = true
		details.StreamURL = fmt.Sprintf("/api/stream/%d", details.ID)
	}

	return c.JSON(http.StatusOK, details)
}

func (h *MovieHandler) loadComments(details *models.MovieDetails) {
	rows, err := h.db.Query("SELECT id, user_id, username, content, created_at FROM comments WHERE movie_id = ? ORDER BY created_at DESC", details.ID)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		var comment models.Comment
		rows.Scan(&comment.ID, &comment.UserID, &comment.Username, &comment.Content, &comment.CreatedAt)
		comment.MovieID = details.ID
		details.Comments = append(details.Comments, comment)
	}
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

	var filePath string
	err := h.db.QueryRow("SELECT file_path FROM downloaded_movies WHERE movie_id = ? AND quality = ?",
		movieID, quality).Scan(&filePath)

	if err == nil && filePath != "" {
		if _, err := os.Stat(filePath); err == nil {
			return h.serveVideoFile(c, filePath)
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
