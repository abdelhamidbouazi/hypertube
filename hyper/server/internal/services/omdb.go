package services

import (
	"net/http"
	"server/internal/models"
	"sync"
	"time"
)

type OMDB struct {
	apiKey         string
	genreCacheTime time.Time
	genreCache     sync.Map // map[string]int
	client         *http.Client
}

func NewOMDB(apiKey string, genreCacheTime time.Time, client *http.Client) *OMDB {
	return &OMDB{
		apiKey:         apiKey,
		genreCacheTime: genreCacheTime,
		client:         client,
	}
}

func (w *OMDB) DiscoverMovies(p MovieDiscoverParams) ([]models.Movie, error) {
	var movies []models.Movie
	return movies, nil
}

func (w *OMDB) GetIMDbID(movieID int) (*models.MovieDetails, error) {
	return nil, nil
}

func (w *OMDB) GetMovieDetails(movieID string) (*models.MovieDetails, error) {
	return nil, nil
}

func (w *OMDB) getGenreIDs(genres []string) (string, error) {
	var ids string
	var err error

	return ids, err
}
