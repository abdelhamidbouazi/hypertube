package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"server/internal/models"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/anacrolix/torrent"
	"github.com/deflix-tv/imdb2torrent"
	"github.com/grafov/m3u8"
	"go.uber.org/zap"
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
	UserSegmentTrack   sync.Map // map[string]map[int]string - "userID:movieID" -> last segment visited
	SegmentFormatParse string
	SearchSources      map[string]Source
}

func NewMovieService(tmdbKey, omdbKey, watchModeKey string) *MovieService {
	ms := &MovieService{
		apiKey:             tmdbKey,
		omdbKey:            omdbKey,
		client:             &http.Client{Timeout: 10 * time.Second},
		SegmentFormatParse: VideoTranscoderConf.Output.SegmentFilenameFormat,
		torrentSources: []string{
			"1337x",
			"yts",
		},
	}

	ms.SearchSources = map[string]Source{
		"tmdb": NewTMDB(tmdbKey, omdbKey, ms.genreCacheTime, ms.client),
		"omdb": NewOMDB(omdbKey, ms.genreCacheTime, ms.client),
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
func (ms *MovieService) DiscoverMovies(p MovieDiscoverParams, source string) ([]models.Movie, error) {
	var movies []models.Movie
	var err error

	if source != "" {
		src, err := ms.GetSource(source)
		if err != nil {
			return nil, err
		}
		return src.DiscoverMovies(p)
	}

	for _, s := range ms.SearchSources {
		movies, err = s.DiscoverMovies(p)
		if err == nil && len(movies) > 0 {
			return movies, nil
		}
	}

	return movies, err
}

func normalizeGenreName(name string) string {
	// lower-case and trim spaces for matching
	return strings.TrimSpace(strings.ToLower(name))
}

func joinCSV(items []string) string {
	if len(items) == 0 {
		return ""
	}
	return strings.Join(items, ",")
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

// movieMetaGetter implements imdb2torrent.MetaGetter
type movieMetaGetter struct {
	movie models.MovieDetails
}

// GetMovieSimple implements imdb2torrent.MetaGetter
func (m *movieMetaGetter) GetMovieSimple(ctx context.Context, imdbID string) (imdb2torrent.Meta, error) {
	year := 0
	if len(m.movie.ReleaseDate) >= 4 {
		y, err := strconv.Atoi(m.movie.ReleaseDate[:4])
		if err == nil {
			year = y
		}
	}

	Logger.Info(fmt.Sprintf("MetaGetter: Title='%s', Year=%d for IMDb ID %s", m.movie.Title, year, imdbID))

	return imdb2torrent.Meta{
		Title: m.movie.Title,
		Year:  year,
	}, nil
}

// GetTVShowSimple implements imdb2torrent.MetaGetter
func (m *movieMetaGetter) GetTVShowSimple(ctx context.Context, imdbID string, season int, episode int) (imdb2torrent.Meta, error) {
	return imdb2torrent.Meta{}, nil
}

func (ms *MovieService) SearchTorrentsByIMDb(movie models.MovieDetails, metadataTimeout time.Duration) ([]models.TorrentResult, error) {
	imdbID := movie.IMDbID
	var results []models.TorrentResult

	logger := zap.NewNop()
	cache := imdb2torrent.NewInMemoryCache()

	metaGetter := &movieMetaGetter{
		movie: movie,
	}

	ytsGGOptions := imdb2torrent.DefaultYTSclientOpts
	ytsGGOptions.BaseURL = "https://yts.gg"
	ytsGGClient := imdb2torrent.NewYTSclient(
		ytsGGOptions,
		cache,
		logger,
		false,
	)

	ytsLTOptions := imdb2torrent.DefaultYTSclientOpts
	ytsLTOptions.BaseURL = "https://yts.lt"
	ytsLTClient := imdb2torrent.NewYTSclient(
		ytsLTOptions,
		cache,
		logger,
		false,
	)

	ytsAMOptions := imdb2torrent.DefaultYTSclientOpts
	ytsAMOptions.BaseURL = "https://yts.am"
	ytsAMClient := imdb2torrent.NewYTSclient(
		ytsAMOptions,
		cache,
		logger,
		false,
	)

	ytsAGOptions := imdb2torrent.DefaultYTSclientOpts
	ytsAGOptions.BaseURL = "https://yts.ag"
	ytsAGClient := imdb2torrent.NewYTSclient(
		ytsAGOptions,
		cache,
		logger,
		false,
	)

	ibitOptions := imdb2torrent.DefaultIbitClientOpts

	ibitOptions.BaseURL = "https://ibit.unblockedproxy.biz"

	ibitClient := imdb2torrent.NewIbitClient(
		ibitOptions,
		cache,
		logger,
		false,
	)

	tpbClient, _ := imdb2torrent.NewTPBclient(
		imdb2torrent.DefaultTPBclientOpts,
		cache,
		metaGetter,
		logger,
		false,
	)

	leetxClient := imdb2torrent.NewLeetxClient(
		imdb2torrent.DefaultLeetxClientOpts,
		cache,
		metaGetter,
		logger,
		false,
	)

	rarbgClient := imdb2torrent.NewRARBGclient(
		imdb2torrent.DefaultRARBGclientOpts, cache, logger, false)

	siteClients := map[string]imdb2torrent.MagnetSearcher{
		"YTS-GG": ytsGGClient,
		"YTS-LT": ytsLTClient,
		"YTS-AM": ytsAMClient,
		"YTS-AG": ytsAGClient,
		"ibit":   ibitClient,
		"TPB":    tpbClient,
		"1337x":  leetxClient,
		"rarbg":  rarbgClient,
	}

	client := imdb2torrent.NewClient(
		siteClients,
		5*time.Minute,
		logger,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	torrents, err := client.FindMovie(ctx, imdbID)
	if err != nil {
		Logger.Error(fmt.Sprintf("Failed to find torrents for IMDb ID %s: %v", imdbID, err))
		return nil, err
	}

	Logger.Info(fmt.Sprintf("Found %d torrents for IMDb ID %s", len(torrents), imdbID))

	if len(torrents) == 0 {
		return results, nil
	}

	cfg := torrent.NewDefaultClientConfig()
	cfg.DataDir = "/tmp/torrent-metadata"
	cfg.NoDHT = false

	torrentClient, err := torrent.NewClient(cfg)
	if err != nil {
		Logger.Error(fmt.Sprintf("Failed to create torrent client: %v", err))
	}
	defer func() {
		if torrentClient != nil {
			torrentClient.Close()
		}
	}()

	type metadataResult struct {
		index    int
		size     string
		seeders  int
		leechers int
	}
	resultChan := make(chan metadataResult, len(torrents))
	metaCtx, metaCancel := context.WithTimeout(context.Background(), metadataTimeout)
	defer metaCancel()

	for i, t := range torrents {
		go func(idx int, torr imdb2torrent.Result) {
			if torrentClient != nil {
				sizeBytes, s, l, err := fetchTorrentMetadata(metaCtx, torrentClient, torr.MagnetURL)
				if err == nil {
					resultChan <- metadataResult{
						index:    idx,
						size:     formatBytes(sizeBytes),
						seeders:  s,
						leechers: l,
					}
				} else {
					Logger.Error(fmt.Sprintf("Failed to fetch metadata for torrent %s: %v", torr.Title, err))
				}
			}
		}(i, t)
	}

	metadataMap := make(map[int]metadataResult)
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	var firstResultReceived bool

collectLoop:
	for {
		select {
		case result := <-resultChan:
			metadataMap[result.index] = result

			if !firstResultReceived {
				firstResultReceived = true
			}

			if len(metadataMap) == len(torrents) {
				break collectLoop
			}
		case <-ticker.C:
		case <-metaCtx.Done():
			if firstResultReceived {
				break collectLoop
			}
		}
	}

	// Build results with metadata
	for i, t := range torrents {
		size := ""
		seeders := 0
		leechers := 0

		if metadata, ok := metadataMap[i]; ok {
			size = metadata.size
			seeders = metadata.seeders
			leechers = metadata.leechers
		}

		result := models.TorrentResult{
			Name:     t.Title,
			Magnet:   t.MagnetURL,
			Size:     size,
			Seeders:  seeders,
			Leechers: leechers,
			Quality:  t.Quality,
		}

		results = append(results, result)
	}

	sortTorrentsByRatio(results)

	return results, nil
}

func fetchTorrentMetadata(ctx context.Context, client *torrent.Client, magnetURL string) (int64, int, int, error) {
	t, err := client.AddMagnet(magnetURL)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("failed to add magnet: %w", err)
	}
	defer t.Drop()

	// Wait for metadata with timeout
	select {
	case <-t.GotInfo():
		size := t.Length()

		// Get peer stats
		stats := t.Stats()
		seeders := stats.ConnectedSeeders
		leechers := stats.ActivePeers - stats.ConnectedSeeders

		return size, seeders, leechers, nil
	case <-ctx.Done():
		return 0, 0, 0, fmt.Errorf("timeout fetching metadata")
	}
}

// formatBytes converts bytes to human-readable format
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func sortTorrentsByRatio(torrents []models.TorrentResult) {
	sort.Slice(torrents, func(i, j int) bool {
		if torrents[i].Seeders == 0 && torrents[i].Leechers == 0 {
			return false
		}
		if torrents[j].Seeders == 0 && torrents[j].Leechers == 0 {
			return true
		}

		ratioI := float64(torrents[i].Seeders) / float64(torrents[i].Leechers+1)
		ratioJ := float64(torrents[j].Seeders) / float64(torrents[j].Leechers+1)

		if ratioI == ratioJ {
			return torrents[i].Seeders > torrents[j].Seeders
		}

		return ratioI > ratioJ
	})
}

func (ms *MovieService) IsLastSegmentInVariantPlaylist(movieID int, segmentFilename string) bool {
	if cachedLastSegment, ok := ms.LastSegmentCache.Load(movieID); ok {
		return cachedLastSegment.(string) == segmentFilename
	}

	db := PostgresDB()
	if db == nil {
		return false
	}

	var downloadedMovie models.DownloadedMovie
	if err := db.Where("movie_id = ? AND transcoded = ?", movieID, true).First(&downloadedMovie).Error; err == nil {
		if downloadedMovie.LastSegment != "" {
			ms.LastSegmentCache.Store(movieID, downloadedMovie.LastSegment)
			return downloadedMovie.LastSegment == segmentFilename
		}
	}

	return false
}

func (ms *MovieService) GetLastSegmentFromPlaylist(playlistPath string) string {
	file, err := os.Open(playlistPath)
	if err != nil {
		return ""
	}
	defer file.Close()

	playlist, listType, err := m3u8.DecodeFrom(file, true)
	if err != nil || listType != m3u8.MEDIA {
		return ""
	}

	mediaPlaylist, ok := playlist.(*m3u8.MediaPlaylist)
	if !ok {
		return ""
	}

	var lastSegmentURI string
	for _, segment := range mediaPlaylist.Segments {
		if segment != nil && segment.URI != "" {
			lastSegmentURI = segment.URI
		}
	}

	return lastSegmentURI
}

func (ms *MovieService) TrackUserSegment(userID uint, movieID int, segmentFilename string) {
	key := fmt.Sprintf("%d:%d", userID, movieID)
	ms.UserSegmentTrack.Store(key, segmentFilename)
}

func (ms *MovieService) persistWatchHistoryWorker() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		ms.UserSegmentTrack.Range(func(key, value interface{}) bool {
			keyStr := key.(string)
			segmentFilename := value.(string)

			var userID uint
			var movieID int
			if _, err := fmt.Sscanf(keyStr, "%d:%d", &userID, &movieID); err != nil {
				return true
			}

			isLastSegment := ms.IsLastSegmentInVariantPlaylist(movieID, segmentFilename)
			lastSegmentFilename := ms.getLastSegmentFilename(movieID)

			progress := ms.calculateWatchProgress(movieID, segmentFilename, lastSegmentFilename, isLastSegment)

			db := PostgresDB()
			if db == nil {
				return true
			}

			var watchHistory models.WatchHistory
			result := db.Where("user_id = ? AND movie_id = ?", userID, movieID).First(&watchHistory)

			if result.Error == gorm.ErrRecordNotFound {
				watchHistory = models.WatchHistory{
					UserID:        userID,
					MovieID:       movieID,
					LastSegment:   segmentFilename,
					WatchProgress: progress,
					WatchCount:    0,
					WatchedAt:     time.Now(),
				}

				if isLastSegment {
					watchHistory.WatchCount = 1
				}

				db.Create(&watchHistory)
			} else if result.Error == nil {
				updates := map[string]interface{}{
					"last_segment":   segmentFilename,
					"watch_progress": progress,
					"watched_at":     time.Now(),
				}

				if isLastSegment && watchHistory.LastSegment != segmentFilename {
					updates["watch_count"] = gorm.Expr("watch_count + 1")
				}

				db.Model(&watchHistory).Updates(updates)
			}

			return true
		})
	}
}

func (ms *MovieService) cleanupOldHLSFilesWorker(hlsDir string) {
	ticker := time.NewTicker(4 * time.Second)
	defer ticker.Stop()

	for {
		ms.cleanupOldHLSFiles(hlsDir)
		<-ticker.C
	}
}

func (ms *MovieService) cleanupOldHLSFiles(hlsDir string) {
	db := PostgresDB()
	if db == nil {
		// log.Println("Database not available for HLS cleanup")
		return
	}

	type MovieLastWatched struct {
		MovieID   int
		WatchedAt time.Time
	}

	var oldMovies []MovieLastWatched
	oneMonthAgo := time.Now().AddDate(0, -1, 0)

	err := db.Model(&models.WatchHistory{}).
		Select("movie_id, MAX(watched_at) as watched_at").
		Group("movie_id").
		Having("MAX(watched_at) < ?", oneMonthAgo).
		Find(&oldMovies).Error
	if err != nil {
		// log.Printf("Error fetching old watched movies: %v", err)
		return
	}

	if len(oldMovies) == 0 {
		// log.Println("No old HLS files to cleanup")
		return
	}

	// log.Printf("Found %d movies not watched in over a month, cleaning up HLS files", len(oldMovies))

	for _, movie := range oldMovies {
		movieHLSDir := filepath.Join(hlsDir, fmt.Sprintf("%d", movie.MovieID))

		if _, err := os.Stat(movieHLSDir); err == nil {
			// log.Printf("Removing HLS directory for movie %d (last watched: %s)", movie.MovieID, movie.WatchedAt.Format("2006-01-02"))
			if err := os.RemoveAll(movieHLSDir); err != nil {
				// log.Printf("Error removing HLS directory %s: %v", movieHLSDir, err)
			} else {
				// log.Printf("Successfully removed HLS directory for movie %d", movie.MovieID)
			}
		} else {
			// log.Printf("HLS directory not found for movie %d: %s", movie.MovieID, hlsDir)
		}
	}
}

func (ms *MovieService) getLastSegmentFilename(movieID int) string {
	if cachedLastSegment, ok := ms.LastSegmentCache.Load(movieID); ok {
		lastSegment := cachedLastSegment.(string)
		return lastSegment
	}

	db := PostgresDB()
	if db == nil {
		return ""
	}

	var downloadedMovie models.DownloadedMovie
	if err := db.Where("movie_id = ? AND transcoded = ?", movieID, true).First(&downloadedMovie).Error; err != nil {
		return ""
	}

	if downloadedMovie.LastSegment == "" {
		return ""
	}

	ms.LastSegmentCache.Store(movieID, downloadedMovie.LastSegment)

	return downloadedMovie.LastSegment
}

func (ms *MovieService) calculateWatchProgress(movieID int, segmentFilename string, lastSegmentFilename string, isLastSegment bool) float64 {
	if isLastSegment {
		return 100.0
	}

	if lastSegmentFilename != "" {
		segmentNum := 0
		if _, err := fmt.Sscanf(filepath.Base(segmentFilename), ms.SegmentFormatParse, &segmentNum); err != nil {
			return 0.0
		}

		lastSegmentNum := 0
		if _, err := fmt.Sscanf(filepath.Base(lastSegmentFilename), ms.SegmentFormatParse, &lastSegmentNum); err != nil {
			return 0.0
		}

		if lastSegmentNum > 0 {
			progress := (float64(segmentNum) / float64(lastSegmentNum)) * 100.0
			if progress > 100.0 {
				progress = 100.0
			}
			return progress
		}
	}
	return 0.0
}
