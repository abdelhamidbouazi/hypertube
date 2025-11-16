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

func (c *MovieController) ServeHLSFile(ctx echo.Context) error {
	path := ctx.Request().URL.Path
	outputDir := services.VideoTranscoderConf.Output.Directory

	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) < 2 {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid path")
	}

	movieIDStr := parts[1]
	movieID, err := strconv.Atoi(movieIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid movie ID")
	}

	// Check if movie is already transcoded
	var downloadedMovie models.DownloadedMovie
	err = c.db.Where("movie_id = ?", movieID).First(&downloadedMovie).Error
	if err == nil && downloadedMovie.Transcoded {
		// Movie is already transcoded, no need to start stream
		log.Printf("Movie %d is already transcoded, serving existing files", movieID)
	} else {
		// Start transcoding if not already started
		c.streamMu.Lock()
		if !c.streamStarted[movieID] {
			c.streamStarted[movieID] = true
			go c.startMovieStream(movieID, outputDir)
		}
		c.streamMu.Unlock()
	}

	filePath := filepath.Join(outputDir, strings.TrimPrefix(path, "/stream/"))
	log.Printf("Waiting for HLS file: %s", filePath)
	if err := c.waitForFile(filePath, 5*time.Second); err != nil {
		log.Printf("File not found after waiting: %s", filePath)
		return echo.NewHTTPError(http.StatusNotFound, "File not ready")
	}

	log.Printf("Serving HLS file: %s", filePath)
	return ctx.File(filePath)
}

func (c *MovieController) startMovieStream(movieID int, outputDir string) {
	var shouldCleanup bool

	defer func() {
		if shouldCleanup || recover() != nil {
			c.streamMu.Lock()
			delete(c.streamStarted, movieID)
			c.streamMu.Unlock()
		}
	}()

	activeDownload, err := c.findAndDownloadMovie(movieID)
	if err != nil {
		shouldCleanup = true
		return
	}

	hlsOutputDir := filepath.Join(services.VideoTranscoderConf.Output.Directory, fmt.Sprintf("%d", movieID))
	if err := os.MkdirAll(hlsOutputDir, 0755); err != nil {
		shouldCleanup = true
		return
	}

	// Check if this is a completed download (FilePath already set, no active torrent)
	activeDownload.Mu.RLock()
	filePath := activeDownload.FilePath
	videoFile := activeDownload.VideoFile
	status := activeDownload.Status
	activeDownload.Mu.RUnlock()

	if status == "completed" && filePath != "" && videoFile == nil {
		// This is a completed download, skip waiting and go directly to transcoding
		log.Printf("Using completed download for movie %d: %s", movieID, filePath)
	} else {
		// This is an active download, wait for video file to be detected
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

	waitLoop:
		for {
			activeDownload.Mu.RLock()
			videoFile := activeDownload.VideoFile
			activeDownload.Mu.RUnlock()

			if videoFile != nil {
				log.Printf("Video file detected for movie %d: %s", movieID, activeDownload.FilePath)
				break waitLoop
			}

			<-ticker.C
			// Continue waiting indefinitely
		}

		// Wait for sufficient data before transcoding.
		// Start with lower threshold and retry if FFmpeg fails
		minBytesForTranscoding := services.Conf.STREAMING.MinBytesForTranscoding
		minPercentForTranscoding := services.Conf.STREAMING.MinPercentForTranscoding

		// Fallback to defaults if not configured
		if minBytesForTranscoding == 0 {
			minBytesForTranscoding = 5 * 1024 * 1024 // 5MB
		}
		if minPercentForTranscoding == 0 {
			minPercentForTranscoding = 0.5 // 0.5%
		}

		log.Printf("Waiting for sufficient download progress for movie %d (min: %d bytes or %.2f%%)...",
			movieID, minBytesForTranscoding, minPercentForTranscoding)
		for {
			activeDownload.Mu.RLock()
			progress := activeDownload.Progress
			videoFile := activeDownload.VideoFile
			filePath := activeDownload.FilePath
			activeDownload.Mu.RUnlock()

			if videoFile != nil {
				totalSize := videoFile.Length()
				var downloadedBytes int64

				if totalSize > 0 {
					downloadedBytes = int64(float64(totalSize) * progress / 100.0)
				} else {
					// Fallback: stat the partial file on disk if total size isn't available yet
					partPath := filePath + ".part"
					if fi, err := os.Stat(partPath); err == nil {
						downloadedBytes = fi.Size()
					}
				}

				log.Printf("Movie %d download progress: %.2f%%, downloaded %d MB, total %d MB", movieID, progress, downloadedBytes/(1024*1024), videoFile.Length()/(1024*1024))

				if downloadedBytes >= minBytesForTranscoding || progress >= minPercentForTranscoding {
					log.Printf("Sufficient data downloaded for movie %d: %.2f%% (%d MB)", movieID, progress, downloadedBytes/(1024*1024))
					break
				}
			}

			<-ticker.C
			// Continue waiting
		}
	}

	// Retry transcoding loop - keep trying infinitely while download is active
	retryDelay := 10 * time.Second
	attempt := 0

	for {
		attempt++
		// Use .part file if regular file doesn't exist (torrent still downloading)
		inputFile := activeDownload.FilePath
		if _, err := os.Stat(inputFile); os.IsNotExist(err) {
			partFile := inputFile + ".part"
			if _, err := os.Stat(partFile); err == nil {
				inputFile = partFile
				log.Printf("Using .part file for transcoding: %s", partFile)
			}
		}

		log.Printf("Starting FFmpeg transcoding for movie %d (attempt %d) with input: %s", movieID, attempt, inputFile)
		err := c.runFFmpegTranscoding(inputFile, hlsOutputDir)

		if err == nil {
			log.Printf("FFmpeg transcoding completed successfully for movie %d", movieID)

			// Mark movie as transcoded in database
			var downloadedMovie models.DownloadedMovie
			if err := c.db.Where("movie_id = ?", movieID).First(&downloadedMovie).Error; err == nil {
				c.db.Model(&downloadedMovie).Update("transcoded", true)
				log.Printf("Marked movie %d as transcoded in database", movieID)
			}

			break
		}

		// Check if download is still active
		activeDownload.Mu.RLock()
		progress := activeDownload.Progress
		status := activeDownload.Status
		activeDownload.Mu.RUnlock()

		if status == "completed" || progress >= 100.0 {
			// Download is complete, don't retry
			log.Printf("FFmpeg transcoding failed for movie %d after completion: %v", movieID, err)
			shouldCleanup = true
			break
		}

		// Download still active, retry after delay
		log.Printf("FFmpeg transcoding failed for movie %d (attempt %d): %v. Download at %.2f%%, retrying in %v...",
			movieID, attempt, err, progress, retryDelay)
		time.Sleep(retryDelay)
	}
}

