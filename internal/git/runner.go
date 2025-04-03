package git

import (
	"fmt"
	"log"
	"os/exec"
)

type GitRunner struct {
	RepoPath string
}
type GitRunnerInterface interface {
	ResetRepository() error
	UpdateRepository() error
}

// ResetRepository resets the git repository to a clean state.
func (gr *GitRunner) ResetRepository() error {
	cmd := exec.Command("git", "-C", gr.RepoPath, "reset", "--hard", "origin/master")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to reset repository: %s, %v", string(output), err)
	}
	return nil
}

// UpdateRepository updates the repository from 'origin/master' and all its submodules.
func (gr *GitRunner) UpdateRepository() error {
	repoPath := gr.RepoPath

	// git fetch -f --prune
	log.Printf("executing command: git -C %s fetch -f --prune", repoPath)
	cmdFetch := exec.Command("git", "-C", repoPath, "fetch", "-f", "--prune")
	output, err := cmdFetch.CombinedOutput()
	if err != nil {
		return fmt.Errorf("output: %s, failed to pull, %v", string(output), err)
	}
	log.Print(string(output))

	// git reset --hard origin/master
	log.Printf("executing command; git -C %s reset --hard origin/master", repoPath)
	cmd := exec.Command("git", "-C", repoPath, "reset", "--hard", "origin/master")
	output, err = cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("output: %s, failed to reset repository: %v", string(output), err)
	}
	log.Print(string(output))

	// git submodule update --recursive -f
	log.Printf("executing command; git -C %s submodule update --recursive -f", repoPath)
	cmdSubmodule := exec.Command("git", "-C", repoPath, "submodule", "update", "--recursive", "-f")
	output, err = cmdSubmodule.CombinedOutput()
	if err != nil {
		return fmt.Errorf("output: %s, failed to update submodules: %v", string(output), err)
	}
	log.Print(string(output))

	return nil
}
