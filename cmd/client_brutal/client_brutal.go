package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"time"

	"internal/common"
)

func main() {
	// TODO: use flags
	if len(os.Args) < 2 {
		log.Fatalf("Usage: %s <SERVER_ADDR> [<REPOSITORY_CACHE_DIR>]", os.Args[0])
	}

	l := common.NewLoggerWithPrefixAndColor("Main: ")
	serverAddr := os.Args[1]
	repositoryCachePath := path.Join(os.Getenv("HOME"), ".cache/bazel", fmt.Sprintf("_bazel_%s", os.Getenv("USER")), "cache/repos/v1")
	if len(os.Args) == 3 {
		repositoryCachePath = os.Args[2]
		l.Printf("Using repository cache path: %s", repositoryCachePath)
	} else {
		l.Printf("Using default repository cache path: %s", repositoryCachePath)
	}

	log.Print(common.Imafish())

	common.LogSeparator("Syncing repository cache...")

	start := time.Now()
	sizeSynced, err := syncRepositoryCache(serverAddr, repositoryCachePath)
	if err != nil {
		l.Printf("Error syncing repository cache: %s", err)
		return
	}
	// Pretty print the synced size
	l.Printf("Synced size: %.2f MB", float64(sizeSynced)/(1024*1024))

	// Only calculate and display time saved if sizeSynced is larger than 100 MB
	if sizeSynced > 100*1024*1024 {
		executionTime := time.Since(start)
		networkSpeed := 200 * 1024 // 200 KB/s in bytes
		timeSaved := float64(sizeSynced) / float64(networkSpeed)
		humanFriendlyTimeSaved := time.Duration(timeSaved * float64(time.Second))
		durationSaved := humanFriendlyTimeSaved - executionTime
		l.Printf("Time saved by using cache: %s", &durationSaved)
	}
}

type fileObj struct {
	Path string `json:"name"`
	Size int64  `json:"size"`
}

func syncRepositoryCache(serverAddr, repositoryCachePath string) (int64, error) {
	l := common.NewLoggerWithPrefixAndColor("Sync: ")
	client := &http.Client{}
	files, err := getFileList(client, serverAddr)
	if err != nil {
		l.Printf("Error getting file list: %s", err)
		return 0, fmt.Errorf("failed to get file list: %w", err)
	}

	totalfiles := len(files)
	var totalSyncedSize int64
	downloadedFiles := 0
	skippedFiles := 0
	for i, file := range files {
		l.SmallSeparator("%d/%d, downloaded: %d, skipped: %d ", i+1, totalfiles, downloadedFiles, skippedFiles)
		l.Printf("Processing file: %s, size: %d", file.Path, file.Size)

		localFilePath := path.Join(repositoryCachePath, file.Path)
		shouldDownload, err := checkFileExists(localFilePath, file.Size)
		if err != nil {
			l.Printf("Error checking file %s: %s", localFilePath, err)
			return 0, fmt.Errorf("failed to check file %s: %w", localFilePath, err)
		}
		if !shouldDownload {
			l.Printf("File %s already exists and matches size, skipping.", file.Path)
			skippedFiles++
			continue // File matches size, skip
		}

		// Download the file
		err = downloadFile(client, serverAddr, file, localFilePath)
		if err != nil {
			return 0, fmt.Errorf("failed to download file %s: %w", file.Path, err)
		}
		l.Printf("File %s downloaded successfully.", file.Path)
		downloadedFiles++
		totalSyncedSize += file.Size
	}

	l.SmallSeparator("Summary:")
	l.Printf("Total files: %d, Downloaded: %d, Skipped: %d", totalfiles, downloadedFiles, skippedFiles)

	return totalSyncedSize, nil
}

