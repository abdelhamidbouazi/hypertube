package models

import (
	"sync"
	"time"

	"github.com/anacrolix/torrent"
	"gorm.io/gorm"
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
	Transcoded   bool      `gorm:"default:false" json:"transcoded"`
	LastSegment  string    `gorm:"size:50" json:"last_segment"`
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
	ID          int     `json:"id"`
	Title       string  `json:"title"`
	ReleaseDate string  `json:"release_date"`
	PosterPath  string  `json:"poster_path"`
	Overview    string  `json:"overview"`
	Language    string  `json:"original_language,omitempty"`
	VoteAverage float64 `json:"vote_average"`
	GenreIDs    []int   `json:"genre_ids,omitempty"`
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
	Language     string    `json:"original_language,omitempty"`
	IsAvailable  bool      `json:"is_available"`
	IsWatched    bool      `json:"is_watched"`
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
	gorm.Model
	MovieID  int    `json:"movie_id"`
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
	Content  string `json:"content"`
}

type WatchHistory struct {
	gorm.Model
	UserID        uint      `gorm:"not null;index" json:"user_id"`
	MovieID       int       `gorm:"not null" json:"movie_id"`
	MovieTitle    string    `gorm:"size:500" json:"movie_title"`
	PosterPath    string    `gorm:"size:500" json:"poster_path"`
	WatchedAt     time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"watched_at"`
	Duration      int       `json:"duration"`
	LastPosition  int       `json:"last_position"`
	LastSegment   string    `gorm:"size:50" json:"last_segment"`
	WatchProgress float64   `gorm:"type:decimal(5,2);default:0" json:"watch_progress"`
	WatchCount    int       `gorm:"default:0" json:"watch_count"`
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
	RootDir        string
	MovieID        int          `json:"movie_id"`
	Quality        string       `json:"quality"`
	Progress       float64      `json:"progress"`
	Status         string       `json:"status"`
	StreamReady    bool         `json:"stream_ready"`
	StreamingReady bool         `json:"streaming_ready"`
	FilePath       string       `json:"file_path"`
	SubtitlePath   string       `json:"subtitle_path"`
	StartedAt      time.Time    `json:"started_at"`
	CompletedAt    *time.Time   `json:"completed_at,omitempty"`
	Mu             sync.RWMutex `json:"-"`
}

type TranscodeJob struct {
	ID        string    `json:"id"`
	InputPath string    `json:"input_path"`
	Status    string    `json:"status"`
	Progress  float64   `json:"progress"`
	StartedAt time.Time `json:"started_at"`
}
