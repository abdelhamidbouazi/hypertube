package controllers

import (
	"fmt"
	"net/http"

	"server/internal/services"

	"github.com/labstack/echo/v4"
)

type TorrentController struct {
	torrentService *services.TorrentService
	movieService   *services.MovieService
}

func NewTorrentController(ts *services.TorrentService, ms *services.MovieService) *TorrentController {
	return &TorrentController{
		torrentService: ts,
		movieService:   ms,
	}
}

func (c *TorrentController) SearchTorrents(ctx echo.Context) error {
	title := ctx.QueryParam("title")
	year := ctx.QueryParam("year")

	results, err := c.movieService.SearchTorrents(title, year)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to search torrents")
	}

	return ctx.JSON(http.StatusOK, results)
}

func (c *TorrentController) StartDownload(ctx echo.Context) error {
	var req struct {
		MovieID int    `json:"movie_id"`
		Magnet  string `json:"magnet"`
		Quality string `json:"quality"`
	}

	if err := ctx.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request")
	}

	download, err := c.torrentService.GetOrStartDownload(req.MovieID, req.Magnet, req.Quality)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to start download")
	}

	return ctx.JSON(http.StatusOK, map[string]interface{}{
		"status":       download.Status,
		"progress":     download.Progress,
		"stream_ready": download.StreamReady,
	})
}

func (c *TorrentController) GetDownloadProgress(ctx echo.Context) error {
	movieID := ctx.QueryParam("movie_id")
	quality := ctx.QueryParam("quality")

	if movieID == "" || quality == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Movie ID and quality are required")
	}

	downloadKey := fmt.Sprintf("%s-%s", movieID, quality)

	c.torrentService.Mu.RLock()
	dl, exists := c.torrentService.Downloads[downloadKey]
	c.torrentService.Mu.RUnlock()

	if !exists {
		return echo.NewHTTPError(http.StatusNotFound, "Download not found")
	}

	return ctx.JSON(http.StatusOK, map[string]interface{}{
		"status":       dl.Status,
		"progress":     dl.Progress,
		"stream_ready": dl.StreamReady,
		"quality":      dl.Quality,
	})
}
