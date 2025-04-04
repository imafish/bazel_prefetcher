package main

import (
	"log"
	"os"
	"path"
	"strings"

	"internal/git"
)

func main() {
	// Example usage
	serverConfig, err := readServerConfig("../../config/server.json")
	if err != nil {
		log.Printf("Error reading server config: %s", err)
		return
	}
	log.Printf("Server Config: %+v\n", serverConfig)

	prefetchConfig, err := readPrefetchConfig("../../config/prefetch.json")
	if err != nil {
		log.Println("Error reading prefetch config:", err)
		return
	}
	log.Printf("Prefetch Config: %+v\n", prefetchConfig)

	// replace $home with the actual home directory in serverConfig
	serverConfig.Server.Workdir = strings.ReplaceAll(serverConfig.Server.Workdir, "$home", os.Getenv("HOME"))

	process(serverConfig, prefetchConfig)
	updateGit(serverConfig)
}

func process(config *ServerConfig, prefetchConfig *PrefetchConfig) {
	// Example function to process the server and prefetch configurations
	log.Println("Processing server and prefetch configurations...")
	log.Printf("Server Workdir: %s\n", config.Server.Workdir)
	log.Printf("Downloader Name: %s\n", config.Downloader.Name)
	log.Printf("Prefetch Packages: %+v\n", prefetchConfig.Packages)

	log.Printf("updating git repository.")
	updateGit(config)
	log.Println("updating git completed.")
}

func updateGit(config *ServerConfig) error {
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
