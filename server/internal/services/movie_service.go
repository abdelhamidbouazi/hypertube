package services

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
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
	apiKey           string
	omdbKey          string
	client           *http.Client
	torrentSources   []string
	genreCache       sync.Map // map[string]int
	genreCacheTime   time.Time
	StreamStatus     sync.Map // map[int]map[string]interface{}
	MasterPlaylists  sync.Map // map[int]*m3u8.MasterPlaylist - movieID -> master playlist
	LastSegmentCache sync.Map // map[int]string - movieID -> last segment filename
	UserSegmentTrack sync.Map // map[string]map[int]string - "userID:movieID" -> last segment visited
}

func NewMovieService(tmdbKey, omdbKey string) *MovieService {
	ms := &MovieService{
		apiKey:  tmdbKey,
		omdbKey: omdbKey,
		client:  &http.Client{Timeout: 10 * time.Second},
		torrentSources: []string{
			"1337x",
			"yts",
		},
	}

	go ms.persistWatchHistoryWorker()
	go ms.cleanupOldHLSFilesWorker(Conf.STREAMING.HLSOutputDir)

	return ms
}

// MovieDiscoverParams encapsulates supported filters for discovering movies
type MovieDiscoverParams struct {
	Genres    []string
	YearFrom  *int
	YearTo    *int
	MinRating *float64
	Sort      string
	Page      int
}

