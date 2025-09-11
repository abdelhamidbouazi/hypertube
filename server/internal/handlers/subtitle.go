package handlers

import (
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/labstack/echo/v4"
)

type SubtitleHandler struct {
	db *sql.DB
}

func NewSubtitleHandler(db *sql.DB) *SubtitleHandler {
	return &SubtitleHandler{
		db: db,
	}
}

func (h *SubtitleHandler) GetSubtitles(c echo.Context) error {
	movieID := c.QueryParam("movie_id")
	language := c.QueryParam("lang")
	if language == "" {
		language = "en"
	}

	var filePath string
	err := h.db.QueryRow("SELECT file_path FROM subtitles WHERE movie_id = ? AND language = ?", movieID, language).Scan(&filePath)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Subtitles not found")
	}

	ext := strings.ToLower(filepath.Ext(filePath))
	if ext == ".srt" {
		c.Response().Header().Set("Content-Type", "text/vtt")
		convertSRTtoVTT(filePath, c.Response().Writer)
		return nil
	}

	return c.File(filePath)
}

func convertSRTtoVTT(srtPath string, w io.Writer) {
	file, err := os.Open(srtPath)
	if err != nil {
		return
	}
	defer file.Close()

	fmt.Fprintln(w, "WEBVTT\n")

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
