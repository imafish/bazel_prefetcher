package httpserver

import (
	"internal/common"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
)

func serveFiles(config *common.ServerConfig) func(http.ResponseWriter, *http.Request) {
	// Set the root directory to serve
	rootDir := path.Join(config.Server.Workdir, "data")

	// Create a file server that preserves paths
	return func(w http.ResponseWriter, r *http.Request) {
		l := common.NewLoggerWithPrefixAndColor("FileServer: ")
		// Clean the path to prevent directory traversal
		requestedPath := filepath.Clean(r.URL.Path)
		l.Print(requestedPath)
		requestedPath = strings.TrimPrefix(requestedPath, "/files")
		fullPath := filepath.Join(rootDir, requestedPath)
		l.Printf("requested path `%s` mapped to `%s`", requestedPath, fullPath)

		// Check if the file exists
		fileInfo, err := os.Stat(fullPath)
		if err != nil {
			if os.IsNotExist(err) {
				http.NotFound(w, r)
				return
			}
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// If it's a directory, show directory listing
		if fileInfo.IsDir() {
			serveDirListing(w, r, fullPath, requestedPath)
			return
		}

		// For files, set proper headers and serve
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Content-Disposition", "attachment; filename="+filepath.Base(fullPath))
		w.Header().Set("Content-Length", strconv.Itoa(int(fileInfo.Size())))
		http.ServeFile(w, r, fullPath)
	}
}

func serveDirListing(w http.ResponseWriter, _ *http.Request, dirPath, webPath string) {
	l := common.NewLoggerWithPrefixAndColor("DirListing: ")
	// Open the directory
	dir, err := os.Open(dirPath)
	if err != nil {
		http.Error(w, "Error reading directory", http.StatusInternalServerError)
		return
	}
	defer dir.Close()

	// Read directory contents
	fileInfos, err := dir.Readdir(-1)
	if err != nil {
		http.Error(w, "Error reading directory contents", http.StatusInternalServerError)
		return
	}

	// Generate HTML listing
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if webPath == "" {
		webPath = "/"
	}
	html := "<html><head><title>Directory listing for " + webPath + "</title></head><body>"
	html += "<h1>Directory listing for " + webPath + "</h1><hr><ul>"

	l.Printf("Listing directory: %s, %s", webPath, dirPath)
	// Add parent directory link if not at root
	if webPath != "/" {
		parentPath := filepath.Dir(webPath)
		if parentPath == "." {
			parentPath = "/"
		}
		html += "<li><a href=\"" + parentPath + "\">../</a></li>"
	}

	webPath = strings.TrimPrefix(webPath, "/")
	for _, fi := range fileInfos {
		name := fi.Name()
		if strings.HasPrefix(name, ".") {
			continue // Skip hidden files
		}

		linkPath := filepath.Join(webPath, name)
		if fi.IsDir() {
			html += "<li><a href=\"" + linkPath + "\">" + name + "/</a></li>"
		} else {
			html += "<li><a href=\"" + linkPath + "\" download>" + name + "</a></li>"
		}
	}

	html += "</ul><hr></body></html>"
	w.Write([]byte(html))
}
