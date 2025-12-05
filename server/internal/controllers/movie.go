package controllers

import (
	"fmt"
	"io"
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
	"github.com/grafov/m3u8"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

type MovieController struct {
	movieService        *services.MovieService
	torrentService      *services.TorrentService
	subtitleService     *services.SubtitleService
	db                  *gorm.DB
	websocketController *WebSocketController
	streamStarted       map[int]bool
	streamStatus        map[int]map[string]interface{}
	streamMu            sync.Mutex
}

func getLanguageLabel(code string) string {
	languageLabels := map[string]string{
		"en": "English",
		"fr": "French",
		"es": "Spanish",
		"ar": "Arabic",
		"de": "German",
		"it": "Italian",
		"pt": "Portuguese",
		"ru": "Russian",
		"ja": "Japanese",
		"ko": "Korean",
		"zh": "Chinese",
		"hi": "Hindi",
		"nl": "Dutch",
		"pl": "Polish",
		"tr": "Turkish",
	}
	if label, ok := languageLabels[code]; ok {
		return label
	}
	return code
}

func NewMovieController(ms *services.MovieService, ts *services.TorrentService, db *gorm.DB, wsc *WebSocketController) *MovieController {
	subdlAPIKey := services.Conf.MOVIE_APIS.SUBDL.APIKey
	subtitleService, err := services.NewSubtitleService(subdlAPIKey)
	if err != nil {
		log.Printf("Warning: Failed to initialize subtitle service: %v", err)
	}

	return &MovieController{
		movieService:   ms,
		torrentService: ts,
		// transcodingService:  tcs,
		subtitleService:     subtitleService,
		db:                  db,
		websocketController: wsc,
		streamStarted:       make(map[int]bool),
		streamStatus:        make(map[int]map[string]interface{}),
	}
}

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
	return nil
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

	err = c.checkFileExits(filePath)
	if err != nil {
		err = c.ensureMovieIsPartiallyDownloadedAndStartedTranscoding(movieID, outputDir)
		if err != nil {
			return err
		}

		if err := c.waitForFile(filePath, 5*time.Second); err != nil {
			c.streamMu.Lock()
			status, exists := c.streamStatus[movieID]
			c.streamMu.Unlock()

			if !exists {
				status = map[string]interface{}{
					"stage":   "initializing",
					"message": "Stream initialization in progress",
				}
			}

			return ctx.JSON(http.StatusAccepted, map[string]interface{}{
				"error":  status["message"],
				"status": status,
			})
		}
	}
	return ctx.File(filePath)
}

