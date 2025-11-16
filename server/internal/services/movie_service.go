package services

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"time"
	"strings"

	"server/internal/models"
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

func (ms *MovieService) SearchTorrents(query string, year string) ([]models.TorrentResult, error) {
	var results []models.TorrentResult

	ytsResults, err := ms.searchYTS(query, year)
	if err != nil {
		return nil, err
	}
	results = append(results, ytsResults...)

	return results, nil
}

func (ms *MovieService) SearchTorrentsByIMDb(imdbID string) ([]models.TorrentResult, error) {
	var results []models.TorrentResult

	// YTS supports IMDb ID search
	ytsResults, err := ms.searchYTSByIMDb(imdbID)
	if err != nil {
		log.Printf("Error searching YTS by IMDb: %v", err)
	} else {
		results = append(results, ytsResults...)
	}

	return results, nil
}

func (ms *MovieService) GetIMDbIDFromTMDb(movieID string) (string, error) {
	baseURL := fmt.Sprintf("https://api.themoviedb.org/3/movie/%s/external_ids", movieID)
	params := url.Values{}
	params.Add("api_key", ms.apiKey)

	fullURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	resp, err := ms.client.Get(fullURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var externalIDs struct {
		IMDbID string `json:"imdb_id"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&externalIDs); err != nil {
		return "", err
	}

	if externalIDs.IMDbID == "" {
		return "", fmt.Errorf("no IMDb ID found for TMDB movie %s", movieID)
	}

	return externalIDs.IMDbID, nil
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

func (ms *MovieService) searchYTS(query string, year string) ([]models.TorrentResult, error) {
	apiURL := fmt.Sprintf("https://yts.mx/api/v2/list_movies.json?query_term=%s&year=%s",
		url.QueryEscape(query), year)

	resp, err := ms.client.Get(apiURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var ytsResp struct {
		Status string `json:"status"`
		Data   struct {
			Movies []struct {
				Title    string `json:"title"`
				Year     int    `json:"year"`
				Torrents []struct {
					URL     string `json:"url"`
					Hash    string `json:"hash"`
					Quality string `json:"quality"`
					Seeds   int    `json:"seeds"`
					Peers   int    `json:"peers"`
					Size    string `json:"size"`
				} `json:"torrents"`
			} `json:"movies"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&ytsResp); err != nil {
		return nil, err
	}

	var results []models.TorrentResult
	for _, movie := range ytsResp.Data.Movies {
		for _, t := range movie.Torrents {
			if t.Hash == "" {
				log.Printf("Warning: Torrent hash missing for movie %s", movie.Title)
				continue
			}
			magnet := fmt.Sprintf("magnet:?xt=urn:btih:%s&dn=%s", t.Hash, url.QueryEscape(movie.Title))

			results = append(results, models.TorrentResult{
				Name:     fmt.Sprintf("%s (%d) [%s]", movie.Title, movie.Year, t.Quality),
				Magnet:   magnet,
				Size:     t.Size,
				Seeders:  t.Seeds,
				Leechers: t.Peers,
				Quality:  t.Quality,
			})
		}
	}

	return results, nil
}

func (ms *MovieService) searchYTSByIMDb(imdbID string) ([]models.TorrentResult, error) {
	apiURL := fmt.Sprintf("https://yts.mx/api/v2/list_movies.json?query_term=%s", imdbID)

	resp, err := ms.client.Get(apiURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var ytsResp struct {
		Status string `json:"status"`
		Data   struct {
			Movies []struct {
				Title    string `json:"title"`
				Year     int    `json:"year"`
				IMDbCode string `json:"imdb_code"`
				Torrents []struct {
					URL     string `json:"url"`
					Hash    string `json:"hash"`
					Quality string `json:"quality"`
					Seeds   int    `json:"seeds"`
					Peers   int    `json:"peers"`
					Size    string `json:"size"`
				} `json:"torrents"`
			} `json:"movies"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&ytsResp); err != nil {
		return nil, err
	}

	var results []models.TorrentResult
	for _, movie := range ytsResp.Data.Movies {
		// Verify IMDb ID matches
		if movie.IMDbCode != imdbID {
			continue
		}

		for _, t := range movie.Torrents {
			if t.Hash == "" {
				log.Printf("Warning: Torrent hash missing for movie %s", movie.Title)
				continue
			}
			magnet := fmt.Sprintf("magnet:?xt=urn:btih:%s&dn=%s", t.Hash, url.QueryEscape(movie.Title))

			results = append(results, models.TorrentResult{
				Name:     fmt.Sprintf("%s (%d) [%s]", movie.Title, movie.Year, t.Quality),
				Magnet:   magnet,
				Size:     t.Size,
				Seeders:  t.Seeds,
				Leechers: t.Peers,
				Quality:  t.Quality,
			})
		}
	}

	return results, nil
}
