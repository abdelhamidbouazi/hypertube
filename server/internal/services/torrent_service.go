package services

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"server/internal/models"

	"github.com/anacrolix/torrent"
	"gorm.io/gorm"
)

type TorrentService struct {
	client      *torrent.Client
	Downloads   map[string]*models.TorrentDownload
	downloadDir string
	Mu          sync.RWMutex
	db          *gorm.DB
}

func NewTorrentService(downloadDir string, db *gorm.DB) *TorrentService {
	if err := os.MkdirAll(downloadDir, 0755); err != nil {
		log.Fatalf("Failed to create download directory: %v", err)
	}

	cfg := torrent.NewDefaultClientConfig()
	cfg.DataDir = downloadDir
	cfg.Seed = true
	cfg.Debug = false
	cfg.ListenPort = 0

	cfg.DisableUTP = false
	cfg.NoDHT = false
	cfg.DisablePEX = false

	client, err := torrent.NewClient(cfg)
	if err != nil {
		log.Fatalf("Failed to start torrent client: %v", err)
	}

	ts := &TorrentService{
		client:      client,
		Downloads:   make(map[string]*models.TorrentDownload),
		downloadDir: downloadDir,
		db:          db,
	}

	go ts.cleanupOldMovies()
	return ts
}

func (ts *TorrentService) cleanupOldMovies() {
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		cutoffTime := time.Now().AddDate(0, -1, 0)

		var oldMovies []models.DownloadedMovie
		if err := ts.db.Where("last_watched < ?", cutoffTime).Find(&oldMovies).Error; err != nil {
			log.Printf("Error querying old movies: %v", err)
			continue
		}

		for _, movie := range oldMovies {
			if err := os.Remove(movie.FilePath); err != nil {
				log.Printf("Error removing file %s: %v", movie.FilePath, err)
			}

			if err := ts.db.Delete(&movie).Error; err != nil {
				log.Printf("Error deleting movie record: %v", err)
			} else {
				log.Printf("Cleaned up movie %d (unwatched for >1 month)", movie.MovieID)
			}
		}
	}
}

func (ts *TorrentService) GetOrStartDownload(movieID int, magnet string, quality string) (*models.TorrentDownload, error) {
	downloadKey := fmt.Sprintf("%d-%s", movieID, quality)

	ts.Mu.Lock()
	defer ts.Mu.Unlock()

	log.Printf("Starting download for movie %d with quality %s", movieID, quality)

	var downloadedMovie models.DownloadedMovie
	err := ts.db.Where("movie_id = ? AND quality = ?", movieID, quality).First(&downloadedMovie).Error
	if err == nil && downloadedMovie.FilePath != "" {
		if _, err := os.Stat(downloadedMovie.FilePath); err == nil {
			log.Printf("Movie %d with quality %s already downloaded", movieID, quality)
			ts.db.Model(&downloadedMovie).Update("last_watched", time.Now())

			return &models.TorrentDownload{
				MovieID:        movieID,
				Quality:        quality,
				Progress:       100,
				Status:         "completed",
				StreamReady:    true,
				StreamingReady: true,
				FilePath:       downloadedMovie.FilePath,
				CompletedAt:    &[]time.Time{time.Now()}[0],
			}, nil
		} else {
			// File was deleted, clean up database record
			log.Printf("Movie %d with quality %s marked as downloaded but file missing, deleting database record", movieID, quality)
			ts.db.Delete(&downloadedMovie)
		}
	}

	magnet = ts.addTrackersToMagnet(magnet)

	log.Printf("Adding magnet to client: %s", magnet)
	t, err := ts.client.AddMagnet(magnet)
	if err != nil {
		log.Printf("Error adding magnet: %v", err)
		return nil, fmt.Errorf("failed to add magnet: %w", err)
	}

	dl := &models.TorrentDownload{
		Torrent:   t,
		MovieID:   movieID,
		Quality:   quality,
		Status:    "initializing",
		StartedAt: time.Now(),
	}

	ts.Downloads[downloadKey] = dl

	go ts.monitorDownload(dl, downloadKey, magnet)

	return dl, nil
}

func (ts *TorrentService) monitorDownload(dl *models.TorrentDownload, downloadKey, magnet string) {
	log.Printf("Starting monitor for %s", downloadKey)

	select {
	case <-dl.Torrent.GotInfo():
		log.Printf("Got torrent info for %s: %s", downloadKey, dl.Torrent.Name())
	case <-time.After(3 * time.Minute):
		log.Printf("Timeout waiting for torrent info for %s", downloadKey)
		ts.handleDownloadError(dl, downloadKey, "metadata timeout")
		return
	}

	videoFile := ts.findLargestVideoFile(dl.Torrent)
	if videoFile == nil {
		log.Printf("No video file found in torrent for %s", downloadKey)
		ts.handleDownloadError(dl, downloadKey, "no video file found")
		return
	}

	dl.Mu.Lock()
	dl.VideoFile = videoFile
	dl.FilePath = filepath.Join(ts.downloadDir, videoFile.Path())
	dl.Status = "downloading"
	dl.Mu.Unlock()

	log.Printf("Video file found: %s (%.2f MB)", videoFile.Path(), float64(videoFile.Length())/1024/1024)

	ts.prioritizeVideoFile(videoFile)

	dl.Torrent.DownloadAll()

	ts.monitorProgress(dl, downloadKey, magnet)
}

