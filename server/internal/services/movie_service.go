package services

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"server/internal/models"
)

type MovieService struct {
	apiKey         string
	omdbKey        string
	client         *http.Client
	torrentSources []string
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
	}
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
			ID          int    `json:"id"`
			Title       string `json:"title"`
			ReleaseDate string `json:"release_date"`
			PosterPath  string `json:"poster_path"`
			Overview    string `json:"overview"`
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
		})
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
