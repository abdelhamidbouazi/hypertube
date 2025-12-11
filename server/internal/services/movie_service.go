package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"server/internal/models"
	"server/internal/utils"
	"strings"
	"sync"
	"time"

	"github.com/anacrolix/torrent"
	"github.com/grafov/m3u8"
	"gorm.io/gorm"
)

type MovieService struct {
	apiKey             string
	omdbKey            string
	client             *http.Client
	torrentSources     []string
	genreCacheTime     time.Time
	StreamStatus       sync.Map // map[int]map[string]interface{}
	MasterPlaylists    sync.Map // map[int]*m3u8.MasterPlaylist - movieID -> master playlist
	LastSegmentCache   sync.Map // map[int]string - movieID -> last segment filename
	UserWatchedMovies  sync.Map // map[string]map[int]string - "userID:movieID" -> last segment visited
	SegmentFormatParse string
	SearchSources      map[string]Source
	db                 *gorm.DB
	websocketService   *WebSocketService
	subtitleService    *SubtitleService
	torrentService     *TorrentService
}

func NewMovieService(tmdbKey, omdbKey, watchModeKey string, db *gorm.DB, ws *WebSocketService, subtitleService *SubtitleService, torrentService *TorrentService) *MovieService {
	ms := &MovieService{
		apiKey:             tmdbKey,
		omdbKey:            omdbKey,
		client:             &http.Client{Timeout: 10 * time.Second},
		SegmentFormatParse: VideoTranscoderConf.Output.SegmentFilenameFormat,
		torrentSources: []string{
			"1337x",
			"yts",
		},
		db:               db,
		websocketService: ws,
		subtitleService:  subtitleService,
		torrentService:   torrentService,
	}

	ms.SearchSources = map[string]Source{
		"tmdb": NewTMDB(tmdbKey, omdbKey, ms.genreCacheTime, ms.client),
		// "omdb": NewOMDB(omdbKey, ms.genreCacheTime, ms.client),
	}

	go ms.persistWatchHistoryWorker()
	go ms.cleanupOldHLSFilesWorker(Conf.STREAMING.HLSOutputDir)

	return ms
}

func (ms *MovieService) GetSource(key string) (Source, error) {
	src, ok := ms.SearchSources[key]
	if !ok {
		return nil, fmt.Errorf("Source not found")
	}
	return src, nil
}

// DiscoverMovies calls TMDB discover endpoint with filters
func (ms *MovieService) DiscoverMovies(p MovieDiscoverParams, userID *uint, source string) ([]DiscoverMoviesResp, error) {
	var movies []models.Movie
	var movieIDs []int
	var err error

	if source != "" {
		src, err := ms.GetSource(source)
		if err == nil {
			movies, movieIDs, err = src.DiscoverMovies(p)
			if err != nil {
				return nil, err
			}
		}
	} else {
		for _, s := range ms.SearchSources {
			movies, movieIDs, err = s.DiscoverMovies(p)
			if err == nil && len(movies) > 0 {
				break
			}
		}
	}

	if err != nil {
		return nil, err
	}
	var result []DiscoverMoviesResp
	var watchHistory []models.WatchHistory
	watchHistoryMap := make(map[int]bool)

	if userID != nil {
		err = ms.db.Model(&models.WatchHistory{}).
			Where("user_id = ?", userID).
			Where("movie_id IN ?", movieIDs).
			// Where("watch_count > ?", 0).
			Find(&watchHistory).Error

		for _, wh := range watchHistory {
			watchHistoryMap[wh.MovieID] = true
		}
	}

	for _, m := range movies {
		isWatched := false
		if userID != nil {
			isWatched = watchHistoryMap[m.ID]
		}
		result = append(result, DiscoverMoviesResp{
			ID:          m.ID,
			Title:       m.Title,
			ReleaseDate: m.ReleaseDate,
			PosterPath:  m.PosterPath,
			Overview:    m.Overview,
			Language:    m.Language,
			VoteAverage: m.VoteAverage,
			GenreIDs:    m.GenreIDs,
			IsWatched:   isWatched,
		})
	}

	return result, err
}

