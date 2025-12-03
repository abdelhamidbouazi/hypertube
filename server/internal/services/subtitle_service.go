package services

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type SubtitleService struct {
	apiKey     string
	httpClient *http.Client
	baseURL    string
}

type SubdlSearchResult struct {
	Status    bool                `json:"status"`
	Subtitles []SubdlSubtitleInfo `json:"subtitles"`
}

type SubdlSubtitleInfo struct {
	Name         string      `json:"name"`
	ReleaseName  string      `json:"release_name"`
	Lang         string      `json:"lang"`
	Author       string      `json:"author"`
	URL          string      `json:"url"`
	SubtitlePage string      `json:"subtitlePage"`
	Season       interface{} `json:"season"`
	Episode      interface{} `json:"episode"`
	Language     string      `json:"language"`
	HI           interface{} `json:"hi"`
	DownloadURL  string      `json:"download_url"`
}

type SubdlDownloadResponse struct {
	Status bool   `json:"status"`
	URL    string `json:"url"`
}

func NewSubtitleService(apiKey string) (*SubtitleService, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("subdl API key is required")
	}

	return &SubtitleService{
		apiKey:  apiKey,
		baseURL: "https://api.subdl.com/api/v1",
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

func (s *SubtitleService) CreateSubtitlesDirectory(movieID int, baseDir string) (string, error) {
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create base directory %s: %w", baseDir, err)
	}

	dirPath := filepath.Join(baseDir, fmt.Sprintf("%d", movieID))
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return "", fmt.Errorf("failed to create movie directory %s: %w", dirPath, err)
	}
	return dirPath, nil
}

func (s *SubtitleService) DownloadSubtitles(tmdbID int, outputDir string) int {
	languages := []string{"en", "fr", "es", "de", "it", "pt", "ar", "ru"}

	downloadedCount := 0

	if s.apiKey == "" {
		Logger.Error("Subdl API key is not configured")
		return 0
	}

	for _, lang := range languages {
		searchURL := fmt.Sprintf("%s/subtitles?api_key=%s&tmdb_id=%d&type=movie&languages=%s",
			s.baseURL, s.apiKey, tmdbID, lang)

		Logger.Debug(fmt.Sprintf("Searching subtitles for %s: tmdb_id=%d, lang=%s", lang, tmdbID, lang))

		req, err := http.NewRequest("GET", searchURL, nil)
		if err != nil {
			Logger.Error(fmt.Sprintf("Failed to create request for %s: %v", lang, err))
			continue
		}

		resp, err := s.httpClient.Do(req)
		if err != nil {
			Logger.Error(fmt.Sprintf("Failed to search subtitles for %s: %v", lang, err))
			continue
		}

		if resp.StatusCode != http.StatusOK {
			Logger.Warn(fmt.Sprintf("Search failed for %s with status: %d", lang, resp.StatusCode))
			resp.Body.Close()
			continue
		}

		var searchResult SubdlSearchResult
		if err := json.NewDecoder(resp.Body).Decode(&searchResult); err != nil {
			Logger.Error(fmt.Sprintf("Failed to decode search response for %s: %v", lang, err))
			resp.Body.Close()
			continue
		}
		resp.Body.Close()

		Logger.Debug(fmt.Sprintf("Search result for %s: status=%v, subtitles_count=%d", lang, searchResult.Status, len(searchResult.Subtitles)))

		if !searchResult.Status || len(searchResult.Subtitles) == 0 {
			Logger.Warn(fmt.Sprintf("No subtitles found for language: %s (status=%v, count=%d)", lang, searchResult.Status, len(searchResult.Subtitles)))
			continue
		}

		subtitle := searchResult.Subtitles[0]

		downloadURL := fmt.Sprintf("https://dl.subdl.com%s.zip", subtitle.URL)

		Logger.Debug(fmt.Sprintf("Downloading subtitle from: %s", downloadURL))

		if err := s.downloadFile(downloadURL, outputDir, lang); err != nil {
			Logger.Error(fmt.Sprintf("Failed to download subtitle file for %s: %v", lang, err))
			continue
		}

		Logger.Info(fmt.Sprintf("Successfully downloaded subtitle for language: %s", lang))
		downloadedCount++

		if downloadedCount >= 5 {
			break
		}
	}

	return downloadedCount
}

func (s *SubtitleService) downloadFile(url, outputDir, language string) error {
	resp, err := s.httpClient.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status: %d", resp.StatusCode)
	}

	zipData, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	zipReader, err := zip.NewReader(bytes.NewReader(zipData), int64(len(zipData)))
	if err != nil {
		return fmt.Errorf("failed to read zip archive: %w", err)
	}

	extracted := false
	for _, file := range zipReader.File {
		if strings.HasSuffix(strings.ToLower(file.Name), ".srt") {
			rc, err := file.Open()
			if err != nil {
				Logger.Warn(fmt.Sprintf("Failed to open file %s in zip: %v", file.Name, err))
				continue
			}

			outputPath := filepath.Join(outputDir, fmt.Sprintf("%s.srt", language))
			outFile, err := os.Create(outputPath)
			if err != nil {
				rc.Close()
				return fmt.Errorf("failed to create output file: %w", err)
			}

			_, err = io.Copy(outFile, rc)
			rc.Close()
			outFile.Close()

			if err != nil {
				return fmt.Errorf("failed to write file: %w", err)
			}

			Logger.Info(fmt.Sprintf("Saved subtitle to: %s", outputPath))
			extracted = true
			break
		}
	}

	if !extracted {
		return fmt.Errorf("no SRT file found in zip archive")
	}

	return nil
}
