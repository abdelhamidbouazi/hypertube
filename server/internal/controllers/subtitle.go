package controllers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"server/internal/models"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

type SubtitleController struct {
	db *gorm.DB
}

func NewSubtitleController(db *gorm.DB) *SubtitleController {
	return &SubtitleController{
		db: db,
	}
}

func (c *SubtitleController) GetSubtitles(ctx echo.Context) error {
	movieID := ctx.QueryParam("movie_id")
	language := ctx.QueryParam("lang")
	if language == "" {
		user, exists := ctx.Get("model").(models.User)
		if exists && user.PreferredLanguage != "" {
			language = user.PreferredLanguage
		} else {
			language = "en" // Default to English
		}
	}

	var subtitle models.Subtitle
	err := c.db.Where("movie_id = ? AND language = ?", movieID, language).First(&subtitle).Error
	if err != nil {
		if language != "en" {
			err = c.db.Where("movie_id = ? AND language = ?", movieID, "en").First(&subtitle).Error
			if err != nil {
				return echo.NewHTTPError(http.StatusNotFound, "Subtitles not found")
			}
		} else {
			return echo.NewHTTPError(http.StatusNotFound, "Subtitles not found")
		}
	}

	filePath := subtitle.FilePath

	ext := strings.ToLower(filepath.Ext(filePath))
	if ext == ".srt" {
		ctx.Response().Header().Set("Content-Type", "text/vtt")
		convertSRTtoVTT(filePath, ctx.Response().Writer)
		return nil
	}

	return ctx.File(filePath)
}

func (c *SubtitleController) GetAvailableLanguages(ctx echo.Context) error {
	movieID := ctx.QueryParam("movie_id")
	if movieID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "movie_id is required")
	}

	var subtitles []models.Subtitle
	err := c.db.Where("movie_id = ?", movieID).Find(&subtitles).Error
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve available languages")
	}

	languages := make([]string, 0, len(subtitles))
	for _, subtitle := range subtitles {
		languages = append(languages, subtitle.Language)
	}

	return ctx.JSON(http.StatusOK, map[string]interface{}{
		"movie_id":            movieID,
		"available_languages": languages,
		"default_language":    "en",
	})
}

func (c *SubtitleController) GetSubtitleRecommendations(ctx echo.Context) error {
	movieID := ctx.QueryParam("movie_id")
	if movieID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "movie_id is required")
	}

	user, exists := ctx.Get("model").(models.User)
	if !exists {
		return echo.NewHTTPError(http.StatusUnauthorized, "User not authenticated")
	}
	var subtitles []models.Subtitle
	err := c.db.Where("movie_id = ?", movieID).Find(&subtitles).Error
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve available languages")
	}

	availableLanguages := make([]string, 0, len(subtitles))
	for _, subtitle := range subtitles {
		availableLanguages = append(availableLanguages, subtitle.Language)
	}

	userLang := user.PreferredLanguage
	if userLang == "" {
		userLang = "en"
	}
	recommendSubtitles := true
	recommendedLanguage := userLang

	userLangAvailable := false
	englishAvailable := false
	for _, lang := range availableLanguages {
		if lang == userLang {
			userLangAvailable = true
		}
		if lang == "en" {
			englishAvailable = true
		}
	}

	if !userLangAvailable {
		if englishAvailable {
			recommendedLanguage = "en"
		} else if len(availableLanguages) > 0 {
			recommendedLanguage = availableLanguages[0]
		} else {
			recommendSubtitles = false
		}
	}

	return ctx.JSON(http.StatusOK, map[string]interface{}{
		"movie_id":                movieID,
		"user_preferred_language": userLang,
		"available_languages":     availableLanguages,
		"recommend_subtitles":     recommendSubtitles,
		"recommended_language":    recommendedLanguage,
	})
}

func convertSRTtoVTT(srtPath string, w io.Writer) {
	file, err := os.Open(srtPath)
	if err != nil {
		return
	}
	defer file.Close()

	fmt.Fprintln(w, "WEBVTT")

	buf := make([]byte, 4096)
	for {
		n, err := file.Read(buf)
		if n > 0 {
			text := string(buf[:n])
			text = strings.ReplaceAll(text, ",", ".")
			w.Write([]byte(text))
		}
		if err != nil {
			break
		}
	}
}
