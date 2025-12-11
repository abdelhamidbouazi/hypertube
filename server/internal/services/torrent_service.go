package services

import (
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"server/internal/models"
	"strings"
	"sync"
	"time"

	anacrolixlog "github.com/anacrolix/log"
	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/storage"
	"github.com/anacrolix/torrent/types/infohash"
	"gorm.io/gorm"
)

type TorrentService struct {
	client      *torrent.Client
	Downloads   sync.Map // map[string]*models.TorrentDownload
	downloadDir string
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

	// Disable IPv6 to prevent "network unreachable" errors
	cfg.DisableIPv6 = true
	cfg.DisableUTP = false
	cfg.NoDHT = false
	cfg.DisablePEX = false

	// Configure torrent library to only log Critical level (effectively disabling most logs)
	cfg.Logger = anacrolixlog.Default.FilterLevel(anacrolixlog.Never)

	client, err := torrent.NewClient(cfg)
	if err != nil {
		log.Fatalf("Failed to start torrent client: %v", err)
	}

	ts := &TorrentService{
		client:      client,
		downloadDir: downloadDir,
		db:          db,
	}

	return ts
}

func (ts *TorrentService) GetOrStartDownload(movieID int, infoHash string) (*models.TorrentDownload, error) {
	downloadKey := fmt.Sprintf("%d", movieID)

	var downloadedMovie models.DownloadedMovie
	err := ts.db.Where("movie_id = ?", movieID).First(&downloadedMovie).Error
	if err == nil && downloadedMovie.FilePath != "" {
		if _, err := os.Stat(downloadedMovie.FilePath); err == nil {
			ts.db.Model(&downloadedMovie).Update("last_watched", time.Now())

			return &models.TorrentDownload{
				MovieID:        movieID,
				Progress:       100,
				Status:         "completed",
				StreamReady:    true,
				StreamingReady: true,
				FilePath:       downloadedMovie.FilePath,
				CompletedAt:    &[]time.Time{time.Now()}[0],
			}, nil
		} else {
			// log.Printf("Movie %d marked as downloaded but file missing, deleting database record", movieID)
			ts.db.Delete(&downloadedMovie)
		}
	}

	movieDownloadDir := filepath.Join(ts.downloadDir, fmt.Sprintf("%d", movieID))
	if err := os.MkdirAll(movieDownloadDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	hashBytes, err := hex.DecodeString(infoHash)
	if err != nil {
		return nil, fmt.Errorf("invalid infohash format: %w", err)
	}

	var ih infohash.T
	copy(ih[:], hashBytes)

	// log.Printf("Starting download for movie %d with infohash: %s", movieID, infoHash)

	ops := torrent.AddTorrentOpts{
		InfoHash: ih,
		Storage:  storage.NewFile(movieDownloadDir),
	}

	t, _ := ts.client.AddTorrentOpt(ops)

	ts.addTrackersToTorrent(t)

	<-t.GotInfo()

	mi := t.Metainfo()

	magnet, err := mi.MagnetV2()
	if err != nil {
		return nil, fmt.Errorf("failed to get magnet link: %w", err)
	}

	dl := &models.TorrentDownload{
		Torrent:   t,
		MovieID:   movieID,
		Status:    "initializing",
		StartedAt: time.Now(),
	}

	ts.Downloads.Store(downloadKey, dl)

	go ts.monitorDownload(dl, downloadKey, magnet.String(), movieDownloadDir)

	return dl, nil
}

func (ts *TorrentService) monitorDownload(dl *models.TorrentDownload, downloadKey, magnet, movieDownloadDir string) {
	// log.Printf("Starting monitor for %s", downloadKey)

	select {
	case <-dl.Torrent.GotInfo():
		// log.Printf("Got torrent info for %s: %s", downloadKey, dl.Torrent.Name())
	case <-time.After(3 * time.Minute):
		// log.Printf("Timeout waiting for torrent info for %s", downloadKey)
		ts.handleDownloadError(dl, downloadKey, "metadata timeout")
		return
	}

	videoFile := ts.findLargestVideoFile(dl.Torrent)
	if videoFile == nil {
		// log.Printf("No video file found in torrent for %s", downloadKey)
		ts.handleDownloadError(dl, downloadKey, "no video file found")
		return
	}

	dl.Mu.Lock()
	dl.VideoFile = videoFile
	dl.FilePath = filepath.Join(movieDownloadDir, videoFile.Path())
	dl.RootDir = movieDownloadDir
	dl.Status = "downloading"
	dl.Mu.Unlock()

	// log.Printf("Video file found: %s (%.2f MB)", videoFile.Path(), float64(videoFile.Length())/1024/1024)

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
			// log.Printf("Torrent closed for %s", downloadKey)
			ts.handleDownloadError(dl, downloadKey, "torrent closed")
			return
		case <-ticker.C:
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
				// log.Printf("Stream ready for %s (%.2f MB downloaded)", downloadKey, float64(completed)/1024/1024)
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

				// log.Printf("Download completed for %s", downloadKey)
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
				// log.Printf("Download appears stalled for %s", downloadKey)
				ts.handleDownloadError(dl, downloadKey, "download stalled")
				return
			}

			if currentProgress-lastLoggedPercent >= 5.0 || completed >= total {
				// log.Printf("Download progress for %s: %.2f%% (%.2f/%.2f MB)",
				// 	downloadKey, currentProgress,
				// 	float64(completed)/1024/1024, float64(total)/1024/1024)
				lastLoggedPercent = currentProgress
			}
		}
	}
}

func (ts *TorrentService) handleDownloadError(dl *models.TorrentDownload, downloadKey, reason string) {
	dl.Mu.Lock()
	dl.Status = "error"
	dl.Mu.Unlock()

	ts.Downloads.Delete(downloadKey)

	if dl.Torrent != nil {
		dl.Torrent.Drop()
	}

	// log.Printf("Download error for %s: %s", downloadKey, reason)
}

func (ts *TorrentService) addTrackersToTorrent(t *torrent.Torrent) {
	// Use reliable UDP trackers only to avoid decode errors and IPv6 issues
	trackers := [][]string{
		{"udp://tracker.opentrackr.org:1337/announce"},
		{"udp://open.stealth.si:80/announce"},
		{"udp://tracker.openbittorrent.com:6969/announce"},
		{"udp://opentracker.i2p.rocks:6969/announce"},
		{"udp://tracker.internetwarriors.net:1337/announce"},
		{"udp://tracker.leechers-paradise.org:6969/announce"},
		{"udp://coppersurfer.tk:6969/announce"},
		{"udp://tracker.zer0day.to:1337/announce"},
	}

	t.AddTrackers(trackers)
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
		// log.Printf("Error saving downloaded movie: %v", err)
	}
}

func (ts *TorrentService) RemoveTorrentFiles(movieID int, quality string, hlsOutputDir string) error {
	downloadKey := fmt.Sprintf("%d-%s", movieID, quality)

	dlInterface, ok := ts.Downloads.Load(downloadKey)
	if ok {
		dl, ok := dlInterface.(*models.TorrentDownload)
		if ok && dl.Torrent != nil {
			dl.Torrent.Drop()
		}
		ts.Downloads.Delete(downloadKey)
	}

	if hlsOutputDir != "" {
		if _, err := os.Stat(hlsOutputDir); err == nil {
			if err := os.RemoveAll(hlsOutputDir); err != nil {
				return fmt.Errorf("failed to remove HLS output directory: %w", err)
			}
		}
	}

	return nil
}
