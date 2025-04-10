package prefetcher

import (
	"crypto/sha256"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"

	"internal/common"
)

func AnalyzePrefetchItems(config *common.ServerConfig, prefetchers []PrefetchMatchers) ([]*PrefetchItem, error) {
	items := make([]*PrefetchItem, 0, len(config.PrefetchConfig.Items))
	for _, i := range prefetchers {
		item, err := analyzePrefetchItem(config, &i)
		if err == os.ErrExist {
			log.Printf("Item %s common.FileExists in bazel cache.", i.Name)
		} else if err != nil {
			log.Printf("Failed to analyze item %s", i.Name)
		} else {
			log.Printf("Got item: %+v", item)
			items = append(items, item)
		}
	}

	return items, nil
}

func analyzePrefetchItem(config *common.ServerConfig, info *PrefetchMatchers) (*PrefetchItem, error) {
	item, err := getDownloadUrlAndHash(info)
	if err != nil {
		log.Printf("Failed to get download URL and hash for item %s: %v", info.Name, err)
		return nil, err
	}

	cacheDir := path.Join(config.Server.Workdir, "data")
	err = checkIfExistsInBazelCache(item, cacheDir)
	if err != nil {
		if err == os.ErrExist {
			log.Printf("The item %s already common.FileExists in bazel cache.", info.Name)
		} else {
			log.Printf("Failed to check if item %s common.FileExists in bazel cache: %v", info.Name, err)
		}
		return nil, err
	}
	return item, nil
}

func checkIfExistsInBazelCache(item *PrefetchItem, cacheDir string) error {
	log.Printf("Checking if item %s common.FileExists in bazel cache...", item.Url)

	hashOfUrl := fmt.Sprintf("%x", sha256.Sum256([]byte(item.Url)))
	if item.Hash == "" {
		// just try to find the id file exist
		l := common.NewLoggerWithPrefixAndColor("DBG: ")
		hashFilename := fmt.Sprintf("id-%s", hashOfUrl)
		l.Print(hashFilename)
		found, parentDir, err := findFileAndReturnParent(cacheDir, hashFilename)
		l.Print(found, parentDir, err)
		if err != nil {
			return err
		}
		if !found {
			return nil
		} else if common.FileExists(path.Join(parentDir, "file")) {
			return os.ErrExist
		} else {
			return nil
		}
	}

	outerDir := path.Join(cacheDir, item.Hash)
	innerFile := path.Join(outerDir, "file")
	hashFile := path.Join(outerDir, fmt.Sprintf("id-%s", hashOfUrl))

	if common.FileExists(innerFile) && common.FileExists(hashFile) {
		log.Printf("Found cache item at: `%s`", outerDir)
		return os.ErrExist
	}

	log.Printf("Item %s does not exist in bazel cache.", item.Url)
	return nil
}

func findFileAndReturnParent(root string, filename string) (bool, string, error) {
	var found bool
	var parentDir string

	err := filepath.WalkDir(root, func(currentPath string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() && d.Name() == filename {
			found = true
			parentDir = path.Dir(currentPath)
			return filepath.SkipDir
		}
		return nil
	})

	return found, parentDir, err
}

func getDownloadUrlAndHash(item *PrefetchMatchers) (*PrefetchItem, error) {
	matched, url, err := item.UrlMatcher.Match()
	if err != nil {
		log.Printf("error when trying to find url package %s, err: %v", item.Name, err)
		return nil, err
	}
	if !matched {
		log.Printf("url of package `%s` not found in src.", item.Name)
		return nil, os.ErrNotExist
	}

	matched, hash, err := item.HashMatcher.Match()
	if err != nil {
		log.Printf("error when trying to find hash for package %s, err: %v", item.Name, err)
		return nil, err
	}
	if !matched {
		log.Printf("hash of package `%s` not found in src.", item.Name)
		return nil, os.ErrNotExist
	}

	return &PrefetchItem{
		Url:  url,
		Hash: hash,
	}, nil
}
