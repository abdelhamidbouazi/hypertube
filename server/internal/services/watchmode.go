package services

import (
	"server/internal/models"
)

type WatchMode struct {
	apiKey string
}

func NewWatchMode(apiKey string) *WatchMode {
	return &WatchMode{
		apiKey: apiKey,
	}
}

func (w WatchMode) DiscoverMovies(p MovieDiscoverParams) ([]models.Movie, error) {
	return nil, nil
}
