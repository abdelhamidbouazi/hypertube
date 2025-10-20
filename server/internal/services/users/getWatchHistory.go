package users

import (
	"time"
	"server/internal/models"
	"server/internal/services"
)

// WatchHistoryItem represents a single watch history entry for Swagger
type WatchHistoryItem struct {
	ID           uint      `json:"id"`
	UserID       uint      `json:"user_id"`
	MovieID      int       `json:"movie_id"`
	MovieTitle   string    `json:"movie_title"`
	PosterPath   string    `json:"poster_path"`
	WatchedAt    time.Time `json:"watched_at"`
	Duration     int       `json:"duration"`
	LastPosition int       `json:"last_position"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type WatchHistoryResponse struct {
	History []models.WatchHistory `json:"history" swaggertype:"array,object"`
	Total   int64                 `json:"total"`
}

func GetUserWatchHistory(userID uint, page int, limit int) (WatchHistoryResponse, error) {
	db := services.PostgresDB()
	var response WatchHistoryResponse

	// Count total records
	err := db.Model(&models.WatchHistory{}).Where("user_id = ?", userID).Count(&response.Total).Error
	if err != nil {
		return response, err
	}

	// Get paginated watch history ordered by most recent first
	offset := (page - 1) * limit
	err = db.Where("user_id = ?", userID).
		Order("watched_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&response.History).Error
	
	if err != nil {
		return response, err
	}

	return response, nil
}