// DiscoverMovies calls TMDB discover endpoint with filters
func (ms *MovieService) DiscoverMovies(p MovieDiscoverParams) ([]models.Movie, error) {
	if p.Page < 1 {
		p.Page = 1
	}

	baseURL := "https://api.themoviedb.org/3/discover/movie"
	params := url.Values{}
	params.Add("api_key", ms.apiKey)
	params.Add("page", strconv.Itoa(p.Page))

	// Sorting
	sortBy := "popularity.desc"
	switch p.Sort {
	case "year", "year_desc":
		sortBy = "primary_release_date.desc"
	case "year_asc":
		sortBy = "primary_release_date.asc"
	case "rating":
		sortBy = "vote_average.desc"
	}
	params.Add("sort_by", sortBy)

	// Year range via dates
	if p.YearFrom != nil {
		params.Add("primary_release_date.gte", fmt.Sprintf("%04d-01-01", *p.YearFrom))
	}
	if p.YearTo != nil {
		params.Add("primary_release_date.lte", fmt.Sprintf("%04d-12-31", *p.YearTo))
	}

	// Min rating
	if p.MinRating != nil {
		params.Add("vote_average.gte", strconv.FormatFloat(*p.MinRating, 'f', -1, 64))
	}

	if len(p.Genres) > 0 {
		ids, err := ms.getGenreIDs(p.Genres)
		if err == nil && len(ids) > 0 {
			params.Add("with_genres", ids)
		}
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

func (ms *MovieService) getGenreIDs(genres []string) (string, error) {
	numeric := true
	for _, g := range genres {
		if _, err := strconv.Atoi(g); err != nil {
			numeric = false
			break
		}
	}
	if numeric {
		return joinCSV(genres), nil
	}

	// Check if we need to refresh the cache
	// Note: We can't use len() on sync.Map, so we'll check cache time only
	if time.Since(ms.genreCacheTime) > 24*time.Hour {
		if err := ms.refreshGenreCache(); err != nil {
			return "", err
		}
	}

	ids := []string{}
	for _, name := range genres {
		key := normalizeGenreName(name)
		if id, ok := ms.genreCache.Load(key); ok {
			ids = append(ids, strconv.Itoa(id.(int)))
		} else {
			// log.Printf("unknown genre name: %s", name)
		}
	}
	return joinCSV(ids), nil
}

func (ms *MovieService) refreshGenreCache() error {
	baseURL := "https://api.themoviedb.org/3/genre/movie/list"
	params := url.Values{}
	params.Add("api_key", ms.apiKey)
	fullURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	resp, err := ms.client.Get(fullURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var payload struct {
		Genres []struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
		} `json:"genres"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return err
	}

	for _, g := range payload.Genres {
		ms.genreCache.Store(normalizeGenreName(g.Name), g.ID)
	}
	ms.genreCacheTime = time.Now()
	return nil
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

func (ms *MovieService) GetMovieDetails(movieID string) (*models.MovieDetails, error) {
	baseURL := fmt.Sprintf("https://api.themoviedb.org/3/movie/%s", movieID)
	params := url.Values{}
	params.Add("api_key", ms.apiKey)
	params.Add("append_to_response", "credits,external_ids")

	fullURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	resp, err := ms.client.Get(fullURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var tmdbResp map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&tmdbResp); err != nil {
		return nil, err
	}

	details := ms.parseTMDBResponse(tmdbResp)

	if imdbID, ok := tmdbResp["imdb_id"].(string); ok && imdbID != "" {
		details.IMDbID = imdbID
		if omdbData := ms.getOMDBData(imdbID); omdbData != nil {
			if rating, ok := omdbData["imdbRating"].(string); ok {
				if r, err := strconv.ParseFloat(rating, 64); err == nil {
					details.VoteAverage = r
				}
			}
		}
	}

	return details, nil
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

// GetMovies returns a default list of movies (popular page 1)
func (ms *MovieService) GetMovies() ([]models.Movie, error) {
	return ms.GetPopularMovies(1)
}

// GetPopularMovies fetches a paginated list of popular movies from TMDB
func (ms *MovieService) GetPopularMovies(page int) ([]models.Movie, error) {
	if page < 1 {
		page = 1
	}
	baseURL := "https://api.themoviedb.org/3/movie/popular"
	params := url.Values{}
	params.Add("api_key", ms.apiKey)
	params.Add("page", strconv.Itoa(page))

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

// GetRandomMovies fetches a page from TMDB discover endpoint and returns a shuffled subset
func (ms *MovieService) GetRandomMovies(page int, pageSize int) ([]models.Movie, error) {
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	baseURL := "https://api.themoviedb.org/3/discover/movie"
	params := url.Values{}
	params.Add("api_key", ms.apiKey)
	params.Add("sort_by", "popularity.desc")
	params.Add("page", strconv.Itoa(page))

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

	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(movies), func(i, j int) { movies[i], movies[j] = movies[j], movies[i] })
	if pageSize < len(movies) {
		movies = movies[:pageSize]
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

	ytsOptions := imdb2torrent.DefaultYTSclientOpts

	ytsOptions.BaseURL = "https://yts.gg"

	ytsClient := imdb2torrent.NewYTSclient(
		ytsOptions,
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
		"YTS":   ytsClient,
		"ibit":  ibitClient,
		"TPB":   tpbClient,
		"1337x": leetxClient,
		"rarbg": rarbgClient,
	}

	client := imdb2torrent.NewClient(
		siteClients,
		15*time.Second,
		logger,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
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

func (ms *MovieService) GetIMDbIDFromTMDb(movieID int) (*models.MovieDetails, error) {
	Logger.Info(fmt.Sprintf("Fetching IMDb ID for TMDB movie %d", movieID))
	movieIDStr := strconv.Itoa(movieID)

	details, err := ms.GetMovieDetails(movieIDStr)
	if err != nil {
		Logger.Error(fmt.Sprintf("Failed to get movie details for TMDB movie %d: %v", movieID, err))
		return nil, fmt.Errorf("failed to get movie details: %w", err)
	}

	if details.IMDbID != "" {
		Logger.Info(fmt.Sprintf("Found IMDb ID %s for TMDB movie %d", details.IMDbID, movieID))
		return details, nil
	}

	Logger.Info(fmt.Sprintf("Fetching external IDs for TMDB movie %d", movieID))
	baseURL := fmt.Sprintf("https://api.themoviedb.org/3/movie/%d/external_ids", movieID)
	params := url.Values{}
	params.Add("api_key", ms.apiKey)

	fullURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	resp, err := ms.client.Get(fullURL)
	if err != nil {
		Logger.Error(fmt.Sprintf("Failed to fetch external IDs for TMDB movie %d: %v", movieID, err))
		return nil, err
	}
	defer resp.Body.Close()

	Logger.Info(fmt.Sprintf("Decoding external IDs for TMDB movie %d", movieID))
	var externalIDs struct {
		IMDbID string `json:"imdb_id"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&externalIDs); err != nil {
		Logger.Error(fmt.Sprintf("Failed to decode external IDs for TMDB movie %d: %v", movieID, err))
		return nil, err
	}

	if externalIDs.IMDbID == "" {
		Logger.Error(fmt.Sprintf("No IMDb ID found for TMDB movie %d", movieID))
		return nil, fmt.Errorf("no IMDb ID found for TMDB movie %d", movieID)
	}

	Logger.Info(fmt.Sprintf("Setting IMDb ID %s for TMDB movie %d", externalIDs.IMDbID, movieID))
	details.IMDbID = externalIDs.IMDbID

	return details, nil
}

func (ms *MovieService) parseTMDBResponse(data map[string]interface{}) *models.MovieDetails {
	details := &models.MovieDetails{}

	if v, ok := data["id"].(float64); ok {
		details.ID = int(v)
	}
	if v, ok := data["title"].(string); ok {
		details.Title = v
	}
	if v, ok := data["overview"].(string); ok {
		details.Overview = v
	}
	if v, ok := data["release_date"].(string); ok {
		details.ReleaseDate = v
	}
	if v, ok := data["runtime"].(float64); ok {
		details.Runtime = int(v)
	}
	if v, ok := data["poster_path"].(string); ok {
		details.PosterPath = v
	}
	if v, ok := data["backdrop_path"].(string); ok {
		details.BackdropPath = v
	}
	if v, ok := data["vote_average"].(float64); ok {
		details.VoteAverage = v
	}

	// Parse credits
	if credits, ok := data["credits"].(map[string]interface{}); ok {
		// Cast
		if cast, ok := credits["cast"].([]interface{}); ok {
			for i, c := range cast {
				if i >= 10 {
					break
				}
				if actor, ok := c.(map[string]interface{}); ok {
					castMember := models.Cast{}
					if v, ok := actor["id"].(float64); ok {
						castMember.ID = int(v)
					}
					if v, ok := actor["name"].(string); ok {
						castMember.Name = v
					}
					if v, ok := actor["character"].(string); ok {
						castMember.Character = v
					}
					if v, ok := actor["profile_path"].(string); ok {
						castMember.ProfilePath = v
					}
					details.Cast = append(details.Cast, castMember)
				}
			}
		}

		if crew, ok := credits["crew"].([]interface{}); ok {
			for _, c := range crew {
				if member, ok := c.(map[string]interface{}); ok {
					job, _ := member["job"].(string)
					name, _ := member["name"].(string)
					id, _ := member["id"].(float64)

					person := models.Person{ID: int(id), Name: name}

					if job == "Director" {
						details.Director = append(details.Director, person)
					} else if job == "Producer" || job == "Executive Producer" {
						details.Producer = append(details.Producer, person)
					}
				}
			}
		}
	}

	if genres, ok := data["genres"].([]interface{}); ok {
		for _, g := range genres {
			if genre, ok := g.(map[string]interface{}); ok {
				genreItem := models.Genre{}
				if v, ok := genre["id"].(float64); ok {
					genreItem.ID = int(v)
				}
				if v, ok := genre["name"].(string); ok {
					genreItem.Name = v
				}
				details.Genres = append(details.Genres, genreItem)
			}
		}
	}

	return details
}

func (ms *MovieService) getOMDBData(imdbID string) map[string]interface{} {
	url := fmt.Sprintf("http://www.omdbapi.com/?i=%s&apikey=%s", imdbID, ms.omdbKey)

	resp, err := ms.client.Get(url)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	var data map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&data)
	return data
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

			progress := ms.calculateWatchProgress(movieID, segmentFilename)

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
	oneMonthAgo := time.Now().AddDate(0, 0, 0).Add(-10 * time.Hour)

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

func (ms *MovieService) calculateWatchProgress(movieID int, segmentFilename string) float64 {
	var segmentNum int
	if _, err := fmt.Sscanf(filepath.Base(segmentFilename), "%d.ts", &segmentNum); err != nil {
		return 0.0
	}

	var lastSegment string
	if cachedLastSegment, ok := ms.LastSegmentCache.Load(movieID); ok {
		lastSegment = cachedLastSegment.(string)
	} else {
		db := PostgresDB()
		if db != nil {
			var downloadedMovie models.DownloadedMovie
			if err := db.Where("movie_id = ? AND transcoded = ?", movieID, true).First(&downloadedMovie).Error; err == nil {
				if downloadedMovie.LastSegment != "" {
					lastSegment = downloadedMovie.LastSegment
					ms.LastSegmentCache.Store(movieID, lastSegment)
				}
			}
		}
	}

	if lastSegment != "" {
		var lastSegmentNum int
		if _, err := fmt.Sscanf(filepath.Base(lastSegment), "%d.ts", &lastSegmentNum); err == nil && lastSegmentNum > 0 {
			progress := (float64(segmentNum) / float64(lastSegmentNum)) * 100.0
			if progress > 100.0 {
				progress = 100.0
			}
			return progress
		}
	}

	return 0.0
}
