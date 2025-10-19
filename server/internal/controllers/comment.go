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

// AddCommentRequest represents the payload to add a comment
type AddCommentRequest struct {
	MovieID  int    `json:"movie_id"`
	Username string `json:"username"`
	Content  string `json:"content"`
}

// CommentResponse represents a comment in responses
type CommentResponse struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Date     string `json:"date"`
	Content  string `json:"content"`
}

// AddComment godoc
//
//	@Summary      Add comment
//	@Description  Add a new comment for a movie
//	@Tags         comments
//	@Accept       json
//	@Produce      json
//	@Security     JWT
//	@Param        body  body      AddCommentRequest  true  "Comment body"
	//	@Success      200   {object}  controllers.CommentResponse
//	@Failure      400   {object}  utils.HTTPError
//	@Failure      401   {object}  utils.HTTPErrorUnauthorized
//	@Failure      500   {object}  utils.HTTPError
//	@Router       /comments/add [post]
func (c *CommentController) AddComment(ctx echo.Context) error {
	var requestComment AddCommentRequest

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

	// GetComments godoc
	//
	//  @Summary      List comments
	//  @Description  Get all comments (latest first)
	//  @Tags         comments
	//  @Accept       json
	//  @Produce      json
	//  @Security     JWT
	//  @Success      200  {array}   CommentResponse
	//  @Failure      500  {object}  utils.HTTPError
	//  @Router       /comments [get]
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

// GetCommentByID godoc
//
//	@Summary      Get comment
//	@Description  Get a single comment by ID
//	@Tags         comments
//	@Accept       json
//	@Produce      json
//	@Security     JWT
//	@Param        id   path      string  true  "Comment ID"
//	@Success      200  {object}  CommentResponse
//	@Failure      400  {object}  utils.HTTPError
//	@Failure      401  {object}  utils.HTTPErrorUnauthorized
//	@Failure      404  {object}  utils.HTTPError
//	@Failure      500  {object}  utils.HTTPError
//	@Router       /comments/{id} [get]
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

	response := CommentResponse{
		ID:       comment.ID,
		Username: comment.Username,
		Date:     comment.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		Content:  comment.Content,
	}

	return ctx.JSON(http.StatusOK, response)
}

// UpdateCommentRequest represents the payload to update a comment
type UpdateCommentRequest struct {
	Content  string `json:"content"`
	Username string `json:"username"`
}

// UpdateComment godoc
//
//	@Summary      Update comment
//	@Description  Update an existing comment (owner only)
//	@Tags         comments
//	@Accept       json
//	@Produce      json
//	@Security     JWT
//	@Param        id    path      string                 true  "Comment ID"
//	@Param        body  body      UpdateCommentRequest   true  "Update body"
//	@Success      200   {object}  map[string]string
//	@Failure      400   {object}  utils.HTTPError
//	@Failure      401   {object}  utils.HTTPErrorUnauthorized
//	@Failure      403   {object}  utils.HTTPError
//	@Failure      404   {object}  utils.HTTPError
//	@Failure      500   {object}  utils.HTTPError
//	@Router       /comments/{id} [patch]
func (c *CommentController) UpdateComment(ctx echo.Context) error {
	id := ctx.Param("id")
	commentID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid comment ID")
	}

	var requestData UpdateCommentRequest

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

// DeleteComment godoc
//
//	@Summary      Delete comment
//	@Description  Delete an existing comment (owner only)
//	@Tags         comments
//	@Accept       json
//	@Produce      json
//	@Security     JWT
//	@Param        id   path      string  true  "Comment ID"
//	@Success      200  {object}  map[string]string
//	@Failure      400  {object}  utils.HTTPError
//	@Failure      401  {object}  utils.HTTPErrorUnauthorized
//	@Failure      403  {object}  utils.HTTPError
//	@Failure      404  {object}  utils.HTTPError
//	@Failure      500  {object}  utils.HTTPError
//	@Router       /comments/{id} [delete]
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
