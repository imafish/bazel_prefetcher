package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"

	"internal/cleanup"
	"internal/common"
	"internal/git"
	"internal/httpserver"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("usage: %s <server_config.json>", os.Args[0])
	}
	serverConfigFile := os.Args[1]

	// load config
	serverConfig, err := common.ReadServerConfigJson(serverConfigFile)
	if err != nil {
		log.Printf("Error reading server config: %s", err)
		return
	}
	serverConfig.Server.Workdir = strings.ReplaceAll(serverConfig.Server.Workdir, "$home", os.Getenv("HOME"))
	serverConfig.SrcDir = path.Join(serverConfig.Server.Workdir, "src")

	common.LogSeparator("server config")
	log.Printf("Server Config: %+v\n", serverConfig)

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
		process(serverConfig)
		return nil
	})
}

func process(config *common.ServerConfig) {
	start := time.Now()
	common.LogSeparator("updating source code...")
	updateGit(config)

	common.LogSeparator("running bazel build --nobuild...")
	runBazelBuild(config)

	common.LogSeparator("cleaning up...")
	if config.Cleanup.Enabled {
		cleanup := cleanup.Cleanup{
			Workdir:      config.Server.Workdir,
			MaxSize:      config.Cleanup.MaxSize,
			TolerantSize: config.Cleanup.TolerantSize,
			MaxAge:       int64(config.Cleanup.MaxAge * 24 * 60 * 60), // Convert days to seconds
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

func runBazelBuild(config *common.ServerConfig) error {
	l := common.NewLoggerWithPrefixAndColor("bazel: ")

	srcDir := path.Join(config.Server.Workdir, "src")
	dataDir := path.Join(config.Server.Workdir, "data")

	repositoryCacheParam := fmt.Sprintf("--repository_cache=%s", dataDir)
	l.Printf("Using repository cache path: %s", dataDir)

	l.Print("cleanning up...")
	err := runOneCommand("bazel", []string{"clean", repositoryCacheParam}, srcDir)
	if err != nil {
		l.Printf("Failed to run bazel clean command: %v", err)
		return err
	}

	bazelCommands := [][]string{
		{"--config=spp_host_gcc", "//platform/aas/intc/lifecycle_state_machine/code:lifecycle_state_machine"},
		{"--config=spp_host_gcc", "//platform/aas/intc/phmheartbeatproxy/code:PhmHeartBeatProxy"},
		{"--config=ipnext_arm64_qnx", "--python=3.8", "//ecu/xpad/xpad-shared/packaging/ipnext/isoc/image:IPNext_HLOS"},
	}

	for _, bc := range bazelCommands {
		bc = append([]string{"build", repositoryCacheParam, "--nobuild"}, bc...)
		retryCnt := 5
		for i := 0; i < retryCnt; i++ {
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
