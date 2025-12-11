package controllers

import "server/internal/models"

// AddCommentRequest represents the payload to add a comment
type AddCommentRequest struct {
	MovieID  int    `json:"movie_id"`
	Username string `json:"username"`
	Content  string `json:"content"`
}

type MovieDetailsDoc struct {
	ID           int               `json:"id"`
	Title        string            `json:"title"`
	Overview     string            `json:"overview"`
	ReleaseDate  string            `json:"release_date"`
	Runtime      int               `json:"runtime"`
	PosterPath   string            `json:"poster_path"`
	BackdropPath string            `json:"backdrop_path"`
	VoteAverage  float64           `json:"vote_average"`
	IMDbID       string            `json:"imdb_id"`
	Language     string            `json:"original_language,omitempty"`
	IsAvailable  bool              `json:"is_available"`
	StreamURL    string            `json:"stream_url"`
	Cast         []models.Cast     `json:"cast"`
	Director     []models.Person   `json:"director"`
	Producer     []models.Person   `json:"producer"`
	Genres       []models.Genre    `json:"genres"`
	Comments     []CommentResponse `json:"comments"`
	IsWatched    bool              `json:"isWatched"`
}

// CommentResponse represents a comment in responses
type CommentResponse struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Avatar   string `json:"avatar"`
	Date     string `json:"date"`
	Content  string `json:"content"`
}
