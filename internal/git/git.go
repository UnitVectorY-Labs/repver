package git

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// IsGitRoot checks if the current working directory is the root of a Git repository.
func IsGitRoot() (bool, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return false, fmt.Errorf("error getting working directory: %w", err)
	}

	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	output, err := cmd.Output()
	if err != nil {
		return false, err
	}

	gitRoot := strings.TrimSpace(string(output))
	return cwd == gitRoot, nil
}

// BranchExists checks if a branch with the given name exists in the Git repository.
func BranchExists(branchName string) (bool, error) {
	cmd := exec.Command("git", "show-ref", "--verify", "--quiet", "refs/heads/"+branchName)
	err := cmd.Run()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok && exitError.ExitCode() == 1 {
			return false, nil
		}
		return false, fmt.Errorf("error checking branch existence: %w", err)
	}
	return true, nil
}

// GetCurrentBranch retrieves the name of the current branch in the Git repository.
func GetCurrentBranch() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("error getting current branch: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

// SwitchToBranch switches to the specified branch in the Git repository.
// Returns the command output for logging purposes.
func SwitchToBranch(branchName string) (string, error) {
	cmd := exec.Command("git", "checkout", branchName)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("error switching to branch %s: %w", branchName, err)
	}
	return string(output), nil
}

// CheckGitClean checks if the Git repository is clean (i.e., no uncommitted changes).
func CheckGitClean() error {
	cmd := exec.Command("git", "status", "--porcelain")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("error checking git status: %w", err)
	}
	if len(output) > 0 {
		return fmt.Errorf("git repository is not clean")
	}
	return nil
}

// CreateAndSwitchBranch creates a new branch and switches to it.
// Returns the command output for logging purposes.
func CreateAndSwitchBranch(branchName string) (string, error) {
	cmd := exec.Command("git", "checkout", "-b", branchName)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("error creating and switching to branch %s: %w", branchName, err)
	}
	return string(output), nil
}

// AddAndCommitFiles adds files to the staging area and commits them with a message.
// Returns the commit output for logging purposes.
func AddAndCommitFiles(fileNames []string, commitMessage string) (string, error) {
	var output strings.Builder

	for _, fileName := range fileNames {
		cmd := exec.Command("git", "add", fileName)
		addOutput, err := cmd.Output()
		if err != nil {
			return "", fmt.Errorf("error adding file %s: %w", fileName, err)
		}
		output.WriteString(string(addOutput))
	}

	cmd := exec.Command("git", "commit", "-m", commitMessage)
	commitOutput, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("error committing changes: %w", err)
	}
	output.WriteString(string(commitOutput))

	return output.String(), nil
}

// PushChanges pushes the changes to the specified remote and branch.
// Returns the command output for logging purposes.
func PushChanges(remote string, branch string) (string, error) {
	cmd := exec.Command("git", "push", remote, branch)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("error pushing changes to %s/%s: %w", remote, branch, err)
	}
	return string(output), nil
}

// DeleteLocalBranch deletes a local branch with the specified name.
// Returns the command output for logging purposes.
func DeleteLocalBranch(branchName string) (string, error) {
	cmd := exec.Command("git", "branch", "-D", branchName)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("error deleting local branch %s: %w", branchName, err)
	}
	return string(output), nil
}
