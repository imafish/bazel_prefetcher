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
	PruneRepository() error
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

// PruneRepository runs `git gc --prune` on repository and submodules
func (gr *GitRunner) PruneRepository() error {
	cmd := exec.Command("git", "-C", gr.RepoPath, "gc", "--prune")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to prune repository: %s, %v", string(output), err)
	}
	cmd = exec.Command("git", "-C", gr.RepoPath, "submodule", "foreach", "git", "gc", "--prune")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to prune submodules: %s, %v", string(output), err)
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
