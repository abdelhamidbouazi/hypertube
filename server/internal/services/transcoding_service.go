package services

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"server/internal/models"
)

type TranscodingService struct {
	activeJobs map[string]*models.TranscodeJob
	mu         sync.RWMutex
}

func NewTranscodingService() *TranscodingService {
	return &TranscodingService{
		activeJobs: make(map[string]*models.TranscodeJob),
	}
}

func (ts *TranscodingService) TranscodeIfNeeded(inputPath string, w http.ResponseWriter, r *http.Request) error {
	if _, err := os.Stat(inputPath); os.IsNotExist(err) {
		return fmt.Errorf("file does not exist: %s", inputPath)
	}

	ext := strings.ToLower(filepath.Ext(inputPath))

	if ext == ".mp4" || ext == ".webm" {
		http.ServeFile(w, r, inputPath)
		return nil
	}

	if ext == ".mkv" || ext == ".avi" || ext == ".mov" {
		return ts.streamTranscode(inputPath, w, r)
	}

	return fmt.Errorf("unsupported video format: %s", ext)
}

func (ts *TranscodingService) streamTranscode(inputPath string, w http.ResponseWriter, r *http.Request) error {
	cmd := exec.Command("ffmpeg",
		"-i", inputPath,
		"-c:v", "libx264",
		"-preset", "ultrafast",
		"-c:a", "aac",
		"-movflags", "frag_keyframe+empty_moov+faststart",
		"-f", "mp4",
		"-fflags", "+nobuffer+flush_packets",
		"-flags", "low_delay",
		"-strict", "experimental",
		"-")

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	w.Header().Set("Content-Type", "video/mp4")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Transfer-Encoding", "chunked")

	_, err = io.Copy(w, stdout)
	if err != nil {
		log.Printf("Error streaming video: %v", err)
	}

	return cmd.Wait()
}
