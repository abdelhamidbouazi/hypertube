package handlers

import (
	"net/http"
	"server/internal/models"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

type CommentHandler struct {
	db *gorm.DB
}

func NewCommentHandler(db *gorm.DB) *CommentHandler {
	return &CommentHandler{
		db: db,
	}
}

func (h *CommentHandler) AddComment(c echo.Context) error {
	var requestComment struct {
		MovieID  int    `json:"movie_id"`
		Username string `json:"username"`
		Content  string `json:"content"`
	}

	if err := c.Bind(&requestComment); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request")
	}

	comment := models.Comment{
		MovieID:  requestComment.MovieID,
		Username: requestComment.Username,
		Content:  requestComment.Content,
	}

	if err := h.db.Create(&comment).Error; err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to add comment")
	}

	return c.JSON(http.StatusOK, comment)
}
