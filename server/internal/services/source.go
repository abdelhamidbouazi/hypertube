package services

import (
	"server/internal/models"
)

type Source interface {
	DiscoverMovies(p MovieDiscoverParams) ([]models.Movie, []int, error)
	getGenreIDs(genres []string) (string, error)
	GetIMDbID(movieID int) (*models.MovieDetails, error)
	GetMovieDetails(movieID string) (*models.MovieDetails, error)
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
