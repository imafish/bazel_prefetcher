package main

import (
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"log"
	"os"
	"path"
	"strings"

	"internal/common"
	"internal/db"
	"internal/downloaders"
	"internal/git"
	"internal/httpserver"
	"internal/prefetcher"
)

func main() {
	if len(os.Args) < 3 {
		log.Fatalf("usage: %s <server_config.json> <prefetches.json>", os.Args[0])
	}
	serverConfigFile := os.Args[1]
	prefetchConfigFile := os.Args[2]

	// load config
	serverConfig, err := common.ReadServerConfig(serverConfigFile, prefetchConfigFile)
	if err != nil {
		log.Printf("Error reading server config: %s", err)
		return
	}
	serverConfig.Server.Workdir = strings.ReplaceAll(serverConfig.Server.Workdir, "$home", os.Getenv("HOME"))
	serverConfig.SrcDir = path.Join(serverConfig.Server.Workdir, "src")

	// create prefetch matchers
	prefetchers, err := prefetcher.CreatePrefetchersFromConfig(serverConfig.SrcDir, serverConfig.PrefetchConfig)
	if err != nil {
		log.Fatalf("Failed to generate prefetchers: %v", err)
	}

	logServerConfig(serverConfig)

	// load database
	database, err := db.CreateAndLoadDatabase(path.Join(serverConfig.Server.Workdir, "prefetch.db"))
	if err != nil {
		log.Printf("Error loading database: %s", err)
		return
	}
	defer database.Close()
	log.Printf("Database loaded successfully: %v", database)

	itemTable := db.NewItemTable(database)
	err = itemTable.Create()
	if err != nil {
		log.Printf("Error creating item table: %s", err)
		return
	}
	log.Printf("Item table created successfully: %v", itemTable)
	serverConfig.ItemTable = itemTable

	// LOGO
	log.Print(common.Imafish())

	// start http server
	go httpserver.StartServer(serverConfig)

	// start scheduler (periodically update repository, parse files and download)
	scheduler, err := NewScheduler(serverConfig.Server.Scheduler.Interval, serverConfig.Server.Scheduler.StartTime, serverConfig.Server.Scheduler.EndTime)
	if err != nil {
		log.Fatalf("Failed to create scheduler object, error: %s", err)
	}
	scheduler.Run(func() error {
		process(serverConfig, prefetchers)
		return nil
	})
}

func logServerConfig(config *common.ServerConfig) {
	common.LogSeparator("server config")
	log.Printf("Server Config: %+v\n", config)
	log.Printf("Got %d prefetcher configs.", len(config.PrefetchConfig.Items))
	for _, item := range config.PrefetchConfig.Items {
		log.Print(item)
	}
}

func process(config *common.ServerConfig, prefetchers []prefetcher.PrefetchMatchers) {
	common.LogSeparator("updating git")
	updateGit(config)

	common.LogSeparator("Analyzing prefetch items...")
	items, err := prefetcher.AnalyzePrefetchItems(prefetchers, path.Join(config.Server.Workdir, "data"))
	if err != nil {
		log.Println("Error analyzing prefetch items:", err)
		return
	}
	log.Printf("Analyzed prefetch items: %+v\n", len(items))

	common.LogSeparator("downloading and saving to data folder...")
	processPrefetchItems(config, items)

	common.LogSeparator("debug print item table")
	config.ItemTable.DebugPrintAll()
}

func processPrefetchItems(config *common.ServerConfig, items []*prefetcher.PrefetchItem) {
	downloadDir := path.Join(config.Server.Workdir, "downloads")
	if err := os.MkdirAll(downloadDir, os.ModePerm); err != nil {
		log.Printf("ERROR: Failed to create download directory: %v", err)
		return
	}

	for _, item := range items {
		common.LogSeparator(item.Url)
		processOneItem(config, item, downloadDir)
	}
}