func (ts *TorrentService) findLargestVideoFile(t *torrent.Torrent) *torrent.File {
	var videoFile *torrent.File
	videoExts := []string{".mp4", ".mkv", ".avi", ".mov", ".wmv", ".webm", ".m4v"}

	for _, f := range t.Files() {
		ext := strings.ToLower(filepath.Ext(f.Path()))
		for _, validExt := range videoExts {
			if ext == validExt {
				if videoFile == nil || f.Length() > videoFile.Length() {
					videoFile = f
				}
				break
			}
		}
	}

	return videoFile
}

func (ts *TorrentService) prioritizeVideoFile(videoFile *torrent.File) {
	totalPieces := videoFile.Torrent().NumPieces()

	firstPieces := int(float64(totalPieces) * 0.1)
	for i := 0; i < firstPieces && i < totalPieces; i++ {
		videoFile.Torrent().Piece(i).SetPriority(torrent.PiecePriorityHigh)
	}
	lastPieces := int(float64(totalPieces) * 0.1)
	for i := totalPieces - lastPieces; i < totalPieces; i++ {
		if i >= 0 {
			videoFile.Torrent().Piece(i).SetPriority(torrent.PiecePriorityHigh)
		}
	}

	videoFile.SetPriority(torrent.PiecePriorityNormal)
}

func (ts *TorrentService) monitorProgress(dl *models.TorrentDownload, downloadKey, magnet string) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	streamingThreshold := int64(5 * 1024 * 1024)
	noProgressCount := 0
	lastProgress := float64(0)
	lastLoggedPercent := -5.0 // Initialize to ensure first log happens

	for {
		select {
		case <-dl.Torrent.Closed():
			log.Printf("Torrent closed for %s", downloadKey)
			ts.handleDownloadError(dl, downloadKey, "torrent closed")
			return
		case <-ticker.C:
			// Update progress
			completed := dl.Torrent.BytesCompleted()
			total := dl.Torrent.Length()

			dl.Mu.Lock()
			dl.Progress = float64(completed) / float64(total) * 100
			currentProgress := dl.Progress
			dl.Mu.Unlock()
			if !dl.StreamingReady && completed > streamingThreshold {
				dl.Mu.Lock()
				dl.StreamingReady = true
				dl.StreamReady = true
				dl.Status = "streaming"
				dl.Mu.Unlock()
				log.Printf("Stream ready for %s (%.2f MB downloaded)", downloadKey, float64(completed)/1024/1024)
				if dl.VideoFile != nil {
					dl.VideoFile.SetPriority(torrent.PiecePriorityNormal)
				}
			}
			if completed >= total {
				dl.Mu.Lock()
				dl.Status = "completed"
				now := time.Now()
				dl.CompletedAt = &now
				dl.Mu.Unlock()

				log.Printf("Download completed for %s", downloadKey)
				ts.saveDownloadedMovie(dl, magnet)
				return
			}

			if currentProgress == lastProgress {
				noProgressCount++
			} else {
				noProgressCount = 0
			}
			lastProgress = currentProgress
			if noProgressCount > 150 && currentProgress < 5 { // 150 * 2 seconds = 5 minutes
				log.Printf("Download appears stalled for %s", downloadKey)
				ts.handleDownloadError(dl, downloadKey, "download stalled")
				return
			}

			// Only log at 5% increments or when download completes
			if currentProgress-lastLoggedPercent >= 5.0 || completed >= total {
				log.Printf("Download progress for %s: %.2f%% (%.2f/%.2f MB)",
					downloadKey, currentProgress,
					float64(completed)/1024/1024, float64(total)/1024/1024)
				lastLoggedPercent = currentProgress
			}
		}
	}
}

func (ts *TorrentService) handleDownloadError(dl *models.TorrentDownload, downloadKey, reason string) {
	dl.Mu.Lock()
	dl.Status = "error"
	dl.Mu.Unlock()

	ts.Mu.Lock()
	delete(ts.Downloads, downloadKey)
	ts.Mu.Unlock()

	if dl.Torrent != nil {
		dl.Torrent.Drop()
	}

	log.Printf("Download error for %s: %s", downloadKey, reason)
}

func (ts *TorrentService) addTrackersToMagnet(magnet string) string {
	trackers := []string{
		"udp://tracker.openbittorrent.com:6969",
		"udp://tracker.opentrackr.org:1337",
		"udp://9.rarbg.to:2710",
		"udp://exodus.desync.com:6969",
		"udp://tracker.cyberia.is:6969",
		"udp://tracker.torrent.eu.org:451",
		"udp://tracker.dler.org:6969",
		"udp://opentracker.i2p.rocks:6969",
	}

	for _, tracker := range trackers {
		if !strings.Contains(magnet, tracker) {
			magnet += "&tr=" + url.QueryEscape(tracker)
		}
	}

	return magnet
}

func (ts *TorrentService) saveDownloadedMovie(dl *models.TorrentDownload, magnet string) {
	fileInfo, _ := os.Stat(dl.FilePath)
	fileSize := int64(0)
	if fileInfo != nil {
		fileSize = fileInfo.Size()
	}

	downloadedMovie := models.DownloadedMovie{
		MovieID:      dl.MovieID,
		Quality:      dl.Quality,
		FilePath:     dl.FilePath,
		MagnetLink:   magnet,
		DownloadedAt: time.Now(),
		LastWatched:  time.Now(),
		FileSize:     fileSize,
	}

	if err := ts.db.Where("movie_id = ? AND quality = ?", dl.MovieID, dl.Quality).
		Assign(&downloadedMovie).
		FirstOrCreate(&downloadedMovie).Error; err != nil {
		log.Printf("Error saving downloaded movie: %v", err)
	}
}
