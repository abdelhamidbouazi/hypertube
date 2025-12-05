package users

import (
	"server/internal/models"
	"server/internal/services"
)

type UserStats struct {
	TotalWatched    int64 `json:"total_watched"`
	TotalComments   int64 `json:"total_comments"`
	TotalWatchTime  int   `json:"total_watch_time"`
	FavoriteGenres  []GenreStats `json:"favorite_genres"`
}

type GenreStats struct {
	GenreID   int    `json:"genre_id"`
	GenreName string `json:"genre_name"`
	Count     int    `json:"count"`
}

func GetUserStats(userID uint) (UserStats, error) {
	db := services.PostgresDB()
	var stats UserStats

	// Count total watched movies
	err := db.Model(&models.WatchHistory{}).Where("user_id = ?", userID).Count(&stats.TotalWatched).Error
	if err != nil {
		return stats, err
	}

	// Count total comments
	err = db.Model(&models.Comment{}).Where("user_id = ?", userID).Count(&stats.TotalComments).Error
	if err != nil {
		return stats, err
	}

	// Calculate total watch time
	var totalDuration int
	err = db.Model(&models.WatchHistory{}).
		Where("user_id = ?", userID).
		Select("COALESCE(SUM(duration), 0)").
		Scan(&totalDuration).Error
	if err != nil {
		return stats, err
	}
	stats.TotalWatchTime = totalDuration
	
	stats.FavoriteGenres = []GenreStats{}

	return stats, nil
}
