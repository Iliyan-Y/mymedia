package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Job represents a download job
type Job struct {
	ID       string `json:"id"`
	URL      string `json:"url"`
	Status   string `json:"status"`
	Progress string `json:"progress"`
}

// Shared job store
var (
	jobs   = make(map[string]*Job)
	jobsMu sync.Mutex
)

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
		cmd := exec.Command("yt-dlp", "--newline", "-f", "best",  "-o", "../../media_library/%(title)s.%(ext)s", videoURL)
		stdout, _ := cmd.StdoutPipe()
		_ = cmd.Start()

		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			jobsMu.Lock()
			//println(line)
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

		fmt.Fprintf(w, "data: %s\n\n", job.Progress)
		flusher.Flush()

		if job.Status == "done" || job.Status == "failed" {
			return
		}

		time.Sleep(1 * time.Second)
	}
}

func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Access-Control-Allow-Origin", "*")
        w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

        if r.Method == "OPTIONS" {
            w.WriteHeader(http.StatusOK)
            return
        }

        next(w, r)
    }
}

func main() {
	http.HandleFunc("/download", corsMiddleware(downloadHandler))
	http.HandleFunc("/progress", corsMiddleware(progressHandler))
	http.HandleFunc("/progress/stream", corsMiddleware(sseProgressHandler))

	fmt.Println("🚀 Server running on http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}
