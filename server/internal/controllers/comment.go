package controllers

import (
	"net/http"
	"server/internal/models"
	"strconv"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

type CommentController struct {
	db *gorm.DB
}

func NewCommentController(db *gorm.DB) *CommentController {
	return &CommentController{
		db: db,
	}
}

func (c *CommentController) AddComment(ctx echo.Context) error {
	var requestComment struct {
		MovieID  int    `json:"movie_id"`
		Username string `json:"username"`
		Content  string `json:"content"`
	}

	if err := ctx.Bind(&requestComment); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request")
	}

	user, exists := ctx.Get("model").(models.User)
	if !exists {
		return echo.NewHTTPError(http.StatusUnauthorized, "User not authenticated")
	}

	comment := models.Comment{
		MovieID:  requestComment.MovieID,
		UserID:   int(user.ID),
		Username: requestComment.Username,
		Content:  requestComment.Content,
	}

	if err := c.db.Create(&comment).Error; err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to add comment")
	}

	return ctx.JSON(http.StatusOK, comment)
}

func (c *CommentController) GetComments(ctx echo.Context) error {
	var comments []models.Comment
	if err := c.db.Order("created_at DESC").Find(&comments).Error; err != nil {
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

	return ctx.JSON(http.StatusOK, response)
}

func (c *CommentController) GetCommentByID(ctx echo.Context) error {
	id := ctx.Param("id")
	commentID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid comment ID")
	}

	var comment models.Comment
	if err := c.db.First(&comment, commentID).Error; err != nil {
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

	return ctx.JSON(http.StatusOK, response)
}

func (c *CommentController) UpdateComment(ctx echo.Context) error {
	id := ctx.Param("id")
	commentID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid comment ID")
	}

	var requestData struct {
		Content  string `json:"content"`
		Username string `json:"username"`
	}

	if err := ctx.Bind(&requestData); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request data")
	}

	var comment models.Comment
	if err := c.db.First(&comment, commentID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return echo.NewHTTPError(http.StatusNotFound, "Comment not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve comment")
	}

	if comment.Username != requestData.Username {
		return echo.NewHTTPError(http.StatusForbidden, "You can only update your own comments")
	}

	if err := c.db.Model(&comment).Update("content", requestData.Content).Error; err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update comment")
	}

	return ctx.JSON(http.StatusOK, map[string]string{"message": "Comment updated successfully"})
}

func (c *CommentController) DeleteComment(ctx echo.Context) error {
	id := ctx.Param("id")
	commentID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid comment ID")
	}

	user, exists := ctx.Get("model").(models.User)
	if !exists {
		return echo.NewHTTPError(http.StatusUnauthorized, "User not authenticated")
	}

	var comment models.Comment
	if err := c.db.First(&comment, commentID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return echo.NewHTTPError(http.StatusNotFound, "Comment not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve comment")
	}

	if comment.UserID != int(user.ID) {
		return echo.NewHTTPError(http.StatusForbidden, "You can only delete your own comments")
	}

	if err := c.db.Delete(&comment).Error; err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to delete comment")
	}

	return ctx.JSON(http.StatusOK, map[string]string{"message": "Comment deleted successfully"})
}
