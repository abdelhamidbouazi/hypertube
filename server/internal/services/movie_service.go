package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"server/internal/models"
	"strconv"
	"strings"
	"time"

	"github.com/anacrolix/torrent"
	"github.com/deflix-tv/imdb2torrent"
	"go.uber.org/zap"
)

type MovieService struct {
	apiKey         string
	omdbKey        string
	client         *http.Client
	torrentSources []string
	genreCache     map[string]int
	genreCacheTime time.Time
}

func NewMovieService(tmdbKey, omdbKey string) *MovieService {
	return &MovieService{
		apiKey:  tmdbKey,
		omdbKey: omdbKey,
		client:  &http.Client{Timeout: 10 * time.Second},
		torrentSources: []string{
			"1337x",
			"yts",
		},
		genreCache: make(map[string]int),
	}
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
	case "name":
		sortBy = "original_title.asc"
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

	if time.Since(ms.genreCacheTime) > 24*time.Hour || len(ms.genreCache) == 0 {
		if err := ms.refreshGenreCache(); err != nil {
			return "", err
		}
	}

	ids := []string{}
	for _, name := range genres {
		key := normalizeGenreName(name)
		if id, ok := ms.genreCache[key]; ok {
			ids = append(ids, strconv.Itoa(id))
		} else {
			log.Printf("unknown genre name: %s", name)
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

	cache := make(map[string]int)
	for _, g := range payload.Genres {
		cache[normalizeGenreName(g.Name)] = g.ID
	}
	ms.genreCache = cache
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

func (ms *MovieService) SearchTorrentsByIMDb(movie models.MovieDetails) ([]models.TorrentResult, error) {
	imdbID := movie.IMDbID
	Logger.Info(fmt.Sprintf("Searching torrents for IMDb ID %s", imdbID))
	var results []models.TorrentResult

	logger := zap.NewNop()
	cache := imdb2torrent.NewInMemoryCache()

	ytsOptions := imdb2torrent.DefaultYTSclientOpts

	ytsOptions.BaseURL = "https://yts.lt"

	ytsClient := imdb2torrent.NewYTSclient(
		ytsOptions,
		cache,
		logger,
		false,
	)

	siteClients := map[string]imdb2torrent.MagnetSearcher{
		"YTS": ytsClient,
	}

	client := imdb2torrent.NewClient(
		siteClients,
		5*time.Second,
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

	for _, t := range torrents {
		size := ""
		seeders := 0
		leechers := 0

		if torrentClient != nil {
			sizeBytes, s, l, err := fetchTorrentMetadata(torrentClient, t.MagnetURL)
			if err == nil {
				size = formatBytes(sizeBytes)
				seeders = s
				leechers = l
			} else {
				Logger.Error(fmt.Sprintf("Failed to fetch metadata for torrent %s: %v", t.Title, err))
			}
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

	return results, nil
}

// fetchTorrentMetadata fetches size and peer info from magnet link
func fetchTorrentMetadata(client *torrent.Client, magnetURL string) (int64, int, int, error) {
	t, err := client.AddMagnet(magnetURL)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("failed to add magnet: %w", err)
	}
	defer t.Drop() // Remove torrent from client when done

	// Wait for metadata with timeout
	select {
	case <-t.GotInfo():
		size := t.Length()

		// Get peer stats
		stats := t.Stats()
		seeders := stats.ConnectedSeeders
		leechers := stats.ActivePeers - stats.ConnectedSeeders

		return size, seeders, leechers, nil
	case <-time.After(30 * time.Second):
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

// Helper function to get movie info from IMDb ID
func (ms *MovieService) getMovieInfoByIMDb(imdbID string) (*struct {
	Title string
	Year  int
}, error,
) {
	// Use OMDb API (you can replace with your existing movie service)
	apiURL := fmt.Sprintf("https://www.omdbapi.com/?i=%s&apikey=YOUR_OMDB_KEY", imdbID)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := ms.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var omdbResp struct {
		Title string `json:"Title"`
		Year  string `json:"Year"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&omdbResp); err != nil {
		return nil, err
	}

	year, _ := strconv.Atoi(omdbResp.Year)

	return &struct {
		Title string
		Year  int
	}{
		Title: omdbResp.Title,
		Year:  year,
	}, nil
}

// Helper function to extract quality from torrent name
func (ms *MovieService) extractQuality(name string) string {
	name = strings.ToUpper(name)
	qualities := []string{"2160P", "4K", "1080P", "720P", "480P", "BLURAY", "WEBRIP", "HDTV"}
	for _, q := range qualities {
		if strings.Contains(name, q) {
			return q
		}
	}
	return "Unknown"
}

// Helper function to format bytes to human readable size
func (ms *MovieService) formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
