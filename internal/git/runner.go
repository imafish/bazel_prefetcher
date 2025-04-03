package git

import (
	"fmt"
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
	return resetRepository(gr.RepoPath)
}

// UpdateRepository updates the repository from 'origin/master' and all its submodules.
func (gr *GitRunner) UpdateRepository() error {
	return updateRepository(gr.RepoPath)
}

func resetRepository(repoPath string) error {
	cmd := exec.Command("git", "-C", repoPath, "reset", "--hard", "origin/master")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to reset repository: %s, %v", string(output), err)
	}
	return nil
}

// updateRepository updates the repository from 'origin/master' and all its submodules.
func updateRepository(repoPath string) error {
	// git fetch origin
	cmdFetch := exec.Command("git", "-C", repoPath, "fetch", "origin")
	if output, err := cmdFetch.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to fetch from origin: %s, %v", string(output), err)
	}

	// git checkout master
	cmdPull := exec.Command("git", "-C", repoPath, "checkout", "master")
	if output, err := cmdPull.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to checkout master: %s, %v", string(output), err)
	}

	// git reset --hard origin/master
	cmd := exec.Command("git", "-C", repoPath, "reset", "--hard", "origin/master")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to reset repository: %s, %v", string(output), err)
	}

	// git submodule update --recursive
	cmdSubmodule := exec.Command("git", "-C", repoPath, "submodule", "update", "--init", "--recursive")
	if output, err := cmdSubmodule.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to update submodules: %s, %v", string(output), err)
	}

	return nil
}
