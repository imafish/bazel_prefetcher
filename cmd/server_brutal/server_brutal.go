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

	end := time.Now()
	common.LogSeparator("summary")
	log.Printf("Time taken: %s", end.Sub(start))
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
		if err := runOneCommand("bazel", bc, srcDir); err != nil {
			l.Printf("Failed to run bazel command: %v", err)
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
