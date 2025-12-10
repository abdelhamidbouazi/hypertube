package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"server/internal/models"
	"strings"
	"sync"
	"time"

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
	UserSegmentTrack   sync.Map // map[string]map[int]string - "userID:movieID" -> last segment visited
	SegmentFormatParse string
	SearchSources      map[string]Source
	db                 *gorm.DB
}

func NewMovieService(tmdbKey, omdbKey, watchModeKey string, db *gorm.DB) *MovieService {
	ms := &MovieService{
		apiKey:             tmdbKey,
		omdbKey:            omdbKey,
		client:             &http.Client{Timeout: 10 * time.Second},
		SegmentFormatParse: VideoTranscoderConf.Output.SegmentFilenameFormat,
		torrentSources: []string{
			"1337x",
			"yts",
		},
		db: db,
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

type DiscoverMoviesResp struct {
	ID          int     `json:"id"`
	Title       string  `json:"title"`
	ReleaseDate string  `json:"release_date"`
	PosterPath  string  `json:"poster_path"`
	Overview    string  `json:"overview"`
	Language    string  `json:"original_language,omitempty"`
	VoteAverage float64 `json:"vote_average"`
	GenreIDs    []int   `json:"genre_ids,omitempty"`
	IsWatched   bool    `json:"isWatched"`
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

type TorrentSearchResult struct {
	InfoHash string
	Name     string
}

func (ms *MovieService) SearchTorrentsByIMDb(movie models.MovieDetails, metadataTimeout time.Duration) ([]TorrentSearchResult, error) {
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

	type Torrent struct {
		Name          string `json:"name"`
		Title         string `json:"title"`
		InfoHash      string `json:"infoHash"`
		FileIdx       int    `json:"fileIdx"`
		BehaviorHints struct {
			BingeGroup string `json:"bingeGroup"`
			Filename   string `json:"filename"`
		} `json:"behaviorHints"`
		Sources []string `json:"sources"`
	}

	var data struct {
		Streams []Torrent `json:"streams"`
	}

	if err := json.NewDecoder(res.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to decode torrentio response for IMDb ID %s: %v", imdbID, err)
	}

	var torrents []Torrent = data.Streams

	for _, t := range torrents {
		results = append(results, TorrentSearchResult{InfoHash: t.InfoHash, Name: t.Title})
	}

	return results, nil
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
