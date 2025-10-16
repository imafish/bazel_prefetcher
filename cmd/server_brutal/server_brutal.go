package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"
	"sync"
	"time"

	"internal/cleanup"
	"internal/common"
	"internal/git"
	"internal/httpserver"
)

type server struct {
	ServerConfig  *common.ServerConfig
	BazelCommands *common.BazelCommandsConfig

	Mtx sync.Mutex
}

func main() {
	if len(os.Args) < 3 {
		log.Fatalf("usage: %s <server_config.json> <bazel_commands_file.json>", os.Args[0])
	}
	serverConfigFile := os.Args[1]
	bazelCommandsConfigFile := os.Args[2]

	server := &server{
		Mtx: sync.Mutex{},
	}
	var err error

	// load config
	server.ServerConfig, err = common.ReadServerConfigJson(serverConfigFile)
	if err != nil {
		log.Printf("Error reading server config: %s", err)
		return
	}
	server.ServerConfig.Server.Workdir = strings.ReplaceAll(server.ServerConfig.Server.Workdir, "$home", os.Getenv("HOME"))
	server.ServerConfig.SrcDir = path.Join(server.ServerConfig.Server.Workdir, "src")

	// load bazel commands
	server.BazelCommands, err = common.ReadBazelCommandsConfigJson(bazelCommandsConfigFile)
	if err != nil {
		log.Printf("Error reading bazel commands config: %s", err)
		return
	}

	common.LogSeparator("server config")
	common.PrintStruct(server, func(s string) {
		log.Printf("%s", s)
	})

	// LOGO
	log.Print(common.Imafish())

	serverConfig := server.ServerConfig

	// start http server
	httpServerBuilder := httpserver.NewHttpServerBuilder(serverConfig)
	httpServerBuilder.ServeFiles()
	httpServerBuilder.ServeApiV1Files()
	httpServer := httpServerBuilder.Build()
	log.Printf("Starting HTTP server on port %d", serverConfig.Server.Port)
	go httpServer.ListenAndServe()

	// start scheduler (periodically update repository, parse files and download)
	scheduler, err := NewScheduler(serverConfig.Server.Scheduler.Interval, serverConfig.Server.Scheduler.StartTime, serverConfig.Server.Scheduler.EndTime)
	if err != nil {
		log.Fatalf("Failed to create scheduler object, error: %s", err)
	}
	scheduler.Run(func() error {
		process(server)
		return nil
	})
}

func process(server *server) {
	config := server.ServerConfig
	start := time.Now()
	common.LogSeparator("updating and cleaning source code...")
	cleanGit(config)
	updateGit(config)

	common.LogSeparator("running bazel build --nobuild...")
	runBazelBuild(server)

	common.LogSeparator("cleaning up...")
	if config.Server.Cleanup.Enabled {
		cleanup := cleanup.Cleanup{
			Workdir:      path.Join(config.Server.Workdir, "data", "content_addressable", "sha256"),
			MaxSize:      config.Server.Cleanup.MaxSize,
			TolerantSize: config.Server.Cleanup.TolerantSize,
			MaxAge:       int64(config.Server.Cleanup.MaxAge * 24 * 60 * 60), // Convert days to seconds
		}
		if err := cleanup.Run(); err != nil {
			log.Printf("Error during cleanup: %s", err)
		}
	} else {
		log.Printf("Cleanup is disabled in the configuration.")
	}

	end := time.Now()
	common.LogSeparator("summary")
	log.Printf("Time taken: %s", end.Sub(start))
}

func cleanGit(config *common.ServerConfig) error {
	l := common.NewLoggerWithPrefixAndColor("git: ")
	// Example function to update git repository
	gitDir := path.Join(config.Server.Workdir, "src")
	log.Println("Pruning git repository at:", gitDir)

	git := git.GitRunner{
		RepoPath: gitDir,
	}

	retryCnt := 2
	var err error
	for i := 0; i < retryCnt; i++ {
		log.Printf("Attempt #%d to prune repository...", i+1)

		err = git.PruneRepository()
		if err != nil {
			l.Printf("Failed to prune repository: %v", err)
		} else {
			l.Printf("Repository pruned successfully after %d attempts", i+1)
			break
		}
	}

	return nil
}

func updateGit(config *common.ServerConfig) error {
	l := common.NewLoggerWithPrefixAndColor("git: ")
	// Example function to update git repository
	gitDir := path.Join(config.Server.Workdir, "src")
	log.Println("Updating git repository at:", gitDir)

	git := git.GitRunner{
		RepoPath: gitDir,
	}

	retryCnt := 3
	var err error
	for i := 0; i < retryCnt; i++ {
		log.Printf("Attempt #%d to update repository...", i+1)

		err = git.UpdateRepository()
		if err != nil {
			l.Printf("Failed to update repository: %v", err)
		} else {
			l.Printf("Repository updated successfully after %d attempts", i+1)
			break
		}
	}

	return nil
}

func runBazelBuild(server *server) error {
	config := server.ServerConfig
	l := common.NewLoggerWithPrefixAndColor("bazel: ")

	srcDir := path.Join(config.Server.Workdir, "src")
	dataDir := path.Join(config.Server.Workdir, "data")

	repositoryCacheParam := fmt.Sprintf("--repository_cache=%s", dataDir)
	l.Printf("Using repository cache path: %s", dataDir)

	// get a copy of bazel commands
	server.Mtx.Lock()
	bazelCommands := make([][]string, len(server.BazelCommands.Commands))
	for i, bc := range server.BazelCommands.Commands {
		bazelCommands[i] = make([]string, len(bc))
		copy(bazelCommands[i], bc)
	}
	server.Mtx.Unlock()

	var err error

	for _, bc := range bazelCommands {
		bc = append(bc, repositoryCacheParam)
		retryCnt := 5
		for i := range retryCnt {
			l.Printf("Attempt #%d to run bazel command...", i+1)
			if err = runOneCommand("bazel", bc, srcDir); err != nil {
				l.Printf("Failed to run bazel command: %v", err)
				l.Printf("Retrying in 5 seconds...")
				time.Sleep(5 * time.Second)
			} else {
				l.Printf("Bazel command executed successfully after %d attempts", i+1)
				break
			}
		}
		if err != nil {
			l.Printf("Failed to run bazel command after %d attempts: %v", retryCnt, err)
			return err
		}
	}

	return nil
}

func runOneCommand(cmd string, params []string, srcDir string) error {
	l := common.NewLoggerWithPrefixAndColor("cmd: ")
	l.Printf("Running command: %s %s", cmd, strings.Join(params, " "))
	command := exec.Command(cmd, params...)
	command.Stderr = os.Stderr
	command.Dir = srcDir
	stdout, err := command.StdoutPipe()
	if err != nil {
		l.Printf("failed to create stdout pipe: %v", err)
		return err
	}
	if err := command.Start(); err != nil {
		l.Printf("failed to start bazel cmd, error: %v", err)
		return err
	}

	go io.Copy(os.Stdout, stdout)

	if err := command.Wait(); err != nil {
		l.Printf("failed to wait for bazel cmd, error: %v", err)
		return err
	}
	return nil
}
