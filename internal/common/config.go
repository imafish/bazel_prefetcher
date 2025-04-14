package common

import (
	"encoding/json"
	"os"

	"internal/db"
)

// Structs for server.json
type ServerConfig struct {
	Server struct {
		Port       int    `json:"port"`
		Host       string `json:"host"`
		Timeout    int    `json:"timeout"`
		Downloader string `json:"downloader"`
		Workdir    string `json:"workdir"`
		Scheduler  struct {
			Interval  int    `json:"interval"`
			StartTime string `json:"start_time"`
			EndTime   string `json:"end_time"`
		} `json:"scheduler"`
	} `json:"server"`
	Downloaders    []DownloaderConfig `json:"downloaders"`
	PrefetchConfig *PrefetchConfig
	ItemTable      *db.ItemTable
	SrcDir         string
}

type DownloaderConfig struct {
	Name        string   `json:"name"`
	Cmd         string   `json:"cmd"`
	DefaultArgs []string `json:"default_args"`
	Args        []struct {
		Matcher struct {
			Type    string `json:"type"`
			Pattern string `json:"pattern"`
		} `json:"matcher"`
		Args []string `json:"args"`
	} `json:"args"`
}

// Structs for prefetch.json

type Package struct {
	Name              string        `json:"name"`
	HashMatcherConfig MatcherConfig `json:"hash_matcher"`
	UrlMatcherConfig  MatcherConfig `json:"url_matcher"`
}

type MatcherConfig struct {
	Type        string `json:"type"`
	File        string `json:"file"`
	Format      string `json:"format"`
	Regex       string `json:"regex"`
	AnchorRegex string `json:"anchor_regex"`
	MaxLines    int    `json:"max_lines"`
}

type PrefetchConfig struct {
	Items []*Package `json:"items"`
}

func ReadServerConfigAll(serverConfigPath string, prefetchConfigPath string) (*ServerConfig, error) {
	serverConfig, err := ReadServerConfigJson(serverConfigPath)
	if err != nil {
		return nil, err
	}

	prefetchConfig, err := ReadPrefetchConfigJson(prefetchConfigPath)
	if err != nil {
		return nil, err
	}

	serverConfig.PrefetchConfig = prefetchConfig

	return serverConfig, nil
}

// Function to read server.json
func ReadServerConfigJson(filePath string) (*ServerConfig, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var config ServerConfig
	if err := json.NewDecoder(file).Decode(&config); err != nil {
		return nil, err
	}
	return &config, nil
}

// Function to read prefetch.json
func ReadPrefetchConfigJson(filePath string) (*PrefetchConfig, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var config PrefetchConfig
	if err := json.NewDecoder(file).Decode(&config); err != nil {
		return nil, err
	}
	return &config, nil
}