func processOneItem(config *common.ServerConfig, item *prefetcher.PrefetchItem, downloadDir string) error {
	randStr := make([]byte, 8)
	rand.Read(randStr)
	filePath := path.Join(downloadDir, fmt.Sprintf("%x", randStr))
	if common.FileExists(filePath) {
		log.Printf("File already downloaded, deleting: %s", filePath)
		os.Remove(filePath)
	}
	defer os.Remove(filePath)

	log.Printf("Downloading file from URL: %s", item.Url)
	err := downloadFile(config, item.Url, filePath)
	if err != nil {
		log.Printf("Failed to download file from %s: %v", item.Url, err)
		item.Error = err
		return err
	}
	log.Printf("File downloaded successfully: %s", filePath)

	// calculate hashes, and update Item object
	err = updateItem(config, item, filePath)
	if err != nil {
		log.Printf("Failed to update item, error is: %v", err)
		err = fmt.Errorf("failed to update item, error is: %w", err)
		item.Error = err
		return err
	}

	cacheDir := path.Join(config.Server.Workdir, "data")
	err = saveAsBazelCache(item, cacheDir)
	if err != nil {
		log.Printf("Failed to move file to bazel cache: %v", err)
		item.Error = fmt.Errorf("failed to move file to bazel cache, error is: %w", err)
		return err
	}

	// save to database
	err = saveItemToDatabase(config, item)
	if err != nil {
		log.Printf("Failed to save item to database: %v", err)
		item.Error = fmt.Errorf("failed to save item to database, error is: %w", err)
		return err
	}

	return nil
}

func downloadFile(config *common.ServerConfig, url, filePath string) error {
	log.Printf("Downloading file from URL: %s to %s", url, filePath)
	downloaderFactory := downloaders.CreateDownloaderFactory(config)
	downloader, err := downloaderFactory.Create(config.Server.Downloader)
	if err != nil {
		log.Printf("Cannot create downloader %s, err = %s", config.Server.Downloader, err)
		return err
	}
	err = downloader.Download(url, filePath)
	if err != nil {
		log.Printf("Failed to download file from %s.", url)
		return err
	}

	return nil
}

func updateItem(_ *common.ServerConfig, item *prefetcher.PrefetchItem, filePath string) error {
	log.Printf("update item information.")

	item.Path = filePath

	// Hash of URL
	item.HashOfUrl = fmt.Sprintf("%x", sha256.Sum256([]byte(item.Url)))

	// compare Hash of File
	hash, err := common.HashOfFile(filePath)
	if err != nil {
		log.Printf("Failed to calculate file path: %v", err)
		err = fmt.Errorf("failed to calculate file path: %w", err)
		return err
	}
	if item.Hash == "" {
		log.Printf("file %s does not have a pre-defined hash. updating it to %s", item.Path, hash)
		item.Hash = hash
	} else if hash != item.Hash {
		err = fmt.Errorf("file `%s` hash does not match. Expected: %s, Actual: %s", filePath, item.Hash, hash)
		log.Print(err.Error())
		return err
	}

	// Size of file
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		err = fmt.Errorf("failed to get file size: %w", err)
		log.Print(err.Error())
		return err
	}
	item.Size = fileInfo.Size()

	return nil
}

func saveItemToDatabase(config *common.ServerConfig, item *prefetcher.PrefetchItem) error {
	log.Printf("Saving item to database: %+v", item)

	newItem := &db.Item{
		Hash:    item.Hash,
		Url:     item.Url,
		UrlHash: item.HashOfUrl,
		Path:    item.Path,
		Size:    item.Size,
	}

	// Insert the item into the database
	err := config.ItemTable.CreateOrUpdate(newItem)
	if err != nil {
		err = fmt.Errorf("failed to insert/update item into database: %v", err)
		log.Print(err)
		return err
	}

	log.Printf("Item saved to database successfully: %+v", newItem)
	return nil
}

func saveAsBazelCache(item *prefetcher.PrefetchItem, cacheDir string) error {
	log.Printf("Placing to bazel cache")
	outerDir := path.Join(cacheDir, item.Hash)
	innerFile := path.Join(outerDir, "file")
	hashFilePath := path.Join(outerDir, fmt.Sprintf("id-%s", item.HashOfUrl))
	os.MkdirAll(outerDir, 0755)

	hashFile, err := os.Create(hashFilePath)
	if err != nil {
		log.Printf("Failed to create file %s", hashFilePath)
		item.Error = fmt.Errorf("failed to create file %s, error is: %v", hashFilePath, err)
		return err
	}
	defer hashFile.Close()

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

func updateGit(config *common.ServerConfig) error {
	// Example function to update git repository
	gitDir := path.Join(config.Server.Workdir, "src")
	log.Println("Updating git repository at:", gitDir)

	git := git.GitRunner{
		RepoPath: gitDir,
	}

	err := git.UpdateRepository()
	if err != nil {
		log.Printf("Failed to update repository: %v", err)
		return err
	}

	return nil
}
