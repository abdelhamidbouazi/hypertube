package controllers

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"server/internal/models"
	"server/internal/services"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/anacrolix/torrent"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

type MovieController struct {
	movieService       *services.MovieService
	torrentService     *services.TorrentService
	transcodingService *services.TranscodingService
	db                 *gorm.DB
	streamStarted      map[int]bool // tracks if startMovieStream has been called per movieID
	streamMu           sync.Mutex   // protects streamStarted map
}

func NewMovieController(ms *services.MovieService, ts *services.TorrentService, tcs *services.TranscodingService, db *gorm.DB) *MovieController {
	return &MovieController{
		movieService:       ms,
		torrentService:     ts,
		transcodingService: tcs,
		db:                 db,
		streamStarted:      make(map[int]bool),
	}
}

// MovieDetailsDoc is a doc-only schema for MovieDetails that avoids gorm.Model in nested types
type MovieDetailsDoc struct {
	ID           int               `json:"id"`
	Title        string            `json:"title"`
	Overview     string            `json:"overview"`
	ReleaseDate  string            `json:"release_date"`
	Runtime      int               `json:"runtime"`
	PosterPath   string            `json:"poster_path"`
	BackdropPath string            `json:"backdrop_path"`
	VoteAverage  float64           `json:"vote_average"`
	IMDbID       string            `json:"imdb_id"`
	Language     string            `json:"original_language,omitempty"`
	IsAvailable  bool              `json:"is_available"`
	StreamURL    string            `json:"stream_url"`
	Cast         []models.Cast     `json:"cast"`
	Director     []models.Person   `json:"director"`
	Producer     []models.Person   `json:"producer"`
	Genres       []models.Genre    `json:"genres"`
	Comments     []CommentResponse `json:"comments"`
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
	genresParam := ctx.QueryParam("genres")       // comma-separated names or ids
	yearFromParam := ctx.QueryParam("yearFrom")   // int
	yearToParam := ctx.QueryParam("yearTo")       // int
	minRatingParam := ctx.QueryParam("minRating") // float
	sortParam := ctx.QueryParam("sort")           // year|year_asc|year_desc|rating

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

func validateAndExtractMovieFilePath(path string) (int, error) {
	var movieId int
	var err error

	services.Logger.Info("Validating path: %s", path)
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

func (c *MovieController) ensureMovieIsPartiallyDownloadedAndStartedTranscoding(movieID int, outputDir string) error {
	services.Logger.Info(fmt.Sprintf("Ensuring movie %d is partially downloaded and started transcoding", movieID))
	var downloadedMovie models.DownloadedMovie
	err := c.db.Where("movie_id = ?", movieID).First(&downloadedMovie).Error
	if err == nil && downloadedMovie.Transcoded {
		services.Logger.Info(fmt.Sprintf("Movie %d is already transcoded", movieID))
		return nil
	}
	if err != nil {
		services.Logger.Info(fmt.Sprintf("Movie %d is not downloaded", movieID))
	}
	services.Logger.Info(fmt.Sprintf("Acquiring lock to check transcoding status for movie %d", movieID))
	c.streamMu.Lock()
	services.Logger.Info(fmt.Sprintf("Checking if transcoding already started for movie %d", movieID))
	if !c.streamStarted[movieID] {
		services.Logger.Info(fmt.Sprintf("Starting transcoding for movie %d", movieID))
		c.streamStarted[movieID] = true
		go c.startMovieStream(movieID, outputDir)
	}
	services.Logger.Info(fmt.Sprintf("Releasing lock after checking transcoding status for movie %d", movieID))
	c.streamMu.Unlock()
	services.Logger.Info(fmt.Sprintf("Transcoding process initiated for movie %d", movieID))
	return err
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

	services.Logger.Info("Requested HLS path: " + path)

	movieID, err := validateAndExtractMovieFilePath(path)
	if err != nil {
		return err
	}

	outputDir := services.VideoTranscoderConf.Output.Directory
	filePath := filepath.Join(outputDir, strings.TrimPrefix(path, "/stream/"))

	services.Logger.Info("Serving HLS file: " + filePath)

	err = c.checkFileExits(filePath)
	if err != nil {
		services.Logger.Info("File does not exist yet: " + filePath)
		err = c.ensureMovieIsPartiallyDownloadedAndStartedTranscoding(movieID, outputDir)
		if err != nil {
			return err
		}

		services.Logger.Info("Waiting for file to be ready: " + filePath)
		if err := c.waitForFile(filePath, 5*time.Second); err != nil {
			return echo.NewHTTPError(http.StatusNotFound, "File not ready")
		}

		services.Logger.Info("Serving file: " + filePath)
	}
	return ctx.File(filePath)
}

func (c *MovieController) startMovieStreamCleanup(shouldCleanup bool, movieID int) {
	services.Logger.Info("Cleaning up startMovieStream for movieID: " + strconv.Itoa(movieID))
	if shouldCleanup || recover() != nil {
		services.Logger.Info("Acquiring lock to clean up streamStarted entry for movieID: " + strconv.Itoa(movieID))
		c.streamMu.Lock()
		delete(c.streamStarted, movieID)
		c.streamMu.Unlock()
		services.Logger.Info("Released lock after cleaning up streamStarted entry for movieID: " + strconv.Itoa(movieID))
	}
}

func (c *MovieController) ensureHLSMovieDirectory(hlsOutputDir string) error {
	return os.MkdirAll(hlsOutputDir, 0755)
}

func (c *MovieController) getTorrentMovieDetails(activeDownload *models.TorrentDownload) (string, *torrent.File, string, float64) {
	activeDownload.Mu.RLock()
	filePath := activeDownload.FilePath
	videoFile := activeDownload.VideoFile
	status := activeDownload.Status
	progress := activeDownload.Progress
	activeDownload.Mu.RUnlock()

	return filePath, videoFile, status, progress
}

func (c *MovieController) waitUntilVideoFileIsReady(
	activeDownload *models.TorrentDownload, filePath string,
	videoFile *torrent.File,
	status string,
) {
	if !(status == "completed" && filePath != "" && videoFile == nil) {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

	waitLoop:
		for {
			_, videoFile, _, _ = c.getTorrentMovieDetails(activeDownload)

			if videoFile != nil {
				break waitLoop
			}

			<-ticker.C
		}

		minBytesForTranscoding := services.Conf.STREAMING.MinBytesForTranscoding
		minPercentForTranscoding := services.Conf.STREAMING.MinPercentForTranscoding

		for {
			filePath, videoFile, _, progress := c.getTorrentMovieDetails(activeDownload)

			if videoFile != nil {
				totalSize := videoFile.Length()
				var downloadedBytes int64

				if totalSize > 0 {
					downloadedBytes = int64(float64(totalSize) * progress / 100.0)
				} else {
					partPath := filePath + ".part"
					if fi, err := os.Stat(partPath); err == nil {
						downloadedBytes = fi.Size()
					}
				}

				if downloadedBytes >= minBytesForTranscoding || progress >= minPercentForTranscoding {
					break
				}
			}

			<-ticker.C
		}
	}
}

func (c *MovieController) tryFFmpegTranscoding(
	activeDownload *models.TorrentDownload,
	movieID int,
	hlsOutputDir string,
	shouldchange *bool,
) {
	retryDelay := 10 * time.Second
	attempt := 0

	for {
		attempt++
		inputFile := activeDownload.FilePath
		if _, err := os.Stat(inputFile); os.IsNotExist(err) {
			partFile := inputFile + ".part"
			if _, err := os.Stat(partFile); err == nil {
				inputFile = partFile
			}
		}

		srtFiles := make([]string, 0)
		err := c.runFFmpegTranscoding(inputFile, srtFiles, hlsOutputDir)
		if err == nil {
			var downloadedMovie models.DownloadedMovie
			if err := c.db.Where("movie_id = ?", movieID).First(&downloadedMovie).Error; err == nil {
				c.db.Model(&downloadedMovie).Update("transcoded", true)
			}
			*shouldchange = false
			break
		}

		_, _, status, progress := c.getTorrentMovieDetails(activeDownload)
		if status == "completed" || progress >= 100.0 {
			*shouldchange = true
			break
		}

		time.Sleep(retryDelay)
	}
}

func (c *MovieController) startMovieStream(movieID int, outputDir string) {
	shouldCleanup := false
	defer c.startMovieStreamCleanup(shouldCleanup, movieID)

	movieIDStr := fmt.Sprintf("%d", movieID)
	hlsOutputDir := filepath.Join(services.VideoTranscoderConf.Output.Directory, movieIDStr)

	activeDownload, err := c.findAndDownloadMovie(movieID)
	if err != nil {
		shouldCleanup = true
		return
	}

	if err := c.ensureHLSMovieDirectory(hlsOutputDir); err != nil {
		shouldCleanup = true
		return
	}

	filePath, videoFile, status, _ := c.getTorrentMovieDetails(activeDownload)
	c.waitUntilVideoFileIsReady(activeDownload, filePath, videoFile, status)
	c.tryFFmpegTranscoding(activeDownload, movieID, hlsOutputDir, &shouldCleanup)
}

func (c *MovieController) runFFmpegTranscoding(inputFile string, srtFiles []string, hlsOutputDir string) error {
	var args []string
	args = append(args, "-i", inputFile)

	args = append(args,
		"-c:v", "libx264",
		"-preset", services.VideoTranscoderConf.Encoding.Preset,
		"-crf", fmt.Sprintf("%d", services.VideoTranscoderConf.Encoding.CRF),
		"-g", fmt.Sprintf("%d", services.VideoTranscoderConf.Encoding.GOPSize),
		"-keyint_min", fmt.Sprintf("%d", services.VideoTranscoderConf.Encoding.KeyintMin),
		"-sc_threshold", fmt.Sprintf("%d", services.VideoTranscoderConf.Encoding.SCThreshold),
	)

	args = append(args,
		"-c:a", "aac",
		"-b:a", services.VideoTranscoderConf.Encoding.AudioBitrate,
		"-ar", fmt.Sprintf("%d", services.VideoTranscoderConf.Encoding.AudioSampleRate),
	)

	variantIndex := 0
	for _, quality := range services.VideoTranscoderConf.Qualities {
		if !quality.Enabled {
			continue
		}

		qualityDir := filepath.Join(hlsOutputDir, quality.Name)
		if err := os.MkdirAll(qualityDir, 0755); err != nil {
			return err
		}

		args = append(args,
			"-map", "0:v:0",
			"-map", "0:a:0",
		)

		args = append(args,
			fmt.Sprintf("-s:v:%d", variantIndex), quality.Resolution,
			fmt.Sprintf("-b:v:%d", variantIndex), quality.VideoBitrate,
			fmt.Sprintf("-maxrate:%d", variantIndex), quality.MaxRate,
			fmt.Sprintf("-bufsize:%d", variantIndex), quality.BufSize,
		)

		variantIndex++
	}

	args = append(args,
		"-f", "hls",
		"-hls_time", fmt.Sprintf("%d", services.VideoTranscoderConf.Output.SegmentTime),
		"-hls_playlist_type", "event",
		"-hls_flags", "temp_file+independent_segments",
		"-hls_segment_filename", filepath.Join(hlsOutputDir, "%v/segment%03d.ts"),
	)

	var varStreamMap strings.Builder
	idx := 0
	for _, quality := range services.VideoTranscoderConf.Qualities {
		if !quality.Enabled {
			continue
		}
		if idx > 0 {
			varStreamMap.WriteString(" ")
		}
		varStreamMap.WriteString(fmt.Sprintf("v:%d,a:%d", idx, idx))
		idx++
	}

	args = append(args,
		"-var_stream_map", varStreamMap.String(),
		"-master_pl_name", "master.m3u8",
	)

	outputPattern := filepath.Join(hlsOutputDir, "%v/playlist.m3u8")
	args = append(args, outputPattern)

	cmd := exec.Command("ffmpeg", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

func (c *MovieController) checkFileExits(filePath string) error {
	services.Logger.Info("Checking if file exists: " + filePath)
	_, err := os.Stat(filePath)
	return err
}

func (c *MovieController) waitForFile(filePath string, timeout time.Duration) error {
	for {
		if err := c.checkFileExits(filePath); err == nil {
			time.Sleep(100 * time.Millisecond)
			return nil
		}
		time.Sleep(500 * time.Millisecond)
	}
}

func (c *MovieController) findAndDownloadMovie(movieID int) (*models.TorrentDownload, error) {
	services.Logger.Info(fmt.Sprintf("Finding and downloading movie with ID: %d", movieID))
	details, err := c.movieService.GetIMDbIDFromTMDb(movieID)

	services.Logger.Info(fmt.Sprintf("Movie details fetched: %+v", details))
	torrents, err := c.movieService.SearchTorrentsByIMDb(*details)

	if len(torrents) == 0 {
		return nil, fmt.Errorf("no suitable torrent found")
	}

	bestTorrent := &torrents[0]

	quality := bestTorrent.Quality
	if quality == "" {
		quality = "720p"
	}

	download, err := c.torrentService.GetOrStartDownload(movieID, bestTorrent.Magnet, quality)
	if err != nil {
		return nil, fmt.Errorf("failed to start download: %w", err)
	}

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		download.Mu.RLock()
		ready := download.StreamingReady
		filePath := download.FilePath
		download.Mu.RUnlock()

		if ready && filePath != "" {
			return download, nil
		}

		<-ticker.C
	}
}

// @Accept       */*
// @Produce      video/mp4
// @Param        id       path   string  true  "Movie ID"
// @Param        quality  query  string  false "Desired quality (default: 720p)"
// @Security     JWT
// @Success      200  {file}    binary  "Video stream"
// @Failure      102  {object}  utils.HTTPError  "Still downloading"
// @Failure      401  {object}  utils.HTTPErrorUnauthorized
// @Failure      404  {object}  utils.HTTPError
// @Router       /stream/{id} [get]
func (c *MovieController) StreamMovie(ctx echo.Context) error {
	movieID := ctx.Param("id")
	quality := ctx.QueryParam("quality")
	if quality == "" {
		quality = "720p" // default quality
	}

	var downloadedMovie models.DownloadedMovie
	err := c.db.Where("movie_id = ? AND quality = ?", movieID, quality).First(&downloadedMovie).Error

	// if the movie is already downloaded
	if err == nil && downloadedMovie.FilePath != "" {
		if _, err := os.Stat(downloadedMovie.FilePath); err == nil {
			return c.serveVideoFile(ctx, downloadedMovie.FilePath)
		}
	}

	// movie is not downloaded
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
