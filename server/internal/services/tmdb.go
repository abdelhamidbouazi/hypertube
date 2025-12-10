package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"server/internal/models"
	"strconv"
	"sync"
	"time"
)

type TMDB struct {
	apiKey         string
	omdbKey        string
	genreCacheTime time.Time
	genreCache     sync.Map // map[string]int
	client         *http.Client
}

func NewTMDB(apiKey, omdbKey string, genreCacheTime time.Time, client *http.Client) *TMDB {
	return &TMDB{
		apiKey:  apiKey,
		client:  client,
		omdbKey: omdbKey,
	}
}

func (t *TMDB) GetMovieDetails(movieID string) (*models.MovieDetails, error) {
	baseURL := fmt.Sprintf("https://api.themoviedb.org/3/movie/%s", movieID)
	params := url.Values{}
	params.Add("api_key", t.apiKey)
	params.Add("append_to_response", "credits,external_ids")

	fullURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	resp, err := t.client.Get(fullURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var tmdbResp map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&tmdbResp); err != nil {
		return nil, err
	}

	details := parseTMDBResponse(tmdbResp)

	if imdbID, ok := tmdbResp["imdb_id"].(string); ok && imdbID != "" {
		details.IMDbID = imdbID
		if omdbData := t.getOMDBData(imdbID); omdbData != nil {
			if rating, ok := omdbData["imdbRating"].(string); ok {
				if r, err := strconv.ParseFloat(rating, 64); err == nil {
					details.VoteAverage = r
				}
			}
		}
	}

	return details, nil
}

func (t *TMDB) getOMDBData(imdbID string) map[string]interface{} {
	url := fmt.Sprintf("http://www.omdbapi.com/?i=%s&apikey=%s", imdbID, t.omdbKey)

	resp, err := t.client.Get(url)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	var data map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&data)
	return data
}

func parseTMDBResponse(data map[string]interface{}) *models.MovieDetails {
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

func (t *TMDB) GetIMDbID(movieID int) (*models.MovieDetails, error) {
	Logger.Info(fmt.Sprintf("Fetching IMDb ID for TMDB movie %d", movieID))
	movieIDStr := strconv.Itoa(movieID)

	details, err := t.GetMovieDetails(movieIDStr)
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
	params.Add("api_key", t.apiKey)

	fullURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	resp, err := t.client.Get(fullURL)
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

func (t *TMDB) getGenreIDs(genres []string) (string, error) {
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
	if time.Since(t.genreCacheTime) > 24*time.Hour {
		if err := t.refreshGenreCache(); err != nil {
			return "", err
		}
	}

	ids := []string{}
	for _, name := range genres {
		key := normalizeGenreName(name)
		if id, ok := t.genreCache.Load(key); ok {
			ids = append(ids, strconv.Itoa(id.(int)))
		} else {
			// log.Printf("unknown genre name: %s", name)
		}
	}
	return joinCSV(ids), nil
}

func (t *TMDB) refreshGenreCache() error {
	baseURL := "https://api.themoviedb.org/3/genre/movie/list"
	params := url.Values{}
	params.Add("api_key", t.apiKey)
	fullURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	resp, err := t.client.Get(fullURL)
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
		t.genreCache.Store(normalizeGenreName(g.Name), g.ID)
	}
	t.genreCacheTime = time.Now()
	return nil
}

func (t *TMDB) DiscoverMovies(p MovieDiscoverParams) ([]models.Movie, []int, error) {
	if p.Page < 1 {
		p.Page = 1
	}

	baseURL := "https://api.themoviedb.org/3/discover/movie"
	params := url.Values{}
	params.Add("api_key", t.apiKey)
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
		ids, err := t.getGenreIDs(p.Genres)
		if err == nil && len(ids) > 0 {
			params.Add("with_genres", ids)
		}
	}

	fullURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	resp, err := t.client.Get(fullURL)
	if err != nil {
		return nil, nil, err
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
		return nil, nil, err
	}

	var movies []models.Movie
	var movieIDs []int

	for _, result := range tmdbResp.Results {
		movieIDs = append(movieIDs, result.ID)
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

	return movies, movieIDs, nil
}
