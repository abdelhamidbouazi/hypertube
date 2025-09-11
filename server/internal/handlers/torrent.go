package handlers

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"server/internal/services"
)

type TorrentHandler struct {
	torrentService *services.TorrentService
	movieService   *services.MovieService
}

func NewTorrentHandler(ts *services.TorrentService, ms *services.MovieService) *TorrentHandler {
	return &TorrentHandler{
		torrentService: ts,
		movieService:   ms,
	}
}

func (h *TorrentHandler) SearchTorrents(c echo.Context) error {
	title := c.QueryParam("title")
	year := c.QueryParam("year")

	results, err := h.movieService.SearchTorrents(title, year)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to search torrents")
	}

	return c.JSON(http.StatusOK, results)
}

func (h *TorrentHandler) StartDownload(c echo.Context) error {
	var req struct {
		MovieID int    `json:"movie_id"`
		Magnet  string `json:"magnet"`
		Quality string `json:"quality"`
	}

	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request")
	}

	download, err := h.torrentService.GetOrStartDownload(req.MovieID, req.Magnet, req.Quality)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to start download")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"status":       download.Status,
		"progress":     download.Progress,
		"stream_ready": download.StreamReady,
	})
}

func (h *TorrentHandler) GetDownloadProgress(c echo.Context) error {
	movieID := c.QueryParam("movie_id")
	quality := c.QueryParam("quality")

	if movieID == "" || quality == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Movie ID and quality are required")
	}

	downloadKey := fmt.Sprintf("%s-%s", movieID, quality)

	h.torrentService.Mu.RLock()
	dl, exists := h.torrentService.Downloads[downloadKey]
	h.torrentService.Mu.RUnlock()

	if !exists {
		return echo.NewHTTPError(http.StatusNotFound, "Download not found")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"status":       dl.Status,
		"progress":     dl.Progress,
		"stream_ready": dl.StreamReady,
		"quality":      dl.Quality,
	})
}