func (ms *MovieService) SearchMovies(query string, year string) ([]models.Movie, error) {
	baseURL := "https://api.themoviedb.org/3/search/movie"
	params := url.Values{}
	params.Add("api_key", ms.apiKey)
	params.Add("query", query)
	if year != "" {
		params.Add("year", year)
	}

	fullURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	resp, err := ms.client.Get(fullURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var tmdbResp struct {
		Results []struct {
			ID          int     `json:"id"`
			Title       string  `json:"title"`
			ReleaseDate string  `json:"release_date"`
			PosterPath  string  `json:"poster_path"`
			Overview    string  `json:"overview"`
			VoteAverage float64 `json:"vote_average"`
			GenreIDs    []int   `json:"genre_ids"`
		} `json:"results"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&tmdbResp); err != nil {
		return nil, err
	}

	var movies []models.Movie
	for _, result := range tmdbResp.Results {
		movies = append(movies, models.Movie{
			ID:          result.ID,
			Title:       result.Title,
			ReleaseDate: result.ReleaseDate,
			PosterPath:  result.PosterPath,
			Overview:    result.Overview,
			VoteAverage: result.VoteAverage,
			GenreIDs:    result.GenreIDs,
		})
	}

	return movies, nil
}

func (ms *MovieService) searchTorrentsByIMDb(movie models.MovieDetails, metadataTimeout time.Duration) ([]TorrentSearchResult, error) {
	imdbID := movie.IMDbID
	var results []TorrentSearchResult

	req, err := http.NewRequest("GET", "https://torrentio.strem.fun/sort=seeders%7Cqualityfilter=brremux,hdrall,dolbyvision,4k,2160p,other,scr,cam,unknown/stream/movie/"+imdbID+".json", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request for IMDb ID %s: %v", imdbID, err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
	res, err := ms.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to find torrents for IMDb ID %s: %v", imdbID, err)
	}

	defer res.Body.Close()

	var data TorrentionResponse

	if err := json.NewDecoder(res.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to decode torrentio response for IMDb ID %s: %v", imdbID, err)
	}

	for _, t := range data.Streams {
		results = append(results, TorrentSearchResult{InfoHash: t.InfoHash, Name: t.Title})
	}

	return results, nil
}

func (ms *MovieService) TrackUserSegment(userID uint, movieID int) {
	key := fmt.Sprintf("%d:%d", userID, movieID)
	ms.UserWatchedMovies.Store(key, true)
}

func (ms *MovieService) persistWatchHistoryWorker() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		ms.UserWatchedMovies.Range(func(key, value interface{}) bool {
			keyStr := key.(string)

			var userID uint
			var movieID int

			if _, err := fmt.Sscanf(keyStr, "%d:%d", &userID, &movieID); err != nil {
				return true
			}

			var watchHistory models.WatchHistory
			result := ms.db.Where("user_id = ? AND movie_id = ?", userID, movieID).First(&watchHistory)

			if result.Error == gorm.ErrRecordNotFound {
				watchHistory = models.WatchHistory{
					UserID:     userID,
					MovieID:    movieID,
					WatchCount: 0,
					WatchedAt:  time.Now(),
				}

				ms.db.Create(&watchHistory)
			}

			return true
		})
	}
}

func (ms *MovieService) cleanupOldHLSFilesWorker(hlsDir string) {
	ticker := time.NewTicker(4 * time.Second)
	defer ticker.Stop()

	type MovieLastWatched struct {
		MovieID   int
		WatchedAt time.Time
	}

	for {
		var oldMovies []MovieLastWatched
		oneMonthAgo := time.Now().AddDate(0, -1, 0)

		err := ms.db.Model(&models.WatchHistory{}).
			Select("movie_id, MAX(watched_at) as watched_at").
			Group("movie_id").
			Having("MAX(watched_at) < ?", oneMonthAgo).
			Find(&oldMovies).Error

		if err != nil || len(oldMovies) == 0 {
			return
		}

		for _, movie := range oldMovies {
			movieHLSDir := filepath.Join(hlsDir, fmt.Sprintf("%d", movie.MovieID))
			if _, err := os.Stat(movieHLSDir); err == nil {
				os.RemoveAll(movieHLSDir)
			}
		}

		<-ticker.C
	}
}

func (ms *MovieService) EnsureMovieIsPartiallyDownloadedAndStartedTranscoding(movieID int, outputDir string) error {
	var downloadedMovie models.DownloadedMovie
	err := ms.db.Where("movie_id = ?", movieID).First(&downloadedMovie).Error
	if err == nil && downloadedMovie.Transcoded {
		Logger.Info(fmt.Sprintf("Movie %d is already transcoded", movieID))
		return nil
	}
	if err != nil {
		Logger.Info(fmt.Sprintf("Movie %d is not downloaded", movieID))
	}
	Logger.Info(fmt.Sprintf("Checking if transcoding already started for movie %d", movieID))
	initialStatus := map[string]interface{}{
		"movieID": movieID,
		"stage":   "initializing",
		"message": "Stream initialization starting",
	}
	if _, loaded := ms.StreamStatus.LoadOrStore(movieID, initialStatus); !loaded {
		Logger.Info(fmt.Sprintf("Starting transcoding for movie %d", movieID))
		go ms.startMovieStream(movieID, outputDir)
	}
	Logger.Info(fmt.Sprintf("Transcoding check completed for movie %d", movieID))
	Logger.Info(fmt.Sprintf("Transcoding process initiated for movie %d", movieID))
	return nil
}

func (ms *MovieService) updateStreamStatus(movieID int, stage string, message string, additionalData map[string]interface{}) {
	status := map[string]interface{}{
		"movieID": movieID,
		"stage":   stage,
		"message": message,
	}

	for key, value := range additionalData {
		status[key] = value
	}

	ms.StreamStatus.Store(movieID, status)

	if ms.websocketService != nil {
		ms.websocketService.UpdateStreamState(movieID, status)
	}
}

func (ms *MovieService) startMovieStream(movieID int, outputDir string) {
	ms.updateStreamStatus(movieID, "initializing", "Starting movie stream", nil)
	movieIDStr := fmt.Sprintf("%d", movieID)
	hlsBaseDir := VideoTranscoderConf.Output.Directory
	if Conf.STREAMING.HLSOutputDir != "" {
		hlsBaseDir = Conf.STREAMING.HLSOutputDir
	}
	hlsOutputDir := filepath.Join(hlsBaseDir, movieIDStr)

	srtFiles := ms.downloadMovieSubtitles(movieID)

	ms.updateStreamStatus(movieID, "downloading", "Finding and downloading movie", nil)
	activeDownload, err := ms.findAndDownloadMovie(movieID)
	if err != nil {
		ms.updateStreamStatus(movieID, "error", "Failed to download movie: "+err.Error(), nil)
		return
	}

	if err := os.MkdirAll(hlsOutputDir, 0755); err != nil {
		ms.updateStreamStatus(movieID, "error", "Failed to create HLS output directory: "+err.Error(), nil)
		return
	}

	filePath, videoFile, status, progress := ms.getTorrentMovieDetails(activeDownload)
	ms.updateStreamStatus(movieID, "downloading", fmt.Sprintf("Downloading: %.1f%% complete", progress), map[string]interface{}{
		"downloadProgress": progress,
		"downloadStatus":   status,
	})
	ms.waitUntilVideoFileIsReady(activeDownload, filePath, videoFile, status)

	if err := ms.convertSubtitlesToHLS(srtFiles, hlsOutputDir); err != nil {
		ms.updateStreamStatus(movieID, "error", "Failed to convert subtitles to HLS: "+err.Error(), nil)
		return
	}

	masterPlaylist, err := ms.createMasterPlaylist(hlsOutputDir, srtFiles)
	if err != nil {
		ms.updateStreamStatus(movieID, "error", "Failed to create master playlist: "+err.Error(), nil)
		return
	}

	ms.updateStreamStatus(movieID, "transcoding", "Converting video to HLS format", map[string]interface{}{
		"transcodingStatus": "in_progress",
	})
	ms.tryFFmpegTranscodingWithPlaylist(activeDownload, movieID, hlsOutputDir, masterPlaylist)
}

func (ms *MovieService) downloadMovieSubtitles(movieID int) []string {
	if ms.subtitleService == nil {
		Logger.Warn("Subtitle service not initialized, skipping subtitle download")
		return []string{}
	}

	baseDir := Conf.STREAMING.SubtitlesDir
	if baseDir == "" {
		baseDir = "subtitles"
	}

	dirPath, err := ms.subtitleService.CreateSubtitlesDirectory(movieID, baseDir)
	if err != nil {
		Logger.Error(fmt.Sprintf("Failed to create subtitles directory for movie %d: %v", movieID, err))
		return []string{}
	}

	Logger.Info(fmt.Sprintf("Downloading subtitles for movie %d from subdl.com", movieID))
	downloadedCount := ms.subtitleService.DownloadSubtitles(movieID, dirPath)
	Logger.Info(fmt.Sprintf("Downloaded %d subtitle(s) for movie %d", downloadedCount, movieID))

	srtFiles, err := FindFilesWithExtension(dirPath, "srt")
	if err != nil {
		Logger.Error(fmt.Sprintf("Error finding downloaded SRT files: %v", err))
		return []string{}
	}

	return srtFiles
}

func (ms *MovieService) findAndDownloadMovie(movieID int) (*models.TorrentDownload, error) {
	ms.updateStreamStatus(movieID, "searching", "Fetching movie information", map[string]interface{}{
		"step": "fetch_details",
	})

	s, _ := ms.GetSource("tmdb")

	details, err := s.GetIMDbID(movieID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch movie details: %w", err)
	}

	ms.updateStreamStatus(movieID, "searching", "Searching torrent sources", map[string]interface{}{
		"step":         "search_torrents",
		"imdb_id":      details.IMDbID,
		"title":        details.Title,
		"release_date": details.ReleaseDate,
	})

	torrents, err := ms.searchTorrentsByIMDb(*details, 30*time.Second)
	if err != nil {
		Logger.Error(fmt.Sprintf("Error searching torrents: %v", err))
		return nil, fmt.Errorf("failed to search torrents: %w", err)
	}

	ms.updateStreamStatus(movieID, "searching", fmt.Sprintf("Found %d torrent(s)", len(torrents)), map[string]interface{}{
		"step":          "torrents_found",
		"torrent_count": len(torrents),
	})

	if len(torrents) == 0 {
		return nil, fmt.Errorf("no suitable torrent found")
	}

	bestTorrent := &torrents[0]

	ms.updateStreamStatus(movieID, "searching", "Selected best torrent", map[string]interface{}{
		"step": "torrent_selected",
		"name": bestTorrent.Name,
	})

	download, err := ms.torrentService.GetOrStartDownload(movieID, bestTorrent.InfoHash)
	if err != nil {
		ms.updateStreamStatus(movieID, "error", "Failed to start download: "+err.Error(), nil)
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
			ms.updateStreamStatus(movieID, "downloading", "Download ready for streaming", map[string]interface{}{
				"step":             "download_ready",
				"downloadProgress": progress,
				"fileName":         filepath.Base(filePath),
			})
			return download, nil
		}

		progressChanged := progress-lastProgress >= 1.0
		if progressChanged || waitCount%5 == 0 {
			if progressChanged && progress > 0 {
				ms.updateStreamStatus(movieID, "downloading", fmt.Sprintf("Downloading: %.1f%% complete", progress), map[string]interface{}{
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

func (ms *MovieService) getTorrentMovieDetails(activeDownload *models.TorrentDownload) (string, *torrent.File, string, float64) {
	activeDownload.Mu.RLock()
	filePath := activeDownload.FilePath
	videoFile := activeDownload.VideoFile
	status := activeDownload.Status
	progress := activeDownload.Progress
	activeDownload.Mu.RUnlock()

	return filePath, videoFile, status, progress
}

func (ms *MovieService) waitUntilVideoFileIsReady(
	activeDownload *models.TorrentDownload, filePath string,
	videoFile *torrent.File,
	status string,
) {
	if !(status == "completed" && filePath != "" && videoFile == nil) {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

	waitLoop:
		for {
			_, videoFile, _, _ = ms.getTorrentMovieDetails(activeDownload)

			if videoFile != nil {
				break waitLoop
			}

			<-ticker.C
		}
	}
}

func (ms *MovieService) convertSubtitlesToHLS(srtFiles []string, hlsOutputDir string) error {
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
			Logger.Error(fmt.Sprintf("Failed to create subtitle directory for %s: %v", lang, err))
			continue
		}

		vttFile := filepath.Join(vttTempDir, fmt.Sprintf("%s.vtt", lang))
		vttFileHandle, err := os.Create(vttFile)
		if err != nil {
			Logger.Error(fmt.Sprintf("Failed to create VTT file for %s: %v", lang, err))
			continue
		}

		ConvertSRTtoVTT(srtFile, vttFileHandle)
		vttFileHandle.Close()

		if _, err := os.Stat(vttFile); os.IsNotExist(err) {
			Logger.Error(fmt.Sprintf("VTT file was not created for %s", lang))
			continue
		}

		destVttFile := filepath.Join(langSubsDir, "subtitle.vtt")
		if err := utils.CopyFile(vttFile, destVttFile); err != nil {
			Logger.Error(fmt.Sprintf("Failed to copy VTT file for %s: %v", lang, err))
			continue
		}

		playlistPath := filepath.Join(langSubsDir, "playlist.m3u8")
		mediaPlaylist, err := m3u8.NewMediaPlaylist(1, 1)
		if err != nil {
			Logger.Error(fmt.Sprintf("Failed to create media playlist for %s: %v", lang, err))
			continue
		}

		mediaPlaylist.MediaType = m3u8.VOD
		mediaPlaylist.SetVersion(3)

		if err := mediaPlaylist.Append("subtitle.vtt", 0.0, ""); err != nil {
			Logger.Error(fmt.Sprintf("Failed to append segment to playlist for %s: %v", lang, err))
			continue
		}
		mediaPlaylist.Close()

		playlistFile, err := os.Create(playlistPath)
		if err != nil {
			Logger.Error(fmt.Sprintf("Failed to create playlist file for %s: %v", lang, err))
			continue
		}
		playlistFile.WriteString(mediaPlaylist.String())
		playlistFile.Close()

		Logger.Info(fmt.Sprintf("Successfully converted subtitles for language: %s", lang))
	}

	return nil
}

func (ms *MovieService) createMasterPlaylist(hlsOutputDir string, srtFiles []string) (*m3u8.MasterPlaylist, error) {
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
						Name:       utils.GetLanguageLabel(lang),
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

	for _, quality := range VideoTranscoderConf.Qualities {
		if !quality.Enabled {
			continue
		}

		uri := fmt.Sprintf("%s/playlist.m3u8", quality.Name)
		params := m3u8.VariantParams{
			Bandwidth:    utils.ParseBandwidth(quality.VideoBitrate),
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

func (ms *MovieService) tryFFmpegTranscodingWithPlaylist(
	activeDownload *models.TorrentDownload,
	movieID int,
	hlsOutputDir string,
	masterPlaylist *m3u8.MasterPlaylist,
) {
	retryDelay := 10 * time.Second
	attempt := 0

	for {
		attempt++

		_, videoFile, _, _ := ms.getTorrentMovieDetails(activeDownload)

		if videoFile == nil {
			time.Sleep(retryDelay)
			continue
		}

		reader := videoFile.NewReader()
		reader.SetResponsive()                // Blocks until pieces are complete
		reader.SetReadahead(10 * 1024 * 1024) // 10MB read-ahead

		err := ms.runFFmpegTranscoding(reader, hlsOutputDir)
		reader.Close()

		if err == nil {
			var downloadedMovie models.DownloadedMovie
			if err := ms.db.Where("movie_id = ?", movieID).First(&downloadedMovie).Error; err == nil {

				outputDir := activeDownload.RootDir
				// log.Printf("Cleaning up torrent files for movie %d dir %s", movieID, outputDir)
				err := ms.torrentService.RemoveTorrentFiles(movieID, downloadedMovie.Quality, outputDir)
				if err != nil {
					// log.Printf("Error cleaning up torrent files for movie %d: %v", movieID, err)
				}
			}

			ms.MasterPlaylists.Store(movieID, masterPlaylist)

			ms.updateStreamStatus(movieID, "ready", "Stream is ready to play", map[string]interface{}{
				"transcodingStatus": "ready",
				"masterPlaylist":    masterPlaylist,
			})
			break
		}

		ms.updateStreamStatus(movieID, "transcoding", fmt.Sprintf("Transcoding attempt %d failed, retrying...", attempt), map[string]interface{}{
			"transcodingStatus": "retrying",
			"attempt":           attempt,
			"error":             err.Error(),
		})

		_, _, status, progress := ms.getTorrentMovieDetails(activeDownload)
		if status == "completed" || progress >= 100.0 {
			ms.updateStreamStatus(movieID, "error", "Transcoding failed after download completed", map[string]interface{}{
				"transcodingStatus": "failed",
				"lastError":         err.Error(),
			})
			break
		}

		time.Sleep(retryDelay)
	}
}

func (ms *MovieService) runFFmpegTranscoding(reader io.Reader, hlsOutputDir string) error {
	var args []string

	args = append(args,
		"-fflags", "+genpts+igndts+discardcorrupt",
		"-err_detect", "ignore_err",
		"-i", "pipe:0")

	args = append(args,
		"-c:v", "libx264",
		"-preset", VideoTranscoderConf.Encoding.Preset,
		"-crf", fmt.Sprintf("%d", VideoTranscoderConf.Encoding.CRF),
		"-g", fmt.Sprintf("%d", VideoTranscoderConf.Encoding.GOPSize),
		"-keyint_min", fmt.Sprintf("%d", VideoTranscoderConf.Encoding.KeyintMin),
		"-sc_threshold", fmt.Sprintf("%d", VideoTranscoderConf.Encoding.SCThreshold),
	)

	variantIndex := 0
	for _, quality := range VideoTranscoderConf.Qualities {
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
			fmt.Sprintf("-b:a:%d", variantIndex), VideoTranscoderConf.Encoding.AudioBitrate,
			fmt.Sprintf("-ar:%d", variantIndex), fmt.Sprintf("%d", VideoTranscoderConf.Encoding.AudioSampleRate),
		)

		if VideoTranscoderConf.Encoding.AudioChannels > 0 {
			args = append(args,
				fmt.Sprintf("-ac:%d", variantIndex), fmt.Sprintf("%d", VideoTranscoderConf.Encoding.AudioChannels),
			)
		}

		variantIndex++
	}

	args = append(args,
		"-f", "hls",
		"-hls_time", fmt.Sprintf("%d", VideoTranscoderConf.Output.SegmentTime),
		"-hls_playlist_type", "event",
		"-hls_flags", "temp_file+independent_segments+omit_endlist",
		"-hls_list_size", "0",
	)

	var varStreamMap strings.Builder
	idx := 0
	for _, quality := range VideoTranscoderConf.Qualities {
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
	// cmd.Stdout = os.Stdout
	// cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}
