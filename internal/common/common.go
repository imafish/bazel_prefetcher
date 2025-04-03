package common

import (
	"crypto/sha256"
	"fmt"
	"io"
	"log"
	"os"
)

func FileExists(p string) bool {
	info, err := os.Stat(p)
	if err != nil {
		if os.IsNotExist(err) {
			return false
		} else {
			log.Printf("Unexpected error: %v", err)
			return false
		}
	}
	return !info.IsDir()
}

func PrettyPrintSize(size int64) string {
	if size < 1024 {
		return fmt.Sprintf("%d B", size)
	} else if size < 1024*1024 {
		return fmt.Sprintf("%.1f KB", float64(size)/1024)
	} else if size < 1024*1024*1024 {
		return fmt.Sprintf("%.1f MB", float64(size)/1024/1024)
	} else {
		return fmt.Sprintf("%.1f GB", float64(size)/1024/1024/1024)
	}
}

func HashOfFile(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}
