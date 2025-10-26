package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"mymedia/common/cors"
)

const mediaDir = "../../media_library"

func main() {
    http.HandleFunc("/api/media", cors.Middleware(listMedia))
    http.HandleFunc("/api/media/stream/", cors.Middleware(streamMedia))
    fmt.Println("Server running on http://localhost:8081")
    http.ListenAndServe(":8081", nil)
}

// listMedia returns a list of available files
func listMedia(w http.ResponseWriter, r *http.Request) {
    entries, err := os.ReadDir(mediaDir)
    if err != nil {
        http.Error(w, err.Error(), 500)
        return
    }

    var files []string
    for _, e := range entries {
        if !e.IsDir() && isMediaFile(e.Name()) {
            files = append(files, e.Name())
        }
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(files)
}

// streamMedia serves media file with Range support
func streamMedia(w http.ResponseWriter, r *http.Request) {
    fileName := strings.TrimPrefix(r.URL.Path, "/api/media/stream/")
    filePath := filepath.Join(mediaDir, filepath.Clean(fileName))

    f, err := os.Open(filePath)
    if err != nil {
        http.Error(w, "File not found", 404)
        return
    }
    defer f.Close()

    fi, _ := f.Stat()
    w.Header().Set("Content-Type", detectMime(filePath))
    http.ServeContent(w, r, fileName, fi.ModTime(), f)
}

// optional helper functions
func detectMime(path string) string {
    if strings.HasSuffix(path, ".mp4") {
        return "video/mp4"
    }
    if strings.HasSuffix(path, ".mp3") {
        return "audio/mpeg"
    }
    return "application/octet-stream"
}

func isMediaFile(name string) bool {
    return strings.HasSuffix(name, ".mp4") || strings.HasSuffix(name, ".mp3") || strings.HasSuffix(name, ".webm")
}
