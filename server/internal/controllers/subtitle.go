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
		language = "en"
	}

	var subtitle models.Subtitle
	err := c.db.Where("movie_id = ? AND language = ?", movieID, language).First(&subtitle).Error
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Subtitles not found")
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
