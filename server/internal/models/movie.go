package models

import (
	"sync"
	"time"

	"github.com/anacrolix/torrent"
)

type DownloadedMovie struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	MovieID      int       `gorm:"not null;uniqueIndex:idx_movie_quality" json:"movie_id"`
	Quality      string    `gorm:"size:10;not null;uniqueIndex:idx_movie_quality" json:"quality"`
	FilePath     string    `gorm:"size:500;not null" json:"file_path"`
	MagnetLink   string    `gorm:"type:text" json:"magnet_link"`
	DownloadedAt time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"downloaded_at"`
	LastWatched  time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"last_watched"`
	FileSize     int64     `gorm:"default:0" json:"file_size"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type Subtitle struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	MovieID   int       `gorm:"not null;uniqueIndex:idx_movie_language" json:"movie_id"`
	Language  string    `gorm:"size:10;not null;uniqueIndex:idx_movie_language" json:"language"`
	FilePath  string    `gorm:"size:500;not null" json:"file_path"`
	CreatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Movie struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	ReleaseDate string `json:"release_date"`
	PosterPath  string `json:"poster_path"`
	Overview    string `json:"overview"`
}

type MovieDetails struct {
	ID           int       `json:"id"`
	Title        string    `json:"title"`
	Overview     string    `json:"overview"`
	ReleaseDate  string    `json:"release_date"`
	Runtime      int       `json:"runtime"`
	PosterPath   string    `json:"poster_path"`
	BackdropPath string    `json:"backdrop_path"`
	VoteAverage  float64   `json:"vote_average"`
	IMDbID       string    `json:"imdb_id"`
	IsAvailable  bool      `json:"is_available"`
	StreamURL    string    `json:"stream_url"`
	Cast         []Cast    `json:"cast"`
	Director     []Person  `json:"director"`
	Producer     []Person  `json:"producer"`
	Genres       []Genre   `json:"genres"`
	Comments     []Comment `json:"comments"`
}

type Cast struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Character   string `json:"character"`
	ProfilePath string `json:"profile_path"`
}

type Person struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type Genre struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type Comment struct {
	ID        int       `json:"id"`
	MovieID   int       `json:"movie_id"`
	UserID    int       `json:"user_id"`
	Username  string    `json:"username"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

type TorrentResult struct {
	Name     string `json:"name"`
	Magnet   string `json:"magnet"`
	Size     string `json:"size"`
	Seeders  int    `json:"seeders"`
	Leechers int    `json:"leechers"`
	Quality  string `json:"quality"`
}

type TorrentDownload struct {
	Torrent        *torrent.Torrent `json:"-"`
	VideoFile      *torrent.File    `json:"-"`
	MovieID        int              `json:"movie_id"`
	Quality        string           `json:"quality"`
	Progress       float64          `json:"progress"`
	Status         string           `json:"status"`
	StreamReady    bool             `json:"stream_ready"`
	StreamingReady bool             `json:"streaming_ready"`
	FilePath       string           `json:"file_path"`
	SubtitlePath   string           `json:"subtitle_path"`
	StartedAt      time.Time        `json:"started_at"`
	CompletedAt    *time.Time       `json:"completed_at,omitempty"`
	Mu             sync.RWMutex     `json:"-"`
}

type TranscodeJob struct {
	ID        string    `json:"id"`
	InputPath string    `json:"input_path"`
	Status    string    `json:"status"`
	Progress  float64   `json:"progress"`
	StartedAt time.Time `json:"started_at"`
}
