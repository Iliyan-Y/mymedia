package downloader

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"mymedia/common/cors"

	"github.com/google/uuid"
)

// Job represents a download job
type Job struct {
	ID       string `json:"id"`
	URL      string `json:"url"`
	Status   string `json:"status"`
	Progress string `json:"progress"`
	FileName string `json:"filename"`
}

// Shared job store
var (
	jobs   = make(map[string]*Job)
	jobsMu sync.Mutex
)

func extractFilenameFromLine(line string) string {
	if !strings.Contains(line, "/media_library/") {
		return ""
	}

	// Regex to match /media_library/<filename>
	re := regexp.MustCompile(`/media_library/(.+?)"`)
	matches := re.FindStringSubmatch(line)
	if len(matches) < 2 {
		return ""
	}

	// Clean up and return just the filename
	filename := filepath.Base(matches[1])
	return filename
}

// startDownloadJob launches yt-dlp in a goroutine
func startDownloadJob(videoURL string) string {
	id := uuid.NewString()
	job := &Job{
		ID:       id,
		URL:      videoURL,
		Status:   "started",
		Progress: "starting...",
	}

	jobsMu.Lock()
	jobs[id] = job
	jobsMu.Unlock()

	go func() {
		// yt-dlp with --newline prints each progress line
		cmd := exec.Command("yt-dlp", "--newline", "-f", "best", "-o", "/media_library/%(title)s.%(ext)s", videoURL)
		stdout, _ := cmd.StdoutPipe()
		_ = cmd.Start()

		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()

			if name := extractFilenameFromLine(line); name != "" {
				jobsMu.Lock()
				job.FileName = name
				jobsMu.Unlock()
			}

			jobsMu.Lock()
			job.Progress = line
			jobsMu.Unlock()
		}

		err := cmd.Wait()
		jobsMu.Lock()
		if err != nil {
			job.Status = "failed"
			job.Progress = fmt.Sprintf("Error: %v", err)
		} else {
			job.Status = "done"
			job.Progress = "Download complete!"
		}
		jobsMu.Unlock()
	}()

	return id
}

// POST or GET /download?url=<video_url>
func downloadHandler(w http.ResponseWriter, r *http.Request) {
	videoURL := r.URL.Query().Get("url")
	if videoURL == "" {
		http.Error(w, "missing 'url' parameter", http.StatusBadRequest)
		return
	}

	id := startDownloadJob(videoURL)
	resp := map[string]string{"job_id": id}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// GET /progress?id=<job_id>
func progressHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")

	jobsMu.Lock()
	job, ok := jobs[id]
	jobsMu.Unlock()

	if !ok {
		http.Error(w, "job not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(job)
}

// GET /progress/stream?id=<job_id> — real-time SSE
func sseProgressHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "missing 'id' parameter", http.StatusBadRequest)
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	for {
		jobsMu.Lock()
		job, ok := jobs[id]
		jobsMu.Unlock()

		if !ok {
			fmt.Fprintf(w, "event: error\ndata: job not found\n\n")
			flusher.Flush()
			return
		}

		// Build SSE payload: include filename when available
		payload := map[string]string{"progress": job.Progress}
		if job.FileName != "" {
			payload["filename"] = job.FileName
		}
		b, _ := json.Marshal(payload)
		fmt.Fprintf(w, "data: %s\n\n", b)

		flusher.Flush()

		if job.Status == "done" || job.Status == "failed" {
			return
		}

		time.Sleep(1 * time.Second)
	}
}

func RegisterHandlers(mux *http.ServeMux) {
	mux.HandleFunc("/api/download", cors.Middleware(downloadHandler))
	mux.HandleFunc("/api/progress", cors.Middleware(progressHandler))
	mux.HandleFunc("/api/progress/stream", cors.Middleware(sseProgressHandler))
}
