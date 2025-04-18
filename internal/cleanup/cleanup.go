package cleanup

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"internal/common"
)

type Cleanup struct {
	Workdir      string
	MaxSize      int64
	TolerantSize int64
	MaxAge       int64

	currentSize int64
	dirInfo     []fileInfo
}

type fileInfo struct {
	Path    string
	ModTime int64
	IsDir   bool
	Size    int64
}

func (c *Cleanup) Run() error {
	// Check if the workdir exists
	if _, err := os.Stat(c.Workdir); os.IsNotExist(err) {
		return fmt.Errorf("workdir does not exist: %s", c.Workdir)
	}

	// Get the current size of the workdir
	var err error
	c.currentSize, c.dirInfo, err = getDirInfo(c.Workdir)
	if err != nil {
		return fmt.Errorf("failed to get directory size: %w", err)
	}

	// Check if the current size exceeds the maximum size
	return c.cleanup()
}

func getDirInfo(path string) (int64, []fileInfo, error) {
	var totalSize int64
	var dirInfo []fileInfo

	entries, err := os.ReadDir(path)
	if err != nil {
		return 0, nil, err
	}

	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			return 0, nil, err
		}

		dirPath := filepath.Join(path, entry.Name())
		dirSize, modTime, err := getDirSizeAndModTime(dirPath)
		if err != nil {
			return 0, nil, err
		}
		// Collect file information
		dirInfo = append(dirInfo, fileInfo{
			Path:    dirPath,
			ModTime: modTime,
			IsDir:   info.IsDir(),
			Size:    dirSize,
		})

		// Add to total size if it's a file
		totalSize += dirSize
	}

	return totalSize, dirInfo, nil
}

func (c *Cleanup) cleanup() error {
	if c.currentSize <= c.TolerantSize {
		return nil
	}
	l := common.NewLoggerWithPrefixAndColor("cleanup: ")

	for _, file := range c.dirInfo {
		if c.currentSize <= c.TolerantSize {
			break
		}

		if c.currentSize > c.MaxSize {
			// Remove the file or directory
			if file.IsDir {
				err := os.RemoveAll(file.Path)
				if err != nil {
					return fmt.Errorf("failed to remove directory %s: %w", file.Path, err)
				}
			} else {
				err := os.Remove(file.Path)
				if err != nil {
					return fmt.Errorf("failed to remove file %s: %w", file.Path, err)
				}
			}
			// Update the current size
			c.currentSize -= file.Size
			l.Printf("Removed %s, current size: %d", file.Path, c.currentSize)
		} else {
			// c.CurrentSize > c.TolerantSize but <= c.MaxSize
			// Check if the file is older than MaxAge
			now := time.Now().Unix()
			if now-file.ModTime > c.MaxAge {
				// Remove the file or directory
				if file.IsDir {
					err := os.RemoveAll(file.Path)
					if err != nil {
						return fmt.Errorf("failed to remove directory %s: %w", file.Path, err)
					}
				} else {
					err := os.Remove(file.Path)
					if err != nil {
						return fmt.Errorf("failed to remove file %s: %w", file.Path, err)
					}
				}
				// Update the current size
				c.currentSize -= file.Size
				l.Printf("Removed %s, current size: %d", file.Path, c.currentSize)
			}
		}
	}
	return nil
}

func getDirSizeAndModTime(path string) (int64, int64, error) {
	var totalSize int64
	var latestModTime int64

	err := filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip the root directory itself
		if p == path {
			return nil
		}

		// Update total size
		if !info.IsDir() {
			totalSize += info.Size()
		}

		// Update latest modification time
		modTime := info.ModTime().Unix()
		if modTime > latestModTime {
			latestModTime = modTime
		}

		return nil
	})

	if err != nil {
		return 0, 0, err
	}

	return totalSize, latestModTime, nil
}
