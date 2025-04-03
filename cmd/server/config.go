package main

import (
	"encoding/json"
	"os"
)

// Structs for server.json
type ServerConfig struct {
	Server struct {
		Port       int    `json:"port"`
		Host       string `json:"host"`
		Timeout    int    `json:"timeout"`
		Downloader string `json:"downloader"`
		Workdir    string `json:"workdir"`
	} `json:"server"`
	Downloader struct {
		Name        string   `json:"name"`
		Cmd         string   `json:"cmd"`
		Pattern     string   `json:"pattern"`
		DefaultArgs []string `json:"default_args"`
		Args        []struct {
			Matcher struct {
				Type    string `json:"type"`
				Pattern string `json:"pattern"`
			} `json:"matcher"`
			Args []string `json:"args"`
		} `json:"args"`
	} `json:"downloader"`
}

// Structs for prefetch.json
type PrefetchConfig struct {
	Packages []struct {
		Name   string `json:"name"`
		Finder struct {
			Type        string `json:"type"`
			Path        string `json:"path"`
			Pattern     string `json:"pattern"`
			Prefix      string `json:"prefix"`
			HashPattern string `json:"hash_pattern"`
		} `json:"finder"`
	} `json:"packages"`
}

// Function to read server.json
func readServerConfig(filePath string) (*ServerConfig, error) {
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
func readPrefetchConfig(filePath string) (*PrefetchConfig, error) {
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