func getFileList(client *http.Client, serverAddr string) ([]fileObj, error) {
	l := common.NewLoggerWithPrefixAndColor("GetFileList: ")
	l.Printf("Querying server for file list...")
	resp, err := client.Get(fmt.Sprintf("http://%s/restapi/v1/files", serverAddr))
	if err != nil {
		return nil, fmt.Errorf("failed to query server: %w", err)
	}
	defer resp.Body.Close()
	l.Printf("Server responded with status: %s", resp.Status)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned non-OK status: %s", resp.Status)
	}

	var files []fileObj
	if err := json.NewDecoder(resp.Body).Decode(&files); err != nil {
		return nil, fmt.Errorf("failed to decode server response: %w", err)
	}
	totalfiles := len(files)
	l.Printf("Number of files to sync: %d", totalfiles)
	return files, nil
}

func checkFileExists(localFilePath string, expectedSize int64) (bool, error) {
	l := common.NewLoggerWithPrefixAndColor("CheckFile: ")
	l.Printf("Checking if file %s exists...", localFilePath)
	if fileInfo, err := os.Stat(localFilePath); err == nil {
		if fileInfo.Size() == expectedSize {
			l.Printf("File %s already exists and matches size", localFilePath)
			return false, nil // File exists and matches size
		}

		// File exists but size does not match, replace it
		l.Printf("File %s exists but size does not match, replacing.", localFilePath)
		if err := os.Remove(localFilePath); err != nil {
			return false, fmt.Errorf("failed to remove mismatched file %s: %w", localFilePath, err)
		}
	}

	if err := os.MkdirAll(path.Dir(localFilePath), 0755); err != nil {
		return false, fmt.Errorf("failed to create directories for %s: %w", localFilePath, err)
	}

	if expectedSize == 0 {
		// Create an empty file
		file, err := os.Create(localFilePath)
		if err != nil {
			return false, fmt.Errorf("failed to create empty file %s: %w", localFilePath, err)
		}
		file.Close()
		return false, nil // File created successfully
	}

	return true, nil
}

func downloadFile(client *http.Client, serverAddr string, file fileObj, localPath string) error {
	l := common.NewLoggerWithPrefixAndColor("Download: ")
	fileUrl := fmt.Sprintf("http://%s/files/%s", serverAddr, file.Path)
	l.Printf("Downloading file from URL: %s to %s", fileUrl, localPath)

	resp, err := client.Get(fileUrl)
	if err != nil {
		return fmt.Errorf("failed to download file %s: %w", fileUrl, err)
	}
	defer resp.Body.Close()

	l.Printf("Server responded with status: %s", resp.Status)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download file %s, server returned: %s", localPath, resp.Status)
	}

	buffer := make([]byte, 10*1024*1024) // 10 MB buffer
	outFile, err := os.Create(localPath)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", localPath, err)
	}
	defer outFile.Close()

	var n int64
	var m int64 = -1
	const OneHundredM = 100 * 1024 * 1024 // 100 MB
	start := time.Now()
	for {
		bytesRead, readErr := resp.Body.Read(buffer)
		if bytesRead > 0 {
			bytesWritten, writeErr := outFile.Write(buffer[:bytesRead])
			if writeErr != nil {
				return fmt.Errorf("failed to write to file %s: %w", localPath, writeErr)
			}
			n += int64(bytesWritten)
			if n/OneHundredM > m && file.Size > OneHundredM { // Check if we've crossed a 100MB boundary
				m = n / OneHundredM
				elapsed := time.Since(start)
				averageSpeed := n / int64(elapsed.Seconds()+1)
				l.Printf("Downloaded: %s / %s, Time: %s, Avg Speed: %s/s",
					prettyPrintBytes(n), prettyPrintBytes(file.Size), elapsed, prettyPrintBytes(averageSpeed))
			}
			// time.Sleep(5 * time.Millisecond)
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return fmt.Errorf("failed to read from response body for file %s: %w", localPath, readErr)
		}
	}

	return nil
}

func prettyPrintBytes(n int64) string {
	if n < 1024 {
		return fmt.Sprintf("%d B", n)
	} else if n < 1024*1024 {
		return fmt.Sprintf("%.2f KB", float64(n)/1024)
	} else if n < 1024*1024*1024 {
		return fmt.Sprintf("%.2f MB", float64(n)/(1024*1024))
	} else {
		return fmt.Sprintf("%.2f GB", float64(n)/(1024*1024*1024))
	}
}
