package main

import (
	"crypto/sha256"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"time"

	"internal/common"
	"internal/prefetcher"
)

func main() {

	// TODO: use flags
	if len(os.Args) < 4 {
		log.Fatalf("Usage: %s <PREFETCHERS.JSON> <REPO_PATH> <SERVER_ADDR>", os.Args[0])
	}

	prefetchConfigFile := os.Args[1]
	srcDir := os.Args[2]
	serverAddr := os.Args[3]

	l := common.NewLoggerWithPrefixAndColor("Main: ")

	prefetchConfig, err := common.ReadPrefetchConfigJson(prefetchConfigFile)
	if err != nil {
		l.Fatalf("failed to load prefetch config: %v", err)
	}

	prefetchers, err := prefetcher.CreatePrefetchersFromConfig(srcDir, prefetchConfig)
	if err != nil {
		log.Fatalf("Failed to generate prefetchers: %v", err)
	}

	log.Print(common.Imafish())

	prefetchItems, err := prefetcher.AnalyzePrefetchItems(prefetchers, bazelCacheDir())
	if err != nil {
		log.Fatalf("Failed to analyze prefetch items: %v", err)
	}
	l.Printf("Got %d items.", len(prefetchItems))

	successful := 0
	total := 0
	for _, item := range prefetchItems {
		err := handlePrefetchItem(item, srcDir, serverAddr)
		if err != nil {
			log.Printf("Failed to download file for item %s: %v", item.Url, err)
		} else {
			log.Printf("Successfully handled file for item %s", item.Url)
			successful += 1
		}
		total += 1
	}

	common.LogSeparator(fmt.Sprintf("DONE: %d/%d", successful, total))
}

func handlePrefetchItem(item *prefetcher.PrefetchItem, srcDir string, serverAddr string) error {
	downloadURL := fmt.Sprintf("%s/files/%s/file", serverAddr, item.Hash)
	err := downloadFile(downloadURL, item)
	if err != nil {
		log.Printf("Error: failed to download file, %v", err)
		return err
	}

	match, err := compareFileHash(item)
	if err != nil {
		log.Printf("Error: failed to compare file hash of %s: %v", item.Path, err)
		return err
	}
	if !match {
		log.Printf("Error: file hash doesn't match for %s", item.Path)
		return err
	}

	err = putFileIntoBazelCache(item, bazelCacheDir())
	if err != nil {
		log.Printf("Error: failed to put file %s into bazel cache: %v", item.Path, err)
		return err
	}

	return nil
}

func bazelCacheDir() string {
	username := os.Getenv("USER")
	bazelCachePath := path.Join(os.Getenv("HOME"), ".cache/bazel", fmt.Sprintf("_bazel_%s", username), "cache/repos/v1")
	return bazelCachePath
}

func downloadFile(url string, item *prefetcher.PrefetchItem) error {
	log.Printf("Downloading file from URL: %s", url)

	// Create a temporary file
	tempFile, err := os.CreateTemp("", "prefetcher_*")
	if err != nil {
		log.Printf("Failed to create temporary file: %v", err)
		return err
	}
	defer tempFile.Close()

	log.Printf("Downloading from %s to %s", url, tempFile.Name())

	// Perform the HTTP GET request
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("Failed to download file from URL %s: %v", url, err)
		return err
	}
	defer resp.Body.Close()

	// Check for HTTP errors
	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("failed to download file, HTTP status: %s", resp.Status)
		log.Print(err)
		return err
	}

	buffer := make([]byte, 1024000)
	for {
		// Read data from the response body
		n, err := resp.Body.Read(buffer)
		if err != nil && err != io.EOF {
			log.Printf("Error reading response body: %v", err)
			return err
		}

		// Break the loop if we reach the end of the file
		if n == 0 {
			break
		}

		// Write the data to the temporary file
		_, writeErr := tempFile.Write(buffer[:n])
		if writeErr != nil {
			log.Printf("Error writing to temporary file: %v", writeErr)
			return writeErr
		}

		// Sleep for 10ms
		time.Sleep(5 * time.Millisecond)
	}
	_, err = io.Copy(tempFile, resp.Body)
	if err != nil {
		log.Printf("Failed to write to temporary file: %v", err)
		return err
	}

	// Update the item's Path to the temporary file path
	item.Path = tempFile.Name()
	item.HashOfUrl = fmt.Sprintf("%x", sha256.Sum256([]byte(item.Url)))
	log.Printf("File downloaded successfully to: %s", item.Path)

	return nil
}

func compareFileHash(item *prefetcher.PrefetchItem) (bool, error) {
	hash, err := common.HashOfFile(item.Path)
	if err != nil {
		log.Printf("Failed to calculate file path: %v", err)
		err = fmt.Errorf("failed to calculate file path: %w", err)
		return false, err
	}

	return hash == item.Hash, nil
}

func putFileIntoBazelCache(item *prefetcher.PrefetchItem, cacheDir string) error {
	log.Printf("Placing to bazel cache")
	cacheDirInside := path.Join(cacheDir, "content_addressable", "sha256")
	outerDir := path.Join(cacheDirInside, item.Hash)
	innerFile := path.Join(outerDir, "file")
	hashFilePath := path.Join(outerDir, fmt.Sprintf("id-%s", item.HashOfUrl))
	os.MkdirAll(outerDir, 0755)

	hashFile, err := os.Create(hashFilePath)
	if err != nil {
		log.Printf("Failed to create file %s", hashFilePath)
		item.Error = fmt.Errorf("failed to create file %s, error is: %v", hashFilePath, err)
		return err
	}
	hashFile.Close()

	err = os.Rename(item.Path, innerFile)
	if err != nil {
		err = fmt.Errorf("failed to move from %s to %s, error: %s", item.Path, innerFile, err)
		log.Print(err.Error())
		item.Error = err
		return err
	}

	log.Printf("File moved to bazel cache: %s", innerFile)
	item.Path = outerDir
	return nil
}