func (c *MovieController) startMovieStreamCleanup(shouldCleanup bool, movieID int) {
	if shouldCleanup || recover() != nil {
		c.streamMu.Lock()
		delete(c.streamStarted, movieID)
		c.streamMu.Unlock()
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

func (c *MovieController) updateStreamStatus(movieID int, stage string, message string, additionalData map[string]interface{}) {
	c.streamMu.Lock()

	status := map[string]interface{}{
		"movieID": movieID,
		"stage":   stage,
		"message": message,
	}

	for key, value := range additionalData {
		status[key] = value
	}

	c.streamStatus[movieID] = status
	c.streamMu.Unlock()

	if c.websocketController != nil {
		c.websocketController.UpdateStreamState(movieID, status)
	}
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

		_, videoFile, _, _ := c.getTorrentMovieDetails(activeDownload)

		if videoFile == nil {
			time.Sleep(retryDelay)
			continue
		}

		reader := videoFile.NewReader()
		reader.SetResponsive()                // Blocks until pieces are complete
		reader.SetReadahead(10 * 1024 * 1024) // 10MB read-ahead

		err := c.runFFmpegTranscoding(reader, hlsOutputDir)
		reader.Close()

		if err == nil {
			var downloadedMovie models.DownloadedMovie
			if err := c.db.Where("movie_id = ?", movieID).First(&downloadedMovie).Error; err == nil {
				c.db.Model(&downloadedMovie).Update("transcoded", true)
			}
			c.updateStreamStatus(movieID, "ready", "Stream is ready to play", map[string]interface{}{
				"transcodingStatus": "ready",
			})
			break
		}

		c.updateStreamStatus(movieID, "transcoding", fmt.Sprintf("Transcoding attempt %d failed, retrying...", attempt), map[string]interface{}{
			"transcodingStatus": "retrying",
			"attempt":           attempt,
			"error":             err.Error(),
		})

		_, _, status, progress := c.getTorrentMovieDetails(activeDownload)
		if status == "completed" || progress >= 100.0 {
			c.updateStreamStatus(movieID, "error", "Transcoding failed after download completed", map[string]interface{}{
				"transcodingStatus": "failed",
				"lastError":         err.Error(),
			})
			*shouldchange = true
			break
		}

		time.Sleep(retryDelay)
	}
}

func (c *MovieController) startMovieStream(movieID int, outputDir string) {
	shouldCleanup := false
	defer c.startMovieStreamCleanup(shouldCleanup, movieID)

	c.updateStreamStatus(movieID, "initializing", "Starting movie stream", nil)
	movieIDStr := fmt.Sprintf("%d", movieID)
	hlsBaseDir := services.VideoTranscoderConf.Output.Directory
	if services.Conf.STREAMING.HLSOutputDir != "" {
		hlsBaseDir = services.Conf.STREAMING.HLSOutputDir
	}
	hlsOutputDir := filepath.Join(hlsBaseDir, movieIDStr)

	srtFiles := c.downloadMovieSubtitles(movieID)

	c.updateStreamStatus(movieID, "downloading", "Finding and downloading movie", nil)
	activeDownload, err := c.findAndDownloadMovie(movieID)
	if err != nil {
		c.updateStreamStatus(movieID, "error", "Failed to download movie: "+err.Error(), nil)
		shouldCleanup = true
		return
	}

	if err := c.ensureHLSMovieDirectory(hlsOutputDir); err != nil {
		c.updateStreamStatus(movieID, "error", "Failed to create HLS output directory: "+err.Error(), nil)
		shouldCleanup = true
		return
	}

	filePath, videoFile, status, progress := c.getTorrentMovieDetails(activeDownload)
	c.updateStreamStatus(movieID, "downloading", fmt.Sprintf("Downloading: %.1f%% complete", progress), map[string]interface{}{
		"downloadProgress": progress,
		"downloadStatus":   status,
	})
	c.waitUntilVideoFileIsReady(activeDownload, filePath, videoFile, status)

	if err := c.convertSubtitlesToHLS(srtFiles, hlsOutputDir); err != nil {
		c.updateStreamStatus(movieID, "error", "Failed to convert subtitles to HLS: "+err.Error(), nil)
		shouldCleanup = true
		return
	}

	_, err = c.createMasterPlaylist(hlsOutputDir, srtFiles)
	if err != nil {
		c.updateStreamStatus(movieID, "error", "Failed to create master playlist: "+err.Error(), nil)
		shouldCleanup = true
		return
	}

	c.updateStreamStatus(movieID, "transcoding", "Converting video to HLS format", map[string]interface{}{
		"transcodingStatus": "in_progress",
	})
	c.tryFFmpegTranscoding(activeDownload, movieID, hlsOutputDir, &shouldCleanup)
}

func (c *MovieController) createMasterPlaylist(hlsOutputDir string, srtFiles []string) (*m3u8.MasterPlaylist, error) {
	masterPlaylistPath := filepath.Join(hlsOutputDir, "master.m3u8")

	masterPlaylist := m3u8.NewMasterPlaylist()
	masterPlaylist.SetVersion(3)

	var subtitleAlternatives []*m3u8.Alternative
	subsDir := filepath.Join(hlsOutputDir, "subs")
	if entries, err := os.ReadDir(subsDir); err == nil {
		for _, entry := range entries {
			if entry.IsDir() {
				lang := entry.Name()
				playlistPath := filepath.Join(subsDir, lang, "playlist.m3u8")
				if _, err := os.Stat(playlistPath); err == nil {
					vttPlaylist := fmt.Sprintf("subs/%s/playlist.m3u8", lang)
					isDefault := lang == "en"
					subtitleAlt := &m3u8.Alternative{
						GroupId:    "subs",
						Type:       "SUBTITLES",
						Name:       getLanguageLabel(lang),
						Language:   lang,
						Default:    isDefault,
						Autoselect: "NO",
						URI:        vttPlaylist,
					}
					subtitleAlternatives = append(subtitleAlternatives, subtitleAlt)
				}
			}
		}
	}

	for _, quality := range services.VideoTranscoderConf.Qualities {
		if !quality.Enabled {
			continue
		}

		uri := fmt.Sprintf("%s/playlist.m3u8", quality.Name)
		params := m3u8.VariantParams{
			Bandwidth:    parseBandwidth(quality.VideoBitrate),
			Resolution:   quality.Resolution,
			Codecs:       "avc1.640028,mp4a.40.2",
			Alternatives: subtitleAlternatives,
		}

		if len(subtitleAlternatives) > 0 {
			params.Subtitles = "subs"
		}

		masterPlaylist.Append(uri, nil, params)
	}

	masterFile, err := os.Create(masterPlaylistPath)
	if err != nil {
		return nil, err
	}
	defer masterFile.Close()

	_, err = masterFile.Write(masterPlaylist.Encode().Bytes())
	return masterPlaylist, err
}

func (c *MovieController) downloadMovieSubtitles(movieID int) []string {
	if c.subtitleService == nil {
		services.Logger.Warn("Subtitle service not initialized, skipping subtitle download")
		return []string{}
	}

	baseDir := services.Conf.STREAMING.SubtitlesDir
	if baseDir == "" {
		baseDir = "subtitles"
	}

	dirPath, err := c.subtitleService.CreateSubtitlesDirectory(movieID, baseDir)
	if err != nil {
		services.Logger.Error(fmt.Sprintf("Failed to create subtitles directory for movie %d: %v", movieID, err))
		return []string{}
	}

	services.Logger.Info(fmt.Sprintf("Downloading subtitles for movie %d from subdl.com", movieID))
	downloadedCount := c.subtitleService.DownloadSubtitles(movieID, dirPath)
	services.Logger.Info(fmt.Sprintf("Downloaded %d subtitle(s) for movie %d", downloadedCount, movieID))

	srtFiles, err := services.FindFilesWithExtension(dirPath, "srt")
	if err != nil {
		services.Logger.Error(fmt.Sprintf("Error finding downloaded SRT files: %v", err))
		return []string{}
	}

	return srtFiles
}

func (c *MovieController) convertSubtitlesToHLS(srtFiles []string, hlsOutputDir string) error {
	if len(srtFiles) == 0 {
		return nil
	}

	vttTempDir := filepath.Join(hlsOutputDir, "subs_vtt_temp")
	if err := os.MkdirAll(vttTempDir, 0755); err != nil {
		return fmt.Errorf("failed to create VTT temp directory: %w", err)
	}
	defer os.RemoveAll(vttTempDir)

	for _, srtFile := range srtFiles {
		baseName := filepath.Base(srtFile)
		lang := strings.TrimSuffix(baseName, filepath.Ext(baseName))

		langSubsDir := filepath.Join(hlsOutputDir, "subs", lang)
		if err := os.MkdirAll(langSubsDir, 0755); err != nil {
			services.Logger.Error(fmt.Sprintf("Failed to create subtitle directory for %s: %v", lang, err))
			continue
		}

		vttFile := filepath.Join(vttTempDir, fmt.Sprintf("%s.vtt", lang))
		vttFileHandle, err := os.Create(vttFile)
		if err != nil {
			services.Logger.Error(fmt.Sprintf("Failed to create VTT file for %s: %v", lang, err))
			continue
		}

		convertSRTtoVTT(srtFile, vttFileHandle)
		vttFileHandle.Close()

		if _, err := os.Stat(vttFile); os.IsNotExist(err) {
			services.Logger.Error(fmt.Sprintf("VTT file was not created for %s", lang))
			continue
		}

		destVttFile := filepath.Join(langSubsDir, "subtitle.vtt")
		if err := copyFile(vttFile, destVttFile); err != nil {
			services.Logger.Error(fmt.Sprintf("Failed to copy VTT file for %s: %v", lang, err))
			continue
		}

		playlistPath := filepath.Join(langSubsDir, "playlist.m3u8")
		mediaPlaylist, err := m3u8.NewMediaPlaylist(1, 1)
		if err != nil {
			services.Logger.Error(fmt.Sprintf("Failed to create media playlist for %s: %v", lang, err))
			continue
		}

		mediaPlaylist.MediaType = m3u8.VOD
		mediaPlaylist.SetVersion(3)

		if err := mediaPlaylist.Append("subtitle.vtt", 0.0, ""); err != nil {
			services.Logger.Error(fmt.Sprintf("Failed to append segment to playlist for %s: %v", lang, err))
			continue
		}
		mediaPlaylist.Close()

		playlistFile, err := os.Create(playlistPath)
		if err != nil {
			services.Logger.Error(fmt.Sprintf("Failed to create playlist file for %s: %v", lang, err))
			continue
		}
		playlistFile.WriteString(mediaPlaylist.String())
		playlistFile.Close()

		services.Logger.Info(fmt.Sprintf("Successfully converted subtitles for language: %s", lang))
	}

	return nil
}

func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}

