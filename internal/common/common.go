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
