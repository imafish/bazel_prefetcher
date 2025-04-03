package main

import (
	"log"
	"os"
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

	updateGit(serverConfig)
}

func updateGit(config *ServerConfig) error {
	// Example function to update git repository
	log.Println("Updating git repository at:", config.Server.Workdir)

	git := git.GitRunner{
		RepoPath: config.Server.Workdir,
	}

	err := git.UpdateRepository()
	if err != nil {
		log.Printf("Failed to update repository: %v", err)
		return err
	}

	return nil
}
