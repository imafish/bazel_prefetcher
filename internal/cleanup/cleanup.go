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
	l := common.NewLoggerWithPrefixAndColor("cleanup: ")

	if _, err := os.Stat(c.Workdir); os.IsNotExist(err) {
		return fmt.Errorf("workdir does not exist: %s", c.Workdir)
	}

	// Get the current size of the workdir
	l.Printf("Calculating current size of workdir: %s", c.Workdir)
	var err error
	c.currentSize, c.dirInfo, err = getDirInfo(c.Workdir)
	if err != nil {
		return fmt.Errorf("failed to get directory size: %w", err)
	}
	l.Printf("Current size of workdir: %s, items: %d", common.PrettyPrintSize(c.currentSize), len(c.dirInfo))

	for _, file := range c.dirInfo {
		l.Printf("File: %s, Size: %s, ModTime: %s", file.Path, common.PrettyPrintSize(file.Size), time.Unix(file.ModTime, 0).Format(time.RFC3339))
	}

	// do the cleanup
	return c.doCleanUp()
}

// TODO: it looks like there's a bug in this code. For directory contains files deepth > 1, it gets the size wrong.
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

func (c *Cleanup) doCleanUp() error {
	l := common.NewLoggerWithPrefixAndColor("cleanup: ")
	if c.currentSize <= c.TolerantSize {
		l.Printf("Current size %s is within the tolerant size %s, no cleanup needed", common.PrettyPrintSize(c.currentSize), common.PrettyPrintSize(c.TolerantSize))
		return nil
	}

	deleted := 0
	sizeFreed := int64(0)

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
			deleted++
			sizeFreed += file.Size
			l.Printf("Removed %s, current size: %s", file.Path, common.PrettyPrintSize(c.currentSize))
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
				deleted++
				sizeFreed += file.Size
				l.Printf("Removed %s, current size: %s", file.Path, common.PrettyPrintSize(c.currentSize))
			}
		}
	}
	l.Printf("Deleted %d items, freed %s", deleted, common.PrettyPrintSize(sizeFreed))
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
