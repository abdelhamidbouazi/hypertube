package handlers

import (
	"database/sql"
	"net/http"
	"github.com/labstack/echo/v4"
)

type CommentHandler struct {
	db *sql.DB
}

func NewCommentHandler(db *sql.DB) *CommentHandler {
	return &CommentHandler{
		db: db,
	}
}


func (h *CommentHandler) AddComment(c echo.Context) error {
	var comment struct {
		MovieID  int    `json:"movie_id"`
		Username string `json:"username"`
		Content  string `json:"content"`
	}

	if err := c.Bind(&comment); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request")
	}

	result, err := h.db.Exec("INSERT INTO comments (movie_id, username, content) VALUES (?, ?, ?)",
		comment.MovieID, comment.Username, comment.Content)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to add comment")
	}

	id, _ := result.LastInsertId()

	return c.JSON(http.StatusOK, map[string]int64{"id": id})
}
