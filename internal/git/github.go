package git

import (
	"fmt"
	"os/exec"
)

// CreateGitHubPullRequest creates a pull request on GitHub using the GitHub CLI.
// Returns the output of the command for logging purposes.
func CreateGitHubPullRequest() (string, error) {
	cmd := exec.Command("gh", "pr", "create", "--fill")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("error creating GitHub pull request: %w", err)
	}

	return string(output), nil
}
