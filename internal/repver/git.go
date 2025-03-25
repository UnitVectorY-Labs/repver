package repver

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// IsGitRoot checks if the current working directory is the root of a Git repository.
func IsGitRoot() (bool, error) {
	// Get the current working directory.
	cwd, err := os.Getwd()
	if err != nil {
		return false, err
	}

	// Execute the Git command to get the root directory of the repository.
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	var out bytes.Buffer
	cmd.Stdout = &out
	// Suppress any error output by directing stderr to nil.
	cmd.Stderr = nil
	if err := cmd.Run(); err != nil {
		// If the command fails, it's likely that we are not in a Git repository.
		return false, nil
	}

	// Trim any trailing newline or spaces from the output.
	gitRoot := strings.TrimSpace(out.String())

	// Compare the Git root with the current working directory.
	return cwd == gitRoot, nil
}

// BranchExists checks if a branch with the given name exists in the Git repository.
func BranchExists(branchName string) (bool, error) {
	cmd := exec.Command("git", "show-ref", "--verify", "--quiet", "refs/heads/"+branchName)
	err := cmd.Run()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok && exitError.ExitCode() == 1 {
			// Branch does not exist
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

	branchName := string(output)
	return branchName[:len(branchName)-1], nil // Remove the newline character
}

// SwitchToBranch switches to the specified branch in the Git repository.
func SwitchToBranch(branchName string) error {
	cmd := exec.Command("git", "checkout", branchName)
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("error switching to branch %s: %w", branchName, err)
	}

	fmt.Println(string(output))
	return nil
}

// CheckGitClean checks if the Git repository is clean (i.e., no uncommitted changes).
func CheckGitClean() error {
	// Check if the git repository is clean
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
func CreateAndSwitchBranch(branchName string) error {
	cmd := exec.Command("git", "checkout", "-b", branchName)
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("error creating and switching to branch %s: %w", branchName, err)
	}

	fmt.Println(string(output))
	return nil
}

// AddAndCommitFiles adds files to the staging area and commits them with a message.
func AddAndCommitFiles(fileNames []string, commitMessage string) error {
	// Add the files to the staging area
	for _, fileName := range fileNames {
		cmd := exec.Command("git", "add", fileName)
		output, err := cmd.Output()
		if err != nil {
			return fmt.Errorf("error adding file %s: %w", fileName, err)
		}
		fmt.Println(string(output))
	}

	// Commit the changes
	cmd := exec.Command("git", "commit", "-m", commitMessage)
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("error committing changes: %w", err)
	}

	fmt.Println(string(output))
	return nil
}

// PushChanges pushes the changes to the specified remote and branch.
func PushChanges(remote string, branch string) error {
	cmd := exec.Command("git", "push", remote, branch)
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("error pushing changes to %s/%s: %w", remote, branch, err)
	}

	fmt.Println(string(output))
	return nil
}

// DeleteLocalBranch deletes a local branch with the specified name.
func DeleteLocalBranch(branchName string) error {
	cmd := exec.Command("git", "branch", "-D", branchName)
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("error deleting local branch %s: %w", branchName, err)
	}

	fmt.Println(string(output))
	return nil
}
