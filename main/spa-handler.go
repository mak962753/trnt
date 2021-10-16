package main

import (
	"net/http"
	"os"
	"path/filepath"
)

type spaHandler struct {
	staticPath string
	indexPath  string
}

func (s spaHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// get the absolute path to prevent directory traversal
	path := filepath.Clean(r.URL.Path)
	// prepend the static dir path
	path = filepath.Join(s.staticPath, path)

	path, err := filepath.Abs(path)

	if err != nil {
		// failed to get the absolute path - respond with a 400 bad request
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// check file exists
	if _, err = os.Stat(path); os.IsNotExist(err) {
		http.ServeFile(w, r, s.indexPath)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.FileServer(http.Dir(s.staticPath)).ServeHTTP(w, r)
}