func parseBandwidth(bitrate string) uint32 {
	bitrate = strings.TrimSpace(bitrate)
	bitrate = strings.ToLower(bitrate)

	if strings.HasSuffix(bitrate, "k") {
		value := strings.TrimSuffix(bitrate, "k")
		if v, err := strconv.ParseFloat(value, 64); err == nil {
			return uint32(v * 1000)
		}
	} else if strings.HasSuffix(bitrate, "m") {
		value := strings.TrimSuffix(bitrate, "m")
		if v, err := strconv.ParseFloat(value, 64); err == nil {
			return uint32(v * 1000000)
		}
	} else {
		if v, err := strconv.ParseFloat(bitrate, 64); err == nil {
			return uint32(v)
		}
	}

	return 0
}

func (c *MovieController) runFFmpegTranscoding(reader io.Reader, hlsOutputDir string) error {
	var args []string

	args = append(args,
		"-fflags", "+genpts+igndts+discardcorrupt",
		"-err_detect", "ignore_err",
		"-i", "pipe:0")

	args = append(args,
		"-c:v", "libx264",
		"-preset", services.VideoTranscoderConf.Encoding.Preset,
		"-crf", fmt.Sprintf("%d", services.VideoTranscoderConf.Encoding.CRF),
		"-g", fmt.Sprintf("%d", services.VideoTranscoderConf.Encoding.GOPSize),
		"-keyint_min", fmt.Sprintf("%d", services.VideoTranscoderConf.Encoding.KeyintMin),
		"-sc_threshold", fmt.Sprintf("%d", services.VideoTranscoderConf.Encoding.SCThreshold),
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

		args = append(args,
			fmt.Sprintf("-c:a:%d", variantIndex), "aac",
			fmt.Sprintf("-b:a:%d", variantIndex), services.VideoTranscoderConf.Encoding.AudioBitrate,
			fmt.Sprintf("-ar:%d", variantIndex), fmt.Sprintf("%d", services.VideoTranscoderConf.Encoding.AudioSampleRate),
		)

		if services.VideoTranscoderConf.Encoding.AudioChannels > 0 {
			args = append(args,
				fmt.Sprintf("-ac:%d", variantIndex), fmt.Sprintf("%d", services.VideoTranscoderConf.Encoding.AudioChannels),
			)
		}

		variantIndex++
	}

	args = append(args,
		"-f", "hls",
		"-hls_time", fmt.Sprintf("%d", services.VideoTranscoderConf.Output.SegmentTime),
		"-hls_playlist_type", "event",
		"-hls_flags", "temp_file+independent_segments+omit_endlist",
		"-hls_list_size", "0",
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
		varStreamMap.WriteString(fmt.Sprintf("v:%d,a:%d,name:%s", idx, idx, quality.Name))
		idx++
	}

	args = append(args,
		"-var_stream_map", varStreamMap.String(),
		"-hls_segment_filename", filepath.Join(hlsOutputDir, "%v", "segment%03d.ts"),
	)

	outputPattern := filepath.Join(hlsOutputDir, "%v", "playlist.m3u8")
	args = append(args, outputPattern)

	cmd := exec.Command("ffmpeg", args...)
	cmd.Stdin = reader
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

func (c *MovieController) checkFileExits(filePath string) error {
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
	c.updateStreamStatus(movieID, "searching", "Fetching movie information", map[string]interface{}{
		"step": "fetch_details",
	})

	details, err := c.movieService.GetIMDbIDFromTMDb(movieID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch movie details: %w", err)
	}

	c.updateStreamStatus(movieID, "searching", "Searching torrent sources", map[string]interface{}{
		"step":         "search_torrents",
		"imdb_id":      details.IMDbID,
		"title":        details.Title,
		"release_date": details.ReleaseDate,
	})

	torrents, err := c.movieService.SearchTorrentsByIMDb(*details, 30*time.Second)

	if err != nil {
		services.Logger.Error(fmt.Sprintf("Error searching torrents: %v", err))
		return nil, fmt.Errorf("failed to search torrents: %w", err)
	}

	c.updateStreamStatus(movieID, "searching", fmt.Sprintf("Found %d torrent(s)", len(torrents)), map[string]interface{}{
		"step":          "torrents_found",
		"torrent_count": len(torrents),
	})

	if len(torrents) == 0 {
		return nil, fmt.Errorf("no suitable torrent found")
	}

	c.updateStreamStatus(movieID, "searching", "Analyzing torrents and fetching metadata", map[string]interface{}{
		"step":          "analyze_torrents",
		"torrent_count": len(torrents),
	})

	bestTorrent := &torrents[0]

	quality := bestTorrent.Quality
	if quality == "" {
		quality = "720p"
	}

	c.updateStreamStatus(movieID, "searching", "Selected best torrent", map[string]interface{}{
		"step":     "torrent_selected",
		"name":     bestTorrent.Name,
		"quality":  quality,
		"size":     bestTorrent.Size,
		"seeders":  bestTorrent.Seeders,
		"leechers": bestTorrent.Leechers,
	})

	c.updateStreamStatus(movieID, "downloading", "Connecting to peers", map[string]interface{}{
		"step":    "start_download",
		"quality": quality,
	})

	download, err := c.torrentService.GetOrStartDownload(movieID, bestTorrent.Magnet, quality)
	if err != nil {
		c.updateStreamStatus(movieID, "error", "Failed to start download: "+err.Error(), nil)
		return nil, fmt.Errorf("failed to start download: %w", err)
	}

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	waitCount := 0
	lastProgress := 0.0

	for {
		download.Mu.RLock()
		ready := download.StreamingReady
		filePath := download.FilePath
		status := download.Status
		progress := download.Progress
		download.Mu.RUnlock()

		if ready && filePath != "" {
			c.updateStreamStatus(movieID, "downloading", "Download ready for streaming", map[string]interface{}{
				"step":             "download_ready",
				"downloadProgress": progress,
				"fileName":         filepath.Base(filePath),
			})
			return download, nil
		}

		progressChanged := progress-lastProgress >= 1.0
		if progressChanged || waitCount%5 == 0 {
			if progressChanged && progress > 0 {
				c.updateStreamStatus(movieID, "downloading", fmt.Sprintf("Downloading: %.1f%% complete", progress), map[string]interface{}{
					"step":             "downloading",
					"downloadProgress": progress,
					"downloadStatus":   status,
				})
				lastProgress = progress
			}
		}

		waitCount++
		<-ticker.C
	}
}
