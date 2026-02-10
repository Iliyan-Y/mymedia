package downloader

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
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

const (
	defaultMediaDir = "../media_library"
)

func getOutputDir() string {
	if env := os.Getenv("MEDIA_LIBRARY_DIR"); env != "" {
		return env
	}
	return defaultMediaDir
}

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
	outDir := getOutputDir()
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		log.Printf("download job %s failed to prepare output dir %s: %v", id, outDir, err)
		jobsMu.Lock()
		job.Status = "failed"
		job.Progress = fmt.Sprintf("Unable to write to %s: %v. Ensure this path is writable or set MEDIA_LIBRARY_DIR to a writable path", outDir, err)
		jobsMu.Unlock()
		return id
	}

	log.Printf("starting download job %s for %s", id, videoURL)
	go func() {
		// yt-dlp with --newline prints each progress line
		cmd := exec.Command("yt-dlp", "--newline", "-f", "best", "-o", filepath.Join(outDir, "%(title)s.%(ext)s"), videoURL)
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			log.Printf("download job %s failed to capture stdout: %v", id, err)
			jobsMu.Lock()
			job.Status = "failed"
			job.Progress = fmt.Sprintf("Error preparing download: %v", err)
			jobsMu.Unlock()
			return
		}
		stderr, err := cmd.StderrPipe()
		if err != nil {
			log.Printf("download job %s failed to capture stderr: %v", id, err)
			jobsMu.Lock()
			job.Status = "failed"
			job.Progress = fmt.Sprintf("Error preparing download: %v", err)
			jobsMu.Unlock()
			return
		}

		if err := cmd.Start(); err != nil {
			log.Printf("download job %s failed to start: %v", id, err)
			jobsMu.Lock()
			job.Status = "failed"
			job.Progress = fmt.Sprintf("Error starting download: %v", err)
			jobsMu.Unlock()
			return
		}

		var wg sync.WaitGroup
		wg.Add(2)
		go scanStream(id, job, stdout, &wg)
		go scanStream(id, job, stderr, &wg)
		wg.Wait()

		err = cmd.Wait()
		jobsMu.Lock()
		if err != nil {
			log.Printf("download job %s failed: %v", id, err)
			job.Status = "failed"
			job.Progress = fmt.Sprintf("Error: %v", err)
		} else {
			log.Printf("download job %s completed", id)
			job.Status = "done"
			job.Progress = "Download complete!"
		}
		jobsMu.Unlock()
	}()
	log.Printf("download job %s started", id)
	 
	return id
}

func scanStream(id string, job *Job, reader io.Reader, wg *sync.WaitGroup) {
	defer wg.Done()
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()
		log.Printf("download job %s: %s", id, line)
		if name := extractFilenameFromLine(line); name != "" {
			jobsMu.Lock()
			job.FileName = name
			jobsMu.Unlock()
		}

		jobsMu.Lock()
		job.Progress = line
		jobsMu.Unlock()
	}
	if err := scanner.Err(); err != nil {
		log.Printf("download job %s stream error: %v", id, err)
	}
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
