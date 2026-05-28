package main

import (
	"net/http"
	"os"
)

// mediaHandler serves a local file for in-app playback (e.g. the Trim previewer's
// <video>). It supports HTTP range requests via http.ServeContent, which the
// webview needs for seeking/scrubbing. Wired into the Wails AssetServer
// middleware in main.go at the "/media" path.
func mediaHandler(w http.ResponseWriter, r *http.Request) {
	p := r.URL.Query().Get("p")
	if p == "" {
		http.Error(w, "missing 'p'", http.StatusBadRequest)
		return
	}
	f, err := os.Open(p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	defer f.Close()
	st, err := f.Stat()
	if err != nil || st.IsDir() {
		http.Error(w, "not a file", http.StatusBadRequest)
		return
	}
	http.ServeContent(w, r, st.Name(), st.ModTime(), f)
}
