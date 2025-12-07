package controllers

import (
	"fmt"
	"net/http"

	"server/internal/models"
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

// StartDownloadRequest represents the request body for starting a torrent download
type StartDownloadRequest struct {
	MovieID int    `json:"movie_id" example:"12345"`
	Magnet  string `json:"magnet" example:"magnet:?xt=urn:btih:..."`
	Quality string `json:"quality" example:"720p"`
}

// StartDownloadResponse represents the response after starting a download
type StartDownloadResponse struct {
	Status      string  `json:"status" example:"downloading"`
	Progress    float64 `json:"progress" example:"12.5"`
	StreamReady bool    `json:"stream_ready" example:"false"`
}

// StartDownload godoc
//
//	@Summary      Start torrent download
//	@Description  Starts a torrent download for a movie using a magnet link and quality
//	@Tags         torrents
//	@Accept       json
//	@Produce      json
//	@Param        body  body      StartDownloadRequest  true  "Download request"
//	@Security     JWT
//	@Success      200  {object}  StartDownloadResponse
//	@Failure      400  {object}  utils.HTTPError
//	@Failure      401  {object}  utils.HTTPErrorUnauthorized
//	@Failure      500  {object}  utils.HTTPError
//	@Router       /torrents/download [post]
func (c *TorrentController) StartDownload(ctx echo.Context) error {
	var req StartDownloadRequest

	if err := ctx.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request")
	}

	download, err := c.torrentService.GetOrStartDownload(req.MovieID, req.Magnet, req.Quality)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to start download")
	}

	return ctx.JSON(http.StatusOK, StartDownloadResponse{
		Status:      download.Status,
		Progress:    download.Progress,
		StreamReady: download.StreamReady,
	})
}

// DownloadProgressResponse represents the torrent progress payload
type DownloadProgressResponse struct {
	Status      string  `json:"status" example:"downloading"`
	Progress    float64 `json:"progress" example:"42.0"`
	StreamReady bool    `json:"stream_ready" example:"true"`
	Quality     string  `json:"quality" example:"720p"`
}

// GetDownloadProgress godoc
//
//	@Summary      Get torrent download progress
//	@Description  Returns current status and progress for a movie download by movie_id and quality
//	@Tags         torrents
//	@Accept       json
//	@Produce      json
//	@Param        movie_id  query  string  true  "Movie ID"
//	@Param        quality   query  string  true  "Quality"
//	@Security     JWT
//	@Success      200  {object}  DownloadProgressResponse
//	@Failure      400  {object}  utils.HTTPError
//	@Failure      401  {object}  utils.HTTPErrorUnauthorized
//	@Failure      404  {object}  utils.HTTPError
//	@Router       /torrents/progress [get]
func (c *TorrentController) GetDownloadProgress(ctx echo.Context) error {
	movieID := ctx.QueryParam("movie_id")
	quality := ctx.QueryParam("quality")

	if movieID == "" || quality == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Movie ID and quality are required")
	}

	downloadKey := fmt.Sprintf("%s-%s", movieID, quality)

	val, exists := c.torrentService.Downloads.Load(downloadKey)
	if !exists {
		return echo.NewHTTPError(http.StatusNotFound, "Download not found")
	}
	dl := val.(*models.TorrentDownload)

	return ctx.JSON(http.StatusOK, DownloadProgressResponse{
		Status:      dl.Status,
		Progress:    dl.Progress,
		StreamReady: dl.StreamReady,
		Quality:     dl.Quality,
	})
}
