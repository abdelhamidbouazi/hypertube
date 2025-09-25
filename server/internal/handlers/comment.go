package handlers

import (
	"net/http"
	"server/internal/models"
	"strconv"

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

	user, exists := c.Get("model").(models.User)
	if !exists {
		return echo.NewHTTPError(http.StatusUnauthorized, "User not authenticated")
	}

	comment := models.Comment{
		MovieID:  requestComment.MovieID,
		UserID:   int(user.ID),
		Username: requestComment.Username,
		Content:  requestComment.Content,
	}

	if err := h.db.Create(&comment).Error; err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to add comment")
	}

	return c.JSON(http.StatusOK, comment)
}

func (h *CommentHandler) GetComments(c echo.Context) error {
	var comments []models.Comment
	if err := h.db.Order("created_at DESC").Find(&comments).Error; err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve comments")
	}

	type CommentResponse struct {
		ID       uint   `json:"id"`
		Username string `json:"username"`
		Date     string `json:"date"`
		Content  string `json:"content"`
	}

	var response []CommentResponse
	for _, comment := range comments {
		response = append(response, CommentResponse{
			ID:       comment.ID,
			Username: comment.Username,
			Date:     comment.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			Content:  comment.Content,
		})
	}

	return c.JSON(http.StatusOK, response)
}

func (h *CommentHandler) GetCommentByID(c echo.Context) error {
	id := c.Param("id")
	commentID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid comment ID")
	}

	var comment models.Comment
	if err := h.db.First(&comment, commentID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return echo.NewHTTPError(http.StatusNotFound, "Comment not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve comment")
	}

	type CommentResponse struct {
		ID       uint   `json:"id"`
		Username string `json:"username"`
		Date     string `json:"date"`
		Content  string `json:"content"`
	}

	response := CommentResponse{
		ID:       comment.ID,
		Username: comment.Username,
		Date:     comment.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		Content:  comment.Content,
	}

	return c.JSON(http.StatusOK, response)
}

func (h *CommentHandler) UpdateComment(c echo.Context) error {
	id := c.Param("id")
	commentID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid comment ID")
	}

	var requestData struct {
		Content  string `json:"content"`
		Username string `json:"username"`
	}

	if err := c.Bind(&requestData); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request data")
	}

	var comment models.Comment
	if err := h.db.First(&comment, commentID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return echo.NewHTTPError(http.StatusNotFound, "Comment not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve comment")
	}

	if comment.Username != requestData.Username {
		return echo.NewHTTPError(http.StatusForbidden, "You can only update your own comments")
	}

	if err := h.db.Model(&comment).Update("content", requestData.Content).Error; err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update comment")
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Comment updated successfully"})
}

func (h *CommentHandler) DeleteComment(c echo.Context) error {
	id := c.Param("id")
	commentID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid comment ID")
	}

	user, exists := c.Get("model").(models.User)
	if !exists {
		return echo.NewHTTPError(http.StatusUnauthorized, "User not authenticated")
	}

	var comment models.Comment
	if err := h.db.First(&comment, commentID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return echo.NewHTTPError(http.StatusNotFound, "Comment not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve comment")
	}

	if comment.UserID != int(user.ID) {
		return echo.NewHTTPError(http.StatusForbidden, "You can only delete your own comments")
	}

	if err := h.db.Delete(&comment).Error; err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to delete comment")
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Comment deleted successfully"})
}
