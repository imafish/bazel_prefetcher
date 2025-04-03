package downloaders

import (
	"fmt"
	"internal/common"
)

type Downloader interface {
	// Download downloads the file from the given URL and saves it to the specified path.
	// It returns an error if the download fails.
	Download(url string, path string) error
}

type DownloaderFactory interface {
	// Create creates a new Downloader instance based on the provided configuration.
	// It returns an error if the creation fails.
	Create(name string) (Downloader, error)
}

type DownloaderFactoryImpl struct {
	// Factories is a map of downloader names to their respective factory functions.
	Factories         map[string]func(*common.DownloaderConfig) Downloader
	DownloaderConfigs map[string]*common.DownloaderConfig
}

func CreateDownloaderFactory(config *common.ServerConfig) DownloaderFactory {
	downloaderConfigs := make(map[string]*common.DownloaderConfig)
	for _, conf := range config.Downloaders {
		downloaderConfigs[conf.Name] = &conf
	}

	return &DownloaderFactoryImpl{
		Factories: map[string]func(*common.DownloaderConfig) Downloader{
			"aria2": func(downloaderConfig *common.DownloaderConfig) Downloader {
				return &Aria2Downloader{DownloaderConfig: downloaderConfig}
			},
		},
		DownloaderConfigs: downloaderConfigs,
	}
}

func (f *DownloaderFactoryImpl) Create(name string) (Downloader, error) {
	if factory, exists := f.Factories[name]; exists {
		return factory(f.DownloaderConfigs[name]), nil
	}
	return nil, fmt.Errorf("downloader %s not found", name)
}