func (c *MovieController) runFFmpegTranscoding(inputFile string, hlsOutputDir string) error {
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

	log.Printf("Starting FFmpeg with command: ffmpeg %v", strings.Join(args, " "))

	cmd := exec.Command("ffmpeg", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		log.Printf("FFmpeg command failed: %v", err)
		return err
	}

	log.Printf("FFmpeg command completed successfully")
	return nil
}

func (c *MovieController) waitForFile(filePath string, timeout time.Duration) error {
	for {
		if _, err := os.Stat(filePath); err == nil {
			time.Sleep(100 * time.Millisecond)
			return nil
		}
		time.Sleep(500 * time.Millisecond)
	}
}

func (c *MovieController) findAndDownloadMovie(movieID int) (*models.TorrentDownload, error) {
	movieIDStr := strconv.Itoa(movieID)
	details, err := c.movieService.GetMovieDetails(movieIDStr)
	if err != nil {
		return nil, fmt.Errorf("failed to get movie details: %w", err)
	}

	var imdbID string
	if details.IMDbID != "" {
		imdbID = details.IMDbID
	} else {
		imdbID, err = c.movieService.GetIMDbIDFromTMDb(movieIDStr)
		if err != nil {
			return nil, fmt.Errorf("failed to get IMDb ID: %w", err)
		}
	}

	torrents, err := c.movieService.SearchTorrentsByIMDb(imdbID)
	if err != nil || len(torrents) == 0 {
		year := ""
		if len(details.ReleaseDate) >= 4 {
			year = details.ReleaseDate[:4]
		}
		torrents, err = c.movieService.SearchTorrents(details.Title, year)
		if err != nil || len(torrents) == 0 {
			return nil, fmt.Errorf("no torrents found for movie")
		}
	}

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
		// Continue waiting indefinitely
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
