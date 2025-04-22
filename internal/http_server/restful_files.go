package httpserver

import (
	"encoding/json"
	"fmt"
	"internal/common"
	"net/http"
	"os"
	"path"
	"path/filepath"
)

// FileInfo represents a file with its name and size
type FileInfo struct {
	Name string `json:"name"`
	Size int64  `json:"size"`
}

// getAllFilesHandler handles GET requests to /restapi/v1/allfiles
func getAllFilesHandler(w http.ResponseWriter, r *http.Request, config *common.ServerConfig) {
	l := common.NewLoggerWithPrefixAndColor("restful_server.getAllFilesHandler: ")
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	dirPath := path.Join(config.Server.Workdir, "data")
	// Verify the directory exists
	info, err := os.Stat(dirPath)
	if err != nil {
		l.Printf("Directory not found: %v", err)
		http.Error(w, fmt.Sprintf("Directory not found: %v", err), http.StatusNotFound)
		return
	}
	if !info.IsDir() {
		l.Printf("Path is not a directory: %s", dirPath)
		http.Error(w, "Path is not a directory", http.StatusBadRequest)
		return
	}

	// Collect file information recursively
	var fileInfos []FileInfo
	err = filepath.Walk(dirPath, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() { // Only include files, not directories
			relPath, err := filepath.Rel(dirPath, filePath)
			if err != nil {
				return err
			}
			fileInfos = append(fileInfos, FileInfo{
				Name: relPath,
				Size: info.Size(),
			})
		}
		return nil
	})
	if err != nil {
		l.Printf("Error walking directory: %v", err)
		http.Error(w, fmt.Sprintf("Error walking directory: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	l.Printf("Sending %d files in response", len(fileInfos))
	if err := json.NewEncoder(w).Encode(fileInfos); err != nil {
		http.Error(w, fmt.Sprintf("Error encoding response: %v", err), http.StatusInternalServerError)
	}
}

func serveApiV1GetAllFiles(config *common.ServerConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		getAllFilesHandler(w, r, config)
	}
}
