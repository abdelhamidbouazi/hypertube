package controllers

import (
	"net/http"
	"server/internal/models"
	"server/internal/services/users"
	"strconv"

	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

type UpdateUserRequest struct {
	FirstName         string `json:"firstname" validate:"omitempty,min=4"`
	LastName          string `json:"lastname" validate:"omitempty,min=4"`
	Email             string `json:"email" validate:"omitempty,email"`
	Password          string `json:"password" validate:"omitempty,min=8"`
	PreferredLanguage string `json:"preferred_language" validate:"omitempty,len=2"`
}

type UpdateUserRequest struct {
	FirstName         string `validate:"min=4" example:"Alan"`
	LastName          string `validate:"min=4" example:"Turing"`
	Email             string `validate:"email" example:"example@email.com"`
	Username          string `validate:"" example:"fturing"`
	PreferredLanguage string `validate:"max=10,omitempty" default:"en" json:"preferred_language" example:"en"`
}


// Update user info godoc
//
//	@Summary		Update user info
//	@Description	update current user info
//	@Tags			users
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			UpdateUserRequest	body		UpdateUserRequest	true	"user info to update"
//	@Success		200	{object}	models.User
//	@Failure		400	{object}	utils.HTTPError
//	@Failure		401	{object}	utils.HTTPErrorUnauthorized
//	@Router			/users/me [patch]
func UpdateUser(c echo.Context) error {
	var req UpdateUserRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	// Get current user from context
	userModel := c.Get("model").(models.User)

	// Update fields if provided
	if req.FirstName != "" {
		userModel.FirstName = req.FirstName
	}
	if req.LastName != "" {
		userModel.LastName = req.LastName
	}
	if req.Email != "" {
		userModel.Email = req.Email
	}
	if req.PreferredLanguage != "" {
		userModel.PreferredLanguage = req.PreferredLanguage
	}
	if req.Password != "" {
		hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), 8)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to hash password")
		}
		userModel.Password = string(hashed)
	}

	// Save changes
	if err := users.UpdateUser(userModel); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update user")
	}

	return c.JSON(http.StatusOK, userModel)
}

// Get users info godoc
//
//	@Summary		List of users
//	@Description	get users info
//	@Tags			users
//	@Produce		json
//	@Success		200	{array}		models.User
//	@Failure		401	{object}	utils.HTTPErrorUnauthorized
//	@Security		JWT
//	@Router			/users [get]
func GetUsers(c echo.Context) error {
	users, err := users.GetUsers()
	if err != nil {
		response := echo.Map{
			"message": err.Error(),
		}
		return c.JSON(http.StatusOK, response)
	}

	response := echo.Map{
		"data": users,
	}

	return c.JSON(http.StatusOK, response)
}

type UpdateUserRequest struct {
	FirstName         string `validate:"min=4" example:"Alan"`
	LastName          string `validate:"min=4" example:"Turing"`
	Email             string `validate:"email" example:"example@email.com"`
	Username          string `validate:"" example:"fturing"`
	PreferredLanguage string `validate:"max=10,omitempty" default:"en" json:"preferred_language" example:"en"`
}

// Update users info godoc
//
//	@Summary		Update a user
//	@Description	update a user
//	@Tags			users
//	@Produce		json
//	@Param		UpdateUserRequest body UpdateUserRequest true "update user info"
//	@Success		200	{array}		models.User
//	@Failure 400 {object} utils.HTTPError
//	@Failure		401	{object}	utils.HTTPErrorUnauthorized
//	@Security		JWT
//	@Router			/users [patch]
func UpdateUser(c echo.Context) error {
	var user UpdateUserRequest

	err := c.Bind(&user)
	if err != nil {
		return echo.ErrBadRequest
	}

	foundUser := c.Get("model").(models.User)

	if user.Email != "" {
		foundUser.Email = user.Email
	}
	if user.Username != "" {
		foundUser.Username = user.Username
	}
	if user.FirstName != "" {
		foundUser.FirstName = user.FirstName
	}
	if user.LastName != "" {
		foundUser.LastName = user.LastName
	}
	if user.PreferredLanguage != "" {
		foundUser.PreferredLanguage = user.PreferredLanguage
	}

	users.UpdateUser(foundUser)
	return c.JSON(http.StatusOK, foundUser)
}

// Get user info godoc
//
//	@Summary		User info
//	@Description	get current user info
//	@Tags			users
//	@Security		JWT
//	@Produce		json
//	@Success		200	{object}	models.User
//	@Failure		401	{object}	utils.HTTPErrorUnauthorized
//	@Router			/users/me [get]
func GetMe(c echo.Context) error {
	return c.JSON(http.StatusOK, c.Get("model"))
}

// Get user stats godoc
//
//	@Summary		User statistics
//	@Description	get user statistics including watch count, comments, and watch time
//	@Tags			users
//	@Security		JWT
//	@Produce		json
//	@Success		200	{object}	users.UserStats
//	@Failure		401	{object}	utils.HTTPErrorUnauthorized
//	@Router			/users/stats [get]
func GetUserStats(c echo.Context) error {
	userModel := c.Get("model").(models.User)

	stats, err := users.GetUserStats(userModel.ID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"error":   "Failed to retrieve user statistics",
			"message": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, stats)
}

// WatchHistoryResponseDoc is for Swagger documentation only
type WatchHistoryResponseDoc struct {
	History []WatchHistoryItemDoc `json:"history"`
	Total   int64                 `json:"total"`
}

// WatchHistoryItemDoc is for Swagger documentation only
type WatchHistoryItemDoc struct {
	ID           uint   `json:"id"`
	UserID       uint   `json:"user_id"`
	MovieID      int    `json:"movie_id"`
	MovieTitle   string `json:"movie_title"`
	PosterPath   string `json:"poster_path"`
	WatchedAt    string `json:"watched_at"`
	Duration     int    `json:"duration"`
	LastPosition int    `json:"last_position"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}

// Get user watch history godoc
//
//	@Summary		User watch history
//	@Description	get user's watch history with pagination
//	@Tags			users
//	@Security		JWT
//	@Produce		json
//	@Param			page	query		int	false	"Page number (default: 1)"
//	@Param			limit	query		int	false	"Items per page (default: 20)"
//	@Success		200	{object}	WatchHistoryResponseDoc
//	@Failure		401	{object}	utils.HTTPErrorUnauthorized
//	@Router			/users/watch-history [get]
func GetUserWatchHistory(c echo.Context) error {
	userModel := c.Get("model").(models.User)

	// Parse pagination parameters
	page := 1
	limit := 20

	if pageStr := c.QueryParam("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if limitStr := c.QueryParam("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	history, err := users.GetUserWatchHistory(userModel.ID, page, limit)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"error":   "Failed to retrieve watch history",
			"message": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, history)
}
