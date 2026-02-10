package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

var buildPath = "./build"

func main() {
	mux := http.NewServeMux()

	staticFileServer := http.FileServer(http.Dir(buildPath))

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") || strings.HasPrefix(r.URL.Path, "/streams/") {
			http.NotFound(w, r) // let the specific handlers for these paths take over
			return
		}

		_, err := os.Stat(filepath.Join(buildPath, r.URL.Path))

		if os.IsNotExist(err) {
			http.ServeFile(w, r, filepath.Join(buildPath, "index.html"))
		} else {
			staticFileServer.ServeHTTP(w, r)
		}
	})

	port := ":8080"
	log.Printf("server starting on http://localhost%s", port)
	err :=http.ListenAndServe(port, mux)
	if err != nil {
		log.Fatal(err)
	}
}
